// @title           Foundry API
// @version         0.1.0
// @description     Spec-driven AI development platform — orchestrates teams of Claude Code agents.

// @host            localhost:8080
// @BasePath        /api
// @schemes         http

// @accept          json
// @produce         json

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/api"
	"github.com/cenron/foundry/internal/broker"
	"github.com/cenron/foundry/internal/cache"
	"github.com/cenron/foundry/internal/config"
	"github.com/cenron/foundry/internal/database"
	"github.com/cenron/foundry/internal/event"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/po"
	"github.com/cenron/foundry/internal/project"
	"github.com/cenron/foundry/internal/runtime"

	_ "github.com/cenron/foundry/api/swagger"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	log.Println("foundry starting...")

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() { _ = db.Close() }()
	log.Println("connected to postgres")

	cacheClient, err := cache.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("cache: %v", err)
	}
	defer func() { _ = cacheClient.Close() }()
	log.Println("connected to redis")

	if err := database.MigrateUp(db, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations applied")

	lib, err := agent.NewLibrary(cfg.AgentLibraryPath)
	if err != nil {
		log.Printf("agent library: %v (continuing without library)", err)
	}

	// Extract embedded PO workspace (CLAUDE.md + playbooks) to foundry home.
	if err := po.DeployPOWorkspace(cfg.FoundryHome); err != nil {
		log.Printf("deploying PO workspace: %v (PO sessions may lack playbooks)", err)
	} else {
		log.Println("PO workspace deployed")
	}

	poManager := po.NewSessionManager(cfg.FoundryHome, cfg.AnthropicAPIKey, cfg.ClaudeVersion)

	deps := api.ServerDeps{
		Cache:        cacheClient,
		Projects:     project.NewStore(db),
		Specs:        project.NewSpecStore(db),
		Tasks:        orchestrator.NewTaskStore(db),
		Agents:       agent.NewStore(db),
		Library:      lib,
		RiskProfiles: project.NewRiskProfileStore(db),
		PO:           poManager,
		FoundryHome:  cfg.FoundryHome,
	}

	var brokerClient *broker.Client

	if cfg.IsLocalMode() {
		log.Println("local mode: skipping RabbitMQ, using local runtime")
		deps.Runtime = runtime.NewLocalRuntime(cfg.MaxConcurrentAgents)
		poManager = po.NewLocalSessionManager(cfg.FoundryHome, cfg.ClaudeVersion)
		deps.PO = poManager
	} else {
		brokerClient, err = broker.Connect(ctx, cfg.RabbitMQURL)
		if err != nil {
			log.Fatalf("broker: %v", err)
		}
		defer func() { _ = brokerClient.Close() }()
		log.Println("connected to rabbitmq")
		deps.Broker = brokerClient
	}

	srv := api.NewServer(deps)

	// Event router — only in Docker mode (needs broker for subscriptions).
	if brokerClient != nil {
		eventStore := event.NewStore(db)
		eventRouter := event.NewRouter(eventStore, srv.Hub(), cacheClient, brokerClient)
		if err := eventRouter.Start(); err != nil {
			log.Printf("event router: %v", err)
		}
	}

	log.Printf("runtime mode: %s", cfg.RuntimeMode)

	httpServer := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      srv.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("listening on :%s", cfg.APIPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")

	// Stop PO sessions before HTTP server — prevents orphaned child processes.
	poManager.StopAll()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}

	log.Println("foundry stopped")
}

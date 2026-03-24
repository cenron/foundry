package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenron/foundry/internal/api"
	"github.com/cenron/foundry/internal/broker"
	"github.com/cenron/foundry/internal/cache"
	"github.com/cenron/foundry/internal/config"
	"github.com/cenron/foundry/internal/database"
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

	brokerClient, err := broker.Connect(ctx, cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("broker: %v", err)
	}
	defer func() { _ = brokerClient.Close() }()
	log.Println("connected to rabbitmq")

	if err := database.MigrateUp(db, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations applied")

	srv := api.NewServer(db, cacheClient, brokerClient)

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}

	log.Println("foundry stopped")
}

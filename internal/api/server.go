package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/broker"
	"github.com/cenron/foundry/internal/cache"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/po"
	"github.com/cenron/foundry/internal/project"
)

// ServerDeps bundles all dependencies for the API server.
type ServerDeps struct {
	Cache        *cache.Client
	Broker       *broker.Client
	Projects     *project.Store
	Specs        *project.SpecStore
	Tasks        *orchestrator.TaskStore
	Agents       *agent.Store
	Library      *agent.Library
	RiskProfiles *project.RiskProfileStore
	PO           *po.SessionManager
}

type Server struct {
	router     *chi.Mux
	hub        *Hub
	channelHub *ChannelHub
	deps       ServerDeps
}

func NewServer(deps ServerDeps) *Server {
	s := &Server{
		router:     chi.NewRouter(),
		hub:        NewHub(),
		channelHub: NewChannelHub(),
		deps:       deps,
	}

	s.setupRoutes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	r := s.router

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(LoggingMiddleware)
	r.Use(CORSMiddleware())

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		s.registerProjectRoutes(r)
		s.registerSpecRoutes(r)
		s.registerTaskRoutes(r)
		s.registerAgentRoutes(r)
		s.registerUsageRoutes(r)
		s.registerRiskProfileRoutes(r)
		s.registerPORoutes(r)
		s.registerLibraryRoutes(r)
	})

	r.Get("/ws", s.hub.HandleWebSocket)
	r.Get("/ws/projects/{id}/events", s.handleProjectEvents)
	r.Get("/ws/agents/{agentId}/logs", s.handleAgentLogs)
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// handleHealth returns the API health status.
//
// @Summary      Health check
// @Description  Returns the health status of the Foundry API
// @Tags         system
// @Produce      json
// @Success      200 {object} HealthResponse
// @Router       /health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

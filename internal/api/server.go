package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"github.com/cenron/foundry/internal/broker"
	"github.com/cenron/foundry/internal/cache"
)

type Server struct {
	router *chi.Mux
	db     *sqlx.DB
	cache  *cache.Client
	broker *broker.Client
	hub    *Hub
}

func NewServer(db *sqlx.DB, cache *cache.Client, broker *broker.Client) *Server {
	s := &Server{
		router: chi.NewRouter(),
		db:     db,
		cache:  cache,
		broker: broker,
		hub:    NewHub(),
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

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
	})

	r.Get("/ws", s.hub.HandleWebSocket)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

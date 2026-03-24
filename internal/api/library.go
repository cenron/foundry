package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/agent"
)

func (s *Server) registerLibraryRoutes(r chi.Router) {
	r.Get("/agents/library", s.handleListLibrary)
}

// handleListLibrary returns all available agent roles from the library.
//
// @Summary      List agent library
// @Description  Returns all available agent role definitions from the agent library
// @Tags         library
// @Produce      json
// @Success      200 {array} agent.AgentDefinition
// @Router       /agents/library [get]
func (s *Server) handleListLibrary(w http.ResponseWriter, r *http.Request) {
	if s.deps.Library == nil {
		RespondJSON(w, http.StatusOK, []agent.AgentDefinition{})
		return
	}

	definitions := s.deps.Library.LoadAll()
	RespondJSON(w, http.StatusOK, definitions)
}

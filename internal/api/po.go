package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/shared"
)

func (s *Server) registerPORoutes(r chi.Router) {
	r.Post("/projects/{id}/po/chat", s.handlePOChat)
	r.Delete("/projects/{id}/po/chat", s.handlePOChatDelete)
	r.Get("/projects/{id}/po/status", s.handlePOStatus)
	r.Post("/projects/{id}/po/planning", s.handlePOPlanning)
	r.Post("/projects/{id}/po/estimation", s.handlePOEstimation)
}

// handlePOChat is a stub for the PO chat endpoint (full implementation in Phase 6).
//
// @Summary      PO chat (stub)
// @Description  Stub endpoint — PO session manager not yet implemented
// @Tags         po
// @Produce      json
// @Param        id path string true "Project ID"
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      501 {object} ErrorResponse "Not implemented"
// @Router       /projects/{id}/po/chat [post]
func (s *Server) handlePOChat(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	RespondJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "PO session manager not yet implemented",
	})
}

// handlePOChatDelete is a stub for ending a PO chat session (full implementation in Phase 6).
//
// @Summary      End PO chat (stub)
// @Description  Stub endpoint — PO session manager not yet implemented
// @Tags         po
// @Produce      json
// @Param        id path string true "Project ID"
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      501 {object} ErrorResponse "Not implemented"
// @Router       /projects/{id}/po/chat [delete]
func (s *Server) handlePOChatDelete(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	RespondJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "PO session manager not yet implemented",
	})
}

// handlePOStatus is a stub for querying PO session status (full implementation in Phase 6).
//
// @Summary      PO status (stub)
// @Description  Stub endpoint — returns inactive status until PO session manager is implemented
// @Tags         po
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Router       /projects/{id}/po/status [get]
func (s *Server) handlePOStatus(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"active":  false,
		"message": "PO session manager not yet implemented",
	})
}

// handlePOPlanning is a stub for triggering PO planning (full implementation in Phase 6).
//
// @Summary      PO planning (stub)
// @Description  Stub endpoint — PO session manager not yet implemented
// @Tags         po
// @Produce      json
// @Param        id path string true "Project ID"
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      501 {object} ErrorResponse "Not implemented"
// @Router       /projects/{id}/po/planning [post]
func (s *Server) handlePOPlanning(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	RespondJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "PO session manager not yet implemented",
	})
}

// handlePOEstimation is a stub for triggering PO estimation (full implementation in Phase 6).
//
// @Summary      PO estimation (stub)
// @Description  Stub endpoint — PO session manager not yet implemented
// @Tags         po
// @Produce      json
// @Param        id path string true "Project ID"
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      501 {object} ErrorResponse "Not implemented"
// @Router       /projects/{id}/po/estimation [post]
func (s *Server) handlePOEstimation(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	RespondJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "PO session manager not yet implemented",
	})
}

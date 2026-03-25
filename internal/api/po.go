package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/po"
	"github.com/cenron/foundry/internal/shared"
)

func (s *Server) registerPORoutes(r chi.Router) {
	r.Post("/projects/{id}/po/chat", s.handlePOChat)
	r.Delete("/projects/{id}/po/chat", s.handlePOChatDelete)
	r.Get("/projects/{id}/po/status", s.handlePOStatus)
	r.Post("/projects/{id}/po/planning", s.handlePOPlanning)
	r.Post("/projects/{id}/po/estimation", s.handlePOEstimation)
}

// handlePOChat sends a message to the PO and starts/continues a chat session.
//
// @Summary      PO chat
// @Description  Send a message to the PO agent for the project
// @Tags         po
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      501 {object} ErrorResponse "Not implemented"
// @Router       /projects/{id}/po/chat [post]
func (s *Server) handlePOChat(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "PO session manager not configured",
		})
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, &shared.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	projectName := chi.URLParam(r, "id") // In production, resolve project name from ID
	_, err := s.deps.PO.StartSession(r.Context(), po.POSessionOpts{
		Type:    "execution-chat",
		Project: projectName,
		Trigger: "user",
		Message: req.Message,
	})
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"status": "session started"})
}

// handlePOChatDelete ends the active PO chat session.
//
// @Summary      End PO chat
// @Description  Close the active PO session for the project
// @Tags         po
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Router       /projects/{id}/po/chat [delete]
func (s *Server) handlePOChatDelete(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "PO session manager not configured",
		})
		return
	}

	projectName := chi.URLParam(r, "id")
	if err := s.deps.PO.StopSession(projectName); err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"status": "session stopped"})
}

// handlePOStatus returns whether a PO session is active.
//
// @Summary      PO status
// @Description  Check if a PO session is active for the project
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

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusOK, map[string]interface{}{
			"active":  false,
			"message": "PO session manager not configured",
		})
		return
	}

	projectName := chi.URLParam(r, "id")
	active := s.deps.PO.IsActive(projectName)

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"active": active,
	})
}

// handlePOPlanning starts a PO planning session.
//
// @Summary      Start PO planning
// @Description  Launch a PO planning session for the project
// @Tags         po
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      501 {object} ErrorResponse "Not implemented"
// @Router       /projects/{id}/po/planning [post]
func (s *Server) handlePOPlanning(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "PO session manager not configured",
		})
		return
	}

	projectName := chi.URLParam(r, "id")
	_, err := s.deps.PO.StartSession(r.Context(), po.POSessionOpts{
		Type:    "planning",
		Project: projectName,
		Trigger: "user",
		Message: "Start planning session.",
	})
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"status": "planning session started"})
}

// handlePOEstimation starts a PO estimation session.
//
// @Summary      Start PO estimation
// @Description  Launch a PO estimation session for the project
// @Tags         po
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      501 {object} ErrorResponse "Not implemented"
// @Router       /projects/{id}/po/estimation [post]
func (s *Server) handlePOEstimation(w http.ResponseWriter, r *http.Request) {
	if _, err := shared.ParseID(chi.URLParam(r, "id")); err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "PO session manager not configured",
		})
		return
	}

	projectName := chi.URLParam(r, "id")
	_, err := s.deps.PO.StartSession(r.Context(), po.POSessionOpts{
		Type:    "estimation",
		Project: projectName,
		Trigger: "system",
		Message: "Generate the execution plan per the playbook.",
	})
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"status": "estimation session started"})
}

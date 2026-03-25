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

// resolveProjectName returns the project name for the given UUID.
// Falls back to the raw UUID string when the project store is unavailable
// or the project is not found (e.g., in test environments with only PO configured).
func (s *Server) resolveProjectName(r *http.Request, id shared.ID) (string, error) {
	if s.deps.Projects == nil {
		return id.String(), nil
	}

	proj, err := s.deps.Projects.GetByID(r.Context(), id)
	if err != nil {
		return "", err
	}

	return proj.Name, nil
}

// poChatRequest is the request body for the PO chat endpoint.
type poChatRequest struct {
	Message string `json:"message"`
}

// handlePOChat sends a message to the PO and starts/continues a chat session.
//
// @Summary      PO chat
// @Description  Send a message to the PO agent for the project
// @Tags         po
// @Accept       json
// @Produce      json
// @Param        id   path string       true "Project ID"
// @Param        body body poChatRequest true "Chat message"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      501 {object} ErrorResponse "Not implemented"
// @Router       /projects/{id}/po/chat [post]
func (s *Server) handlePOChat(w http.ResponseWriter, r *http.Request) {
	parsedID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "PO session manager not configured",
		})
		return
	}

	var req poChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, &shared.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	projectName, err := s.resolveProjectName(r, parsedID)
	if err != nil {
		RespondError(w, err)
		return
	}

	_, err = s.deps.PO.StartSession(r.Context(), po.POSessionOpts{
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
	parsedID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "PO session manager not configured",
		})
		return
	}

	projectName, err := s.resolveProjectName(r, parsedID)
	if err != nil {
		RespondError(w, err)
		return
	}

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
	parsedID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
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

	projectName, err := s.resolveProjectName(r, parsedID)
	if err != nil {
		RespondError(w, err)
		return
	}

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
	parsedID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "PO session manager not configured",
		})
		return
	}

	projectName, err := s.resolveProjectName(r, parsedID)
	if err != nil {
		RespondError(w, err)
		return
	}

	_, err = s.deps.PO.StartSession(r.Context(), po.POSessionOpts{
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
	parsedID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.PO == nil {
		RespondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "PO session manager not configured",
		})
		return
	}

	projectName, err := s.resolveProjectName(r, parsedID)
	if err != nil {
		RespondError(w, err)
		return
	}

	_, err = s.deps.PO.StartSession(r.Context(), po.POSessionOpts{
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

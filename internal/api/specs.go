package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/project"
	"github.com/cenron/foundry/internal/shared"
)

// UpdateSpecRequest holds the body for updating spec content.
type UpdateSpecRequest struct {
	ApprovedContent  string `json:"approved_content"`
	ExecutionContent string `json:"execution_content"`
	TokenEstimate    int    `json:"token_estimate"`
	AgentCount       int    `json:"agent_count"`
}

func (s *Server) registerSpecRoutes(r chi.Router) {
	r.Get("/projects/{id}/spec", s.handleGetSpec)
	r.Put("/projects/{id}/spec", s.handleUpdateSpec)
	r.Post("/projects/{id}/spec/approve", s.handleApproveSpec)
	r.Post("/projects/{id}/spec/reject", s.handleRejectSpec)
}

// handleGetSpec returns the spec for a project.
//
// @Summary      Get spec
// @Description  Returns the spec associated with the given project
// @Tags         specs
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} project.Spec
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      404 {object} ErrorResponse "Spec not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/spec [get]
func (s *Server) handleGetSpec(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	spec, err := s.deps.Specs.GetByProjectID(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, spec)
}

// handleUpdateSpec updates the content of a project's spec.
//
// @Summary      Update spec
// @Description  Replaces the spec content for the given project; creates if none exists
// @Tags         specs
// @Accept       json
// @Produce      json
// @Param        id   path string          true "Project ID"
// @Param        body body UpdateSpecRequest true "Spec content"
// @Success      200 {object} project.Spec
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      404 {object} ErrorResponse "Project not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/spec [put]
func (s *Server) handleUpdateSpec(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var req UpdateSpecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, &shared.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	spec, err := s.deps.Specs.UpdateContent(r.Context(), projectID, project.CreateSpecParams{
		ProjectID:        projectID,
		ApprovedContent:  req.ApprovedContent,
		ExecutionContent: req.ExecutionContent,
		TokenEstimate:    req.TokenEstimate,
		AgentCount:       req.AgentCount,
	})
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, spec)
}

// handleApproveSpec approves a project's spec and advances the project to approved status.
//
// @Summary      Approve spec
// @Description  Approves the spec for the given project; requires content and token estimate
// @Tags         specs
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} project.Spec
// @Failure      400 {object} ErrorResponse "Spec not ready for approval"
// @Failure      404 {object} ErrorResponse "Spec not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/spec/approve [post]
func (s *Server) handleApproveSpec(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	spec, err := s.deps.Specs.GetByProjectID(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	if spec.ApprovedContent == "" {
		RespondError(w, &shared.ValidationError{Field: "approved_content", Message: "required before approval"})
		return
	}
	if spec.TokenEstimate == 0 {
		RespondError(w, &shared.ValidationError{Field: "token_estimate", Message: "required before approval"})
		return
	}

	if err := s.deps.Specs.UpdateApproval(r.Context(), spec.ID, "approved"); err != nil {
		RespondError(w, err)
		return
	}

	if err := s.deps.Projects.UpdateStatus(r.Context(), projectID, string(shared.ProjectStatusApproved)); err != nil {
		RespondError(w, err)
		return
	}

	spec.ApprovalStatus = "approved"
	RespondJSON(w, http.StatusOK, spec)
}

// handleRejectSpec rejects a project's spec and reverts the project to planning status.
//
// @Summary      Reject spec
// @Description  Rejects the spec for the given project and returns the project to planning
// @Tags         specs
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} project.Spec
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      404 {object} ErrorResponse "Spec not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/spec/reject [post]
func (s *Server) handleRejectSpec(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	spec, err := s.deps.Specs.GetByProjectID(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	if err := s.deps.Specs.UpdateApproval(r.Context(), spec.ID, "rejected"); err != nil {
		RespondError(w, err)
		return
	}

	if err := s.deps.Projects.UpdateStatus(r.Context(), projectID, string(shared.ProjectStatusPlanning)); err != nil {
		RespondError(w, err)
		return
	}

	spec.ApprovalStatus = "rejected"
	RespondJSON(w, http.StatusOK, spec)
}

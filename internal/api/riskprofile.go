package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/project"
	"github.com/cenron/foundry/internal/shared"
)

// UpdateRiskProfileRequest holds the body for updating a risk profile.
type UpdateRiskProfileRequest struct {
	Name           string          `json:"name"`
	LowCriteria    json.RawMessage `json:"low_criteria" swaggertype:"object"`
	MediumCriteria json.RawMessage `json:"medium_criteria" swaggertype:"object"`
	HighCriteria   json.RawMessage `json:"high_criteria" swaggertype:"object"`
	ModelRouting   json.RawMessage `json:"model_routing" swaggertype:"object"`
}

func (s *Server) registerRiskProfileRoutes(r chi.Router) {
	r.Get("/projects/{id}/risk-profile", s.handleGetRiskProfile)
	r.Put("/projects/{id}/risk-profile", s.handleUpdateRiskProfile)
}

// handleGetRiskProfile returns the risk profile for a project (project-specific or global default).
//
// @Summary      Get risk profile
// @Description  Returns the risk profile for the given project; falls back to the global default
// @Tags         risk-profiles
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} project.RiskProfile
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      404 {object} ErrorResponse "Risk profile not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/risk-profile [get]
func (s *Server) handleGetRiskProfile(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	profile, err := s.deps.RiskProfiles.GetByProjectID(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, profile)
}

// handleUpdateRiskProfile updates the risk profile for a project.
//
// @Summary      Update risk profile
// @Description  Updates the risk classification criteria and model routing for the given project
// @Tags         risk-profiles
// @Accept       json
// @Produce      json
// @Param        id   path string                  true "Project ID"
// @Param        body body UpdateRiskProfileRequest true "Risk profile fields"
// @Success      200 {object} project.RiskProfile
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      404 {object} ErrorResponse "Risk profile not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/risk-profile [put]
func (s *Server) handleUpdateRiskProfile(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var req UpdateRiskProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, &shared.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	existing, err := s.deps.RiskProfiles.GetByProjectID(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	name := existing.Name
	if req.Name != "" {
		name = req.Name
	}

	lowCriteria := existing.LowCriteria
	if len(req.LowCriteria) > 0 {
		lowCriteria = req.LowCriteria
	}

	mediumCriteria := existing.MediumCriteria
	if len(req.MediumCriteria) > 0 {
		mediumCriteria = req.MediumCriteria
	}

	highCriteria := existing.HighCriteria
	if len(req.HighCriteria) > 0 {
		highCriteria = req.HighCriteria
	}

	modelRouting := existing.ModelRouting
	if len(req.ModelRouting) > 0 {
		modelRouting = req.ModelRouting
	}

	profile, err := s.deps.RiskProfiles.Update(r.Context(), existing.ID, project.UpdateRiskProfileParams{
		Name:           name,
		LowCriteria:    lowCriteria,
		MediumCriteria: mediumCriteria,
		HighCriteria:   highCriteria,
		ModelRouting:   modelRouting,
	})
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, profile)
}

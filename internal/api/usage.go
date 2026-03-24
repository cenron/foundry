package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/shared"
)

// UsageResponse holds the token usage breakdown for a project.
type UsageResponse struct {
	ProjectID     string      `json:"project_id"`
	TotalTokens   int         `json:"total_tokens"`
	TaskBreakdown []TaskUsage `json:"task_breakdown"`
}

// TaskUsage holds token usage for a single task.
type TaskUsage struct {
	TaskID     string `json:"task_id"`
	Title      string `json:"title"`
	TokenUsage int    `json:"token_usage"`
	ModelTier  string `json:"model_tier"`
}

func (s *Server) registerUsageRoutes(r chi.Router) {
	r.Get("/projects/{id}/usage", s.handleGetUsage)
}

// handleGetUsage returns the token usage breakdown for a project.
//
// @Summary      Get usage
// @Description  Returns token usage aggregated across all tasks for the given project
// @Tags         usage
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} UsageResponse
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/usage [get]
func (s *Server) handleGetUsage(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	tasks, err := s.deps.Tasks.ListByProject(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	breakdown := make([]TaskUsage, 0, len(tasks))
	total := 0

	for _, t := range tasks {
		breakdown = append(breakdown, TaskUsage{
			TaskID:     t.ID.String(),
			Title:      t.Title,
			TokenUsage: t.TokenUsage,
			ModelTier:  t.ModelTier,
		})
		total += t.TokenUsage
	}

	RespondJSON(w, http.StatusOK, UsageResponse{
		ProjectID:     projectID.String(),
		TotalTokens:   total,
		TaskBreakdown: breakdown,
	})
}

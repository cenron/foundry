package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/shared"
)

func (s *Server) registerTaskRoutes(r chi.Router) {
	r.Get("/projects/{id}/tasks", s.handleListTasks)
	r.Get("/projects/{id}/tasks/{taskId}", s.handleGetTask)
}

// handleListTasks returns all tasks for a project, with optional status filter.
//
// @Summary      List tasks
// @Description  Returns all tasks for the given project; optionally filter by status
// @Tags         tasks
// @Produce      json
// @Param        id     path  string true  "Project ID"
// @Param        status query string false "Filter by task status"
// @Success      200 {array}  orchestrator.Task
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/tasks [get]
func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
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

	statusFilter := r.URL.Query().Get("status")
	if statusFilter == "" {
		RespondJSON(w, http.StatusOK, tasks)
		return
	}

	filtered := tasks[:0]
	for _, t := range tasks {
		if t.Status == statusFilter {
			filtered = append(filtered, t)
		}
	}

	RespondJSON(w, http.StatusOK, filtered)
}

// handleGetTask returns a single task by ID.
//
// @Summary      Get task
// @Description  Returns the task with the given ID
// @Tags         tasks
// @Produce      json
// @Param        id     path string true "Project ID"
// @Param        taskId path string true "Task ID"
// @Success      200 {object} orchestrator.Task
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      404 {object} ErrorResponse "Task not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/tasks/{taskId} [get]
func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	_, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	taskID, err := shared.ParseID(chi.URLParam(r, "taskId"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "taskId", Message: "invalid UUID"})
		return
	}

	task, err := s.deps.Tasks.GetByID(r.Context(), taskID)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, task)
}

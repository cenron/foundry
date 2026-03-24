package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/project"
	"github.com/cenron/foundry/internal/shared"
)

// ErrorResponse is used in swagger failure annotations.
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateProjectRequest holds the body for creating a project.
type CreateProjectRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	RepoURL         string   `json:"repo_url"`
	TeamComposition []string `json:"team_composition"`
}

// ProjectListResponse is the paginated response for project listing.
type ProjectListResponse struct {
	Data       []project.Project `json:"data"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalCount int               `json:"total_count"`
}

// UpdateProjectRequest holds the body for patching a project.
type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) registerProjectRoutes(r chi.Router) {
	r.Post("/projects", s.handleCreateProject)
	r.Get("/projects", s.handleListProjects)
	r.Get("/projects/{id}", s.handleGetProject)
	r.Patch("/projects/{id}", s.handleUpdateProject)
}

// handleCreateProject creates a new project.
//
// @Summary      Create project
// @Description  Creates a new project with the given details
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        body body CreateProjectRequest true "Project details"
// @Success      201 {object} project.Project
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects [post]
func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, &shared.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	if req.Name == "" {
		RespondError(w, &shared.ValidationError{Field: "name", Message: "required"})
		return
	}

	p, err := s.deps.Projects.Create(r.Context(), project.CreateProjectParams{
		Name:            req.Name,
		Description:     req.Description,
		RepoURL:         req.RepoURL,
		TeamComposition: req.TeamComposition,
	})
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusCreated, p)
}

// handleListProjects returns a paginated list of projects.
//
// @Summary      List projects
// @Description  Returns a paginated list of all projects
// @Tags         projects
// @Produce      json
// @Param        page      query int false "Page number (default 1)"
// @Param        page_size query int false "Items per page (default 20)"
// @Success      200 {object} ProjectListResponse
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects [get]
func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	projects, total, err := s.deps.Projects.List(r.Context(), page, pageSize)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, shared.PaginatedResponse[project.Project]{
		Data:       projects,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
	})
}

// handleGetProject returns a single project by ID.
//
// @Summary      Get project
// @Description  Returns the project with the given ID
// @Tags         projects
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} project.Project
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      404 {object} ErrorResponse "Project not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id} [get]
func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	p, err := s.deps.Projects.GetByID(r.Context(), id)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, p)
}

// handleUpdateProject updates a project's name and description.
//
// @Summary      Update project
// @Description  Updates the name and/or description of a project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id   path string             true "Project ID"
// @Param        body body UpdateProjectRequest true "Fields to update"
// @Success      200 {object} project.Project
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      404 {object} ErrorResponse "Project not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id} [patch]
func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, &shared.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	p, err := s.deps.Projects.GetByID(r.Context(), id)
	if err != nil {
		RespondError(w, err)
		return
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}

	RespondJSON(w, http.StatusOK, p)
}

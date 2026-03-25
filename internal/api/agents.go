package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/broker"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/runtime"
	"github.com/cenron/foundry/internal/shared"
)

// agentCommand is the payload published to the commands exchange.
type agentCommand struct {
	Action  string `json:"action"`
	AgentID string `json:"agent_id"`
}

// startProjectResponse is returned from handleStartProject.
type startProjectResponse struct {
	Status string              `json:"status"`
	Agent  *agent.Agent        `json:"agent"`
	Task   *orchestrator.Task  `json:"task"`
}

func (s *Server) registerAgentRoutes(r chi.Router) {
	r.Get("/projects/{id}/agents", s.handleListAgents)
	r.Get("/projects/{id}/agents/{agentId}", s.handleGetAgent)
	r.Post("/projects/{id}/agents/{agentId}/pause", s.handlePauseAgent)
	r.Post("/projects/{id}/agents/{agentId}/resume", s.handleResumeAgent)
	r.Post("/projects/{id}/start", s.handleStartProject)
	r.Post("/projects/{id}/pause", s.handlePauseProject)
	r.Post("/projects/{id}/resume", s.handleResumeProject)
}

// handleListAgents returns all agents for a project.
//
// @Summary      List agents
// @Description  Returns all agents for the given project
// @Tags         agents
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {array}  agent.Agent
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/agents [get]
func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	agents, err := s.deps.Agents.ListByProject(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, agents)
}

// handleGetAgent returns a single agent by ID.
//
// @Summary      Get agent
// @Description  Returns the agent with the given ID
// @Tags         agents
// @Produce      json
// @Param        id      path string true "Project ID"
// @Param        agentId path string true "Agent ID"
// @Success      200 {object} agent.Agent
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      404 {object} ErrorResponse "Agent not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/agents/{agentId} [get]
func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	_, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	agentID, err := shared.ParseID(chi.URLParam(r, "agentId"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "agentId", Message: "invalid UUID"})
		return
	}

	a, err := s.deps.Agents.GetByID(r.Context(), agentID)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, a)
}

// handlePauseAgent pauses a single agent by publishing a pause command.
//
// @Summary      Pause agent
// @Description  Publishes a pause command for the given agent
// @Tags         agents
// @Produce      json
// @Param        id      path string true "Project ID"
// @Param        agentId path string true "Agent ID"
// @Success      202 {object} map[string]string
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/agents/{agentId}/pause [post]
func (s *Server) handlePauseAgent(w http.ResponseWriter, r *http.Request) {
	s.publishAgentCommand(w, r, "pause")
}

// handleResumeAgent resumes a single agent by publishing a resume command.
//
// @Summary      Resume agent
// @Description  Publishes a resume command for the given agent
// @Tags         agents
// @Produce      json
// @Param        id      path string true "Project ID"
// @Param        agentId path string true "Agent ID"
// @Success      202 {object} map[string]string
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/agents/{agentId}/resume [post]
func (s *Server) handleResumeAgent(w http.ResponseWriter, r *http.Request) {
	s.publishAgentCommand(w, r, "resume")
}

func (s *Server) publishAgentCommand(w http.ResponseWriter, r *http.Request, action string) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	agentID, err := shared.ParseID(chi.URLParam(r, "agentId"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "agentId", Message: "invalid UUID"})
		return
	}

	body, err := json.Marshal(agentCommand{Action: action, AgentID: agentID.String()})
	if err != nil {
		RespondError(w, err)
		return
	}

	if s.deps.Broker == nil {
		RespondError(w, fmt.Errorf("broker not configured"))
		return
	}

	routingKey := fmt.Sprintf("commands.%s.%s", projectID, agentID)
	if err := s.deps.Broker.Publish(r.Context(), broker.ExchangeCommands, routingKey, body); err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusAccepted, map[string]string{"status": action})
}

// handleStartProject initiates project execution in local mode.
//
// @Summary      Start project
// @Description  Creates an agent and task, sets up workspace, and launches the claude process
// @Tags         projects
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      202 {object} startProjectResponse
// @Failure      400 {object} ErrorResponse "Invalid ID or missing runtime"
// @Failure      404 {object} ErrorResponse "Project not found"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/start [post]
func (s *Server) handleStartProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if s.deps.Runtime == nil {
		RespondError(w, &shared.ValidationError{Field: "runtime", Message: "local runtime not configured"})
		return
	}

	// Load and validate the project.
	proj, err := s.deps.Projects.GetByID(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	if proj.Status != "approved" {
		RespondError(w, &shared.ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("project must be approved before starting, current status: %s", proj.Status),
		})
		return
	}

	// Load the spec for the prompt.
	spec, err := s.deps.Specs.GetByProjectID(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	// Create the agent record.
	a, err := s.deps.Agents.Create(r.Context(), agent.CreateAgentParams{
		ProjectID:   projectID,
		Role:        "frontend-developer",
		Provider:    "claude",
		ContainerID: "local",
	})
	if err != nil {
		RespondError(w, fmt.Errorf("creating agent: %w", err))
		return
	}

	// Create the task record.
	task, err := s.deps.Tasks.Create(r.Context(), orchestrator.CreateTaskParams{
		ProjectID:    projectID,
		Title:        "Execute specification",
		Description:  "Implement the approved specification",
		AssignedRole: "frontend-developer",
	})
	if err != nil {
		RespondError(w, fmt.Errorf("creating task: %w", err))
		return
	}

	// Set up the workspace.
	if err := s.deps.Runtime.Setup(r.Context(), runtime.SetupOpts{
		ProjectID: projectID.String(),
		RepoURL:   proj.RepoURL,
		WorkDir:   s.deps.FoundryHome,
	}); err != nil {
		RespondError(w, fmt.Errorf("setting up workspace: %w", err))
		return
	}

	// Launch the agent process.
	agentProcess, err := s.deps.Runtime.LaunchAgent(r.Context(), runtime.AgentOpts{
		AgentID:   a.ID.String(),
		ProjectID: projectID.String(),
		Role:      a.Role,
		Prompt:    spec.ApprovedContent,
		WorkDir:   s.deps.FoundryHome,
	})
	if err != nil {
		RespondError(w, fmt.Errorf("launching agent: %w", err))
		return
	}

	// Mark task as in_progress.
	if err := s.deps.Tasks.UpdateStatus(r.Context(), task.ID, "in_progress"); err != nil {
		RespondError(w, fmt.Errorf("updating task status: %w", err))
		return
	}

	// Mark project as active.
	if err := s.deps.Projects.UpdateStatus(r.Context(), projectID, "active"); err != nil {
		RespondError(w, fmt.Errorf("updating project status: %w", err))
		return
	}

	// Reload updated task for response.
	task.Status = "in_progress"

	// Monitor the agent in the background: when it exits, mark task done and project completed.
	go s.monitorAgent(agentProcess, projectID.String(), task.ID.String())

	RespondJSON(w, http.StatusAccepted, startProjectResponse{
		Status: "starting",
		Agent:  a,
		Task:   task,
	})
}

// monitorAgent polls IsAgentRunning every 3 seconds and, when the agent exits,
// marks the task done and the project completed.
func (s *Server) monitorAgent(proc *runtime.AgentProcess, projectID, taskID string) {
	// Wait for the done signal from the agent process.
	<-proc.Done

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	taskID2, err := shared.ParseID(taskID)
	if err == nil {
		_ = s.deps.Tasks.UpdateStatus(ctx, taskID2, "done")
	}

	projectID2, err := shared.ParseID(projectID)
	if err == nil {
		_ = s.deps.Projects.UpdateStatus(ctx, projectID2, "completed")
	}
}

// handlePauseProject pauses all agents for a project.
//
// @Summary      Pause project
// @Description  Publishes pause commands for all agents in the project
// @Tags         projects
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      202 {object} map[string]string
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/pause [post]
func (s *Server) handlePauseProject(w http.ResponseWriter, r *http.Request) {
	s.publishProjectCommand(w, r, "pause")
}

// handleResumeProject resumes all agents for a project.
//
// @Summary      Resume project
// @Description  Publishes resume commands for all agents in the project
// @Tags         projects
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      202 {object} map[string]string
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      500 {object} ErrorResponse "Internal error"
// @Router       /projects/{id}/resume [post]
func (s *Server) handleResumeProject(w http.ResponseWriter, r *http.Request) {
	s.publishProjectCommand(w, r, "resume")
}

func (s *Server) publishProjectCommand(w http.ResponseWriter, r *http.Request, action string) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	agents, err := s.deps.Agents.ListByProject(r.Context(), projectID)
	if err != nil {
		RespondError(w, err)
		return
	}

	if s.deps.Broker == nil {
		RespondError(w, fmt.Errorf("broker not configured"))
		return
	}

	for _, a := range agents {
		body, err := json.Marshal(agentCommand{Action: action, AgentID: a.ID.String()})
		if err != nil {
			RespondError(w, err)
			return
		}

		routingKey := fmt.Sprintf("commands.%s.%s", projectID, a.ID)
		if err := s.deps.Broker.Publish(r.Context(), broker.ExchangeCommands, routingKey, body); err != nil {
			RespondError(w, err)
			return
		}
	}

	RespondJSON(w, http.StatusAccepted, map[string]string{"status": action})
}

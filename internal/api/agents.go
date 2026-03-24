package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cenron/foundry/internal/broker"
	"github.com/cenron/foundry/internal/shared"
)

// agentCommand is the payload published to the commands exchange.
type agentCommand struct {
	Action  string `json:"action"`
	AgentID string `json:"agent_id"`
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

	routingKey := fmt.Sprintf("commands.%s.%s", projectID, agentID)
	if err := s.deps.Broker.Publish(r.Context(), broker.ExchangeCommands, routingKey, body); err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusAccepted, map[string]string{"status": action})
}

// handleStartProject stubs the project start operation (full impl in Phase 6).
//
// @Summary      Start project
// @Description  Initiates project execution (stub — full implementation in Phase 6)
// @Tags         projects
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      202 {object} map[string]string
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Router       /projects/{id}/start [post]
func (s *Server) handleStartProject(w http.ResponseWriter, r *http.Request) {
	_, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, &shared.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	RespondJSON(w, http.StatusAccepted, map[string]string{"status": "starting"})
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

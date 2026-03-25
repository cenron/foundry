package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/broker"
	"github.com/cenron/foundry/internal/shared"
)

// AgentPauser lists and updates agent status for pause/resume operations.
type AgentPauser interface {
	ListByProject(ctx context.Context, projectID shared.ID) ([]agent.Agent, error)
	UpdateStatus(ctx context.Context, id shared.ID, status string) error
}

// PauseManager handles pause and resume operations for agents and projects.
type PauseManager struct {
	projects ProjectReader
	agents   AgentPauser
	tasks    *TaskStore
	sm       *StateMachine
	commands CommandPublisher
}

// NewPauseManager constructs a PauseManager with all required dependencies.
func NewPauseManager(
	projects ProjectReader,
	agents AgentPauser,
	tasks *TaskStore,
	sm *StateMachine,
	commands CommandPublisher,
) *PauseManager {
	return &PauseManager{
		projects: projects,
		agents:   agents,
		tasks:    tasks,
		sm:       sm,
		commands: commands,
	}
}

// PauseAgent sends a pause command to the agent, updates its status, and pauses
// any task it is currently working on.
func (m *PauseManager) PauseAgent(ctx context.Context, a agent.Agent) error {
	if err := m.sendAgentCommand(ctx, a, "pause_agent"); err != nil {
		return fmt.Errorf("sending pause command to agent %s: %w", a.ID, err)
	}

	if err := m.agents.UpdateStatus(ctx, a.ID, string(shared.AgentStatusPaused)); err != nil {
		return fmt.Errorf("updating agent status to paused: %w", err)
	}

	if a.CurrentTaskID != nil {
		if err := m.sm.Transition(ctx, *a.CurrentTaskID, "paused"); err != nil {
			log.Printf("pause manager: transitioning task %s to paused: %v", *a.CurrentTaskID, err)
		}
	}

	return nil
}

// PauseProject pauses all active agents belonging to the project.
func (m *PauseManager) PauseProject(ctx context.Context, projectID shared.ID) error {
	agents, err := m.agents.ListByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("listing agents for project %s: %w", projectID, err)
	}

	for _, a := range agents {
		if a.Status != string(shared.AgentStatusActive) {
			continue
		}
		if err := m.PauseAgent(ctx, a); err != nil {
			log.Printf("pause manager: pausing agent %s: %v", a.ID, err)
		}
	}

	if err := m.projects.UpdateStatus(ctx, projectID, string(shared.ProjectStatusPaused)); err != nil {
		return fmt.Errorf("marking project paused: %w", err)
	}

	return nil
}

// ResumeAgent updates the agent status to active, re-queues its paused task,
// and sends a resume command.
func (m *PauseManager) ResumeAgent(ctx context.Context, a agent.Agent) error {
	if err := m.agents.UpdateStatus(ctx, a.ID, string(shared.AgentStatusActive)); err != nil {
		return fmt.Errorf("updating agent status to active: %w", err)
	}

	if a.CurrentTaskID != nil {
		if err := m.sm.Transition(ctx, *a.CurrentTaskID, "assigned"); err != nil {
			log.Printf("pause manager: re-queuing task %s for agent %s: %v", *a.CurrentTaskID, a.ID, err)
		}
	}

	if err := m.sendAgentCommand(ctx, a, "resume_agent"); err != nil {
		return fmt.Errorf("sending resume command to agent %s: %w", a.ID, err)
	}

	return nil
}

// ResumeProject resumes all paused agents belonging to the project.
func (m *PauseManager) ResumeProject(ctx context.Context, projectID shared.ID) error {
	agents, err := m.agents.ListByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("listing agents for project %s: %w", projectID, err)
	}

	for _, a := range agents {
		if a.Status != string(shared.AgentStatusPaused) {
			continue
		}
		if err := m.ResumeAgent(ctx, a); err != nil {
			log.Printf("pause manager: resuming agent %s: %v", a.ID, err)
		}
	}

	if err := m.projects.UpdateStatus(ctx, projectID, string(shared.ProjectStatusActive)); err != nil {
		return fmt.Errorf("marking project active: %w", err)
	}

	return nil
}

func (m *PauseManager) sendAgentCommand(ctx context.Context, a agent.Agent, cmdType string) error {
	body, err := json.Marshal(map[string]string{
		"type":       cmdType,
		"agent_id":   a.ID.String(),
		"project_id": a.ProjectID.String(),
	})
	if err != nil {
		return fmt.Errorf("marshaling %s command: %w", cmdType, err)
	}

	routingKey := fmt.Sprintf("commands.%s.%s", a.ProjectID.String(), a.ID.String())
	return m.commands.Publish(ctx, broker.ExchangeCommands, routingKey, body)
}

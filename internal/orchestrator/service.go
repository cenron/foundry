package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cenron/foundry/internal/shared"
)

// AgentFinder finds available agents for task assignment.
type AgentFinder interface {
	ListAvailableByProject(ctx context.Context, projectID shared.ID) ([]AvailableAgent, error)
	UpdateCurrentTask(ctx context.Context, agentID shared.ID, taskID *shared.ID) error
}

// AvailableAgent is a lightweight view of an agent ready for assignment.
type AvailableAgent struct {
	ID   shared.ID
	Role string
}

// CommandPublisher sends commands to agents via the message broker.
type CommandPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

type Service struct {
	tasks    TaskStateStore
	dag      *DAGResolver
	sm       *StateMachine
	agents   AgentFinder
	commands CommandPublisher
}

func NewService(
	tasks TaskStateStore,
	dag *DAGResolver,
	sm *StateMachine,
	agents AgentFinder,
	commands CommandPublisher,
) *Service {
	return &Service{
		tasks:    tasks,
		dag:      dag,
		sm:       sm,
		agents:   agents,
		commands: commands,
	}
}

// HandleTaskCompleted processes a task completion event:
// 1. Transition task to done
// 2. Find newly unblocked tasks
// 3. Match to available agents by role
// 4. Assign and send commands
func (s *Service) HandleTaskCompleted(ctx context.Context, projectID, taskID shared.ID) error {
	if err := s.sm.Transition(ctx, taskID, "done"); err != nil {
		return fmt.Errorf("transitioning task to done: %w", err)
	}

	return s.assignUnblockedTasks(ctx, projectID)
}

// HandleAgentUnhealthy pauses the agent's current task.
func (s *Service) HandleAgentUnhealthy(ctx context.Context, taskID shared.ID) error {
	if err := s.sm.Transition(ctx, taskID, "paused"); err != nil {
		return fmt.Errorf("pausing task for unhealthy agent: %w", err)
	}
	return nil
}

// AssignUnblockedTasks finds pending tasks with resolved deps and assigns them.
func (s *Service) assignUnblockedTasks(ctx context.Context, projectID shared.ID) error {
	unblocked, err := s.dag.GetUnblockedTasks(ctx, projectID)
	if err != nil {
		return fmt.Errorf("getting unblocked tasks: %w", err)
	}

	if len(unblocked) == 0 {
		return nil
	}

	available, err := s.agents.ListAvailableByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("listing available agents: %w", err)
	}

	agentsByRole := indexAgentsByRole(available)

	for i := range unblocked {
		task := &unblocked[i]
		agents, ok := agentsByRole[task.AssignedRole]
		if !ok || len(agents) == 0 {
			continue
		}

		agent := agents[0]
		agentsByRole[task.AssignedRole] = agents[1:]

		if err := s.assignTask(ctx, task, agent); err != nil {
			log.Printf("orchestrator: assigning task %s to agent %s: %v", task.ID, agent.ID, err)
			continue
		}
	}

	return nil
}

func (s *Service) assignTask(ctx context.Context, task *Task, agent AvailableAgent) error {
	if err := s.sm.Transition(ctx, task.ID, "assigned"); err != nil {
		return fmt.Errorf("transitioning to assigned: %w", err)
	}

	if err := s.agents.UpdateCurrentTask(ctx, agent.ID, &task.ID); err != nil {
		return fmt.Errorf("updating agent current task: %w", err)
	}

	return s.sendAssignCommand(ctx, task, agent)
}

func (s *Service) sendAssignCommand(ctx context.Context, task *Task, agent AvailableAgent) error {
	cmd, _ := json.Marshal(map[string]string{
		"type":       "assign_task",
		"task_id":    task.ID.String(),
		"project_id": task.ProjectID.String(),
		"agent_role": agent.Role,
		"title":      task.Title,
		"description": task.Description,
	})

	routingKey := fmt.Sprintf("commands.%s.%s", task.ProjectID, agent.ID)
	return s.commands.Publish(ctx, "foundry.commands", routingKey, cmd)
}

func indexAgentsByRole(agents []AvailableAgent) map[string][]AvailableAgent {
	m := make(map[string][]AvailableAgent)
	for _, a := range agents {
		m[a.Role] = append(m[a.Role], a)
	}
	return m
}

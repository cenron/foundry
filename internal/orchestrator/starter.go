package orchestrator

import (
	"context"
	"fmt"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/container"
	"github.com/cenron/foundry/internal/project"
	"github.com/cenron/foundry/internal/shared"
)

// ProjectReader reads and updates project records.
type ProjectReader interface {
	GetByID(ctx context.Context, id shared.ID) (*project.Project, error)
	UpdateStatus(ctx context.Context, id shared.ID, status string) error
}

// AgentCreator creates agent records.
type AgentCreator interface {
	Create(ctx context.Context, params agent.CreateAgentParams) (*agent.Agent, error)
}

// ContainerCreator manages team containers.
type ContainerCreator interface {
	CreateTeam(ctx context.Context, cfg container.TeamContainerConfig) (string, error)
	StartTeam(ctx context.Context, containerID string) error
}

// TaskCreator persists new task records.
type TaskCreator interface {
	Create(ctx context.Context, params CreateTaskParams) (*Task, error)
}

// UnblockedTasksReader finds tasks that are ready to execute.
type UnblockedTasksReader interface {
	GetUnblockedTasks(ctx context.Context, projectID shared.ID) ([]Task, error)
}

// TierResolverFunc resolves the model tier for a risk level and provider.
type TierResolverFunc func(riskLevel, provider string) string

// TaskDef describes a task to create when starting a project.
type TaskDef struct {
	Title        string
	Description  string
	RiskLevel    string
	AssignedRole string
	DependsOn    []string // titles of tasks this depends on
}

// StartConfig carries the team composition and task plan for a project start.
type StartConfig struct {
	Roles    []string
	Tasks    []TaskDef
	Provider string // default agent provider (e.g. "claude")
}

// StartProjectParams is the input to ProjectStarter.StartProject.
type StartProjectParams struct {
	ProjectID   shared.ID
	FoundryHome string
	Config      StartConfig
}

// ProjectStarter bootstraps a project: creates agents, spins up the container,
// seeds task records, and fires the initial assignment wave.
type ProjectStarter struct {
	projects   ProjectReader
	agents     AgentCreator
	tasks      TaskCreator
	unblocked  UnblockedTasksReader
	sm         *StateMachine
	containers ContainerCreator
	commands   CommandPublisher
	tiers      TierResolverFunc
}

// NewProjectStarter constructs a ProjectStarter with all required dependencies.
func NewProjectStarter(
	projects ProjectReader,
	agents AgentCreator,
	tasks TaskCreator,
	unblocked UnblockedTasksReader,
	sm *StateMachine,
	containers ContainerCreator,
	commands CommandPublisher,
	tiers TierResolverFunc,
) *ProjectStarter {
	return &ProjectStarter{
		projects:   projects,
		agents:     agents,
		tasks:      tasks,
		unblocked:  unblocked,
		sm:         sm,
		containers: containers,
		commands:   commands,
		tiers:      tiers,
	}
}

// StartProject validates the project is approved, creates agents, starts the
// team container, seeds tasks, and dispatches the initial unblocked work.
func (s *ProjectStarter) StartProject(ctx context.Context, params StartProjectParams) error {
	proj, err := s.projects.GetByID(ctx, params.ProjectID)
	if err != nil {
		return fmt.Errorf("loading project: %w", err)
	}

	if proj.Status != string(shared.ProjectStatusApproved) {
		return &shared.ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("project must be approved to start, current status: %q", proj.Status),
		}
	}

	containerID, err := s.provisionContainer(ctx, params)
	if err != nil {
		return err
	}

	createdAgents, err := s.createAgents(ctx, params, containerID)
	if err != nil {
		return err
	}

	taskIndex, err := s.createTasks(ctx, params)
	if err != nil {
		return err
	}

	if err := s.assignInitialTasks(ctx, params.ProjectID, taskIndex, createdAgents); err != nil {
		return err
	}

	if err := s.projects.UpdateStatus(ctx, params.ProjectID, string(shared.ProjectStatusActive)); err != nil {
		return fmt.Errorf("marking project active: %w", err)
	}

	return nil
}

func (s *ProjectStarter) provisionContainer(ctx context.Context, params StartProjectParams) (string, error) {
	cfg := container.TeamContainerConfig{
		ProjectID:       params.ProjectID.String(),
		SharedVolPath:   params.FoundryHome + "/projects/" + params.ProjectID.String() + "/shared",
		TeamComposition: params.Config.Roles,
	}

	containerID, err := s.containers.CreateTeam(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("creating team container: %w", err)
	}

	if err := s.containers.StartTeam(ctx, containerID); err != nil {
		return "", fmt.Errorf("starting team container: %w", err)
	}

	return containerID, nil
}

func (s *ProjectStarter) createAgents(ctx context.Context, params StartProjectParams, containerID string) ([]agent.Agent, error) {
	provider := params.Config.Provider
	if provider == "" {
		provider = "claude"
	}

	agents := make([]agent.Agent, 0, len(params.Config.Roles))
	for _, role := range params.Config.Roles {
		a, err := s.agents.Create(ctx, agent.CreateAgentParams{
			ProjectID:   params.ProjectID,
			Role:        role,
			Provider:    provider,
			ContainerID: containerID,
		})
		if err != nil {
			return nil, fmt.Errorf("creating agent for role %q: %w", role, err)
		}
		agents = append(agents, *a)
	}

	return agents, nil
}

// createTasks seeds all task records and returns a map from task title to Task
// so dependency resolution can reference tasks by name.
func (s *ProjectStarter) createTasks(ctx context.Context, params StartProjectParams) (map[string]*Task, error) {
	// First pass: create all tasks without dependencies to get their IDs.
	titleToTask := make(map[string]*Task, len(params.Config.Tasks))
	for _, def := range params.Config.Tasks {
		task, err := s.tasks.Create(ctx, CreateTaskParams{
			ProjectID:    params.ProjectID,
			Title:        def.Title,
			Description:  def.Description,
			RiskLevel:    def.RiskLevel,
			AssignedRole: def.AssignedRole,
			DependsOn:    nil,
		})
		if err != nil {
			return nil, fmt.Errorf("creating task %q: %w", def.Title, err)
		}
		titleToTask[def.Title] = task
	}

	return titleToTask, nil
}

func (s *ProjectStarter) assignInitialTasks(
	ctx context.Context,
	projectID shared.ID,
	taskIndex map[string]*Task,
	agents []agent.Agent,
) error {
	unblocked, err := s.unblocked.GetUnblockedTasks(ctx, projectID)
	if err != nil {
		return fmt.Errorf("getting initial unblocked tasks: %w", err)
	}

	agentsByRole := indexCreatedAgentsByRole(agents)

	for i := range unblocked {
		task := &unblocked[i]
		roleAgents, ok := agentsByRole[task.AssignedRole]
		if !ok || len(roleAgents) == 0 {
			continue
		}

		a := roleAgents[0]
		agentsByRole[task.AssignedRole] = roleAgents[1:]

		available := AvailableAgent{ID: a.ID, Role: a.Role}
		if err := s.assignTask(ctx, task, available); err != nil {
			// Non-fatal: log and continue so other tasks still get assigned.
			continue
		}
	}

	return nil
}

func (s *ProjectStarter) assignTask(ctx context.Context, task *Task, a AvailableAgent) error {
	if err := s.sm.Transition(ctx, task.ID, "assigned"); err != nil {
		return fmt.Errorf("transitioning task to assigned: %w", err)
	}

	return publishAssignCommand(ctx, s.commands, task, a)
}

func indexCreatedAgentsByRole(agents []agent.Agent) map[string][]agent.Agent {
	m := make(map[string][]agent.Agent)
	for _, a := range agents {
		m[a.Role] = append(m[a.Role], a)
	}
	return m
}

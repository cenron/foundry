package orchestrator_test

import (
	"context"
	"testing"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/container"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/project"
	"github.com/cenron/foundry/internal/shared"
)

// --- mocks ---

type mockProjectReader struct {
	projects      map[shared.ID]*project.Project
	statusUpdates []projectStatusUpdate
}

type projectStatusUpdate struct {
	id     shared.ID
	status string
}

func newMockProjectReader() *mockProjectReader {
	return &mockProjectReader{projects: make(map[shared.ID]*project.Project)}
}

func (m *mockProjectReader) addProject(status string) *project.Project {
	p := &project.Project{
		ID:     shared.NewID(),
		Name:   "test-project",
		Status: status,
	}
	m.projects[p.ID] = p
	return p
}

func (m *mockProjectReader) GetByID(_ context.Context, id shared.ID) (*project.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, &shared.NotFoundError{Resource: "project", ID: id.String()}
	}
	cp := *p
	return &cp, nil
}

func (m *mockProjectReader) UpdateStatus(_ context.Context, id shared.ID, status string) error {
	m.statusUpdates = append(m.statusUpdates, projectStatusUpdate{id, status})
	if p, ok := m.projects[id]; ok {
		p.Status = status
	}
	return nil
}

type mockAgentCreator struct {
	created []agent.Agent
}

func (m *mockAgentCreator) Create(_ context.Context, params agent.CreateAgentParams) (*agent.Agent, error) {
	a := agent.Agent{
		ID:          shared.NewID(),
		ProjectID:   params.ProjectID,
		Role:        params.Role,
		Provider:    params.Provider,
		ContainerID: params.ContainerID,
		Status:      "active",
	}
	m.created = append(m.created, a)
	return &a, nil
}

type mockContainerCreator struct {
	createdConfigs []container.TeamContainerConfig
	startedIDs     []string
	containerID    string
}

func (m *mockContainerCreator) CreateTeam(_ context.Context, cfg container.TeamContainerConfig) (string, error) {
	m.createdConfigs = append(m.createdConfigs, cfg)
	if m.containerID == "" {
		m.containerID = "container-abc123"
	}
	return m.containerID, nil
}

func (m *mockContainerCreator) StartTeam(_ context.Context, containerID string) error {
	m.startedIDs = append(m.startedIDs, containerID)
	return nil
}

// --- helpers ---

// mockTaskCreator records tasks created via the TaskCreator interface.
type mockTaskCreator struct {
	created []*orchestrator.Task
}

func (m *mockTaskCreator) Create(_ context.Context, params orchestrator.CreateTaskParams) (*orchestrator.Task, error) {
	task := &orchestrator.Task{
		ID:           shared.NewID(),
		ProjectID:    params.ProjectID,
		Title:        params.Title,
		Description:  params.Description,
		RiskLevel:    params.RiskLevel,
		AssignedRole: params.AssignedRole,
		Status:       "pending",
	}
	m.created = append(m.created, task)
	return task, nil
}

func setupStarter(t *testing.T) (
	*orchestrator.ProjectStarter,
	*mockProjectReader,
	*mockAgentCreator,
	*mockContainerCreator,
	*mockTaskStateStore,
	*mockTaskCreator,
	*recordingPublisher,
) {
	t.Helper()

	projects := newMockProjectReader()
	agents := &mockAgentCreator{}
	containers := &mockContainerCreator{}
	taskCreator := &mockTaskCreator{}

	store := newMockTaskStateStore()
	pub := &recordingPublisher{}
	sm := orchestrator.NewStateMachine(store, pub)

	unblocker := &mockUnblockedFinder{}

	starter := orchestrator.NewProjectStarter(
		projects,
		agents,
		taskCreator,
		unblocker,
		sm,
		containers,
		pub,
		func(riskLevel, provider string) string { return "sonnet" },
	)

	return starter, projects, agents, containers, store, taskCreator, pub
}

func makeStartParams(projectID shared.ID) orchestrator.StartProjectParams {
	return orchestrator.StartProjectParams{
		ProjectID:   projectID,
		FoundryHome: "/home/user/foundry",
		Config: orchestrator.StartConfig{
			Roles:    []string{"backend-developer", "frontend-developer"},
			Provider: "claude",
			Tasks: []orchestrator.TaskDef{
				{
					Title:        "Setup repo",
					Description:  "Init the repository",
					RiskLevel:    "low",
					AssignedRole: "backend-developer",
					DependsOn:    nil,
				},
				{
					Title:        "Build UI",
					Description:  "Create frontend components",
					RiskLevel:    "medium",
					AssignedRole: "frontend-developer",
					DependsOn:    []string{"Setup repo"},
				},
			},
		},
	}
}

// --- tests ---

func TestProjectStarter_RejectsNonApprovedProject(t *testing.T) {
	starter, projects, _, _, _, _, _ := setupStarter(t)

	p := projects.addProject("draft")
	params := makeStartParams(p.ID)

	err := starter.StartProject(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for non-approved project, got nil")
	}

	var valErr *shared.ValidationError
	if !isValidationError(err, &valErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestProjectStarter_CreatesAgentsForEachRole(t *testing.T) {
	starter, projects, agentCreator, _, _, _, _ := setupStarter(t)

	p := projects.addProject("approved")
	params := makeStartParams(p.ID)

	if err := starter.StartProject(context.Background(), params); err != nil {
		t.Fatalf("StartProject() error: %v", err)
	}

	if len(agentCreator.created) != len(params.Config.Roles) {
		t.Errorf("created %d agents, want %d", len(agentCreator.created), len(params.Config.Roles))
	}

	rolesSeen := make(map[string]bool)
	for _, a := range agentCreator.created {
		rolesSeen[a.Role] = true
	}
	for _, role := range params.Config.Roles {
		if !rolesSeen[role] {
			t.Errorf("no agent created for role %q", role)
		}
	}
}

func TestProjectStarter_CreatesAndStartsContainer(t *testing.T) {
	starter, projects, _, containers, _, _, _ := setupStarter(t)

	p := projects.addProject("approved")
	params := makeStartParams(p.ID)

	if err := starter.StartProject(context.Background(), params); err != nil {
		t.Fatalf("StartProject() error: %v", err)
	}

	if len(containers.createdConfigs) != 1 {
		t.Fatalf("expected 1 container created, got %d", len(containers.createdConfigs))
	}
	if len(containers.startedIDs) != 1 {
		t.Fatalf("expected 1 container started, got %d", len(containers.startedIDs))
	}
	if containers.startedIDs[0] != containers.containerID {
		t.Errorf("started container ID = %q, want %q", containers.startedIDs[0], containers.containerID)
	}
}

func TestProjectStarter_UpdatesProjectStatusToActive(t *testing.T) {
	starter, projects, _, _, _, _, _ := setupStarter(t)

	p := projects.addProject("approved")
	params := makeStartParams(p.ID)

	if err := starter.StartProject(context.Background(), params); err != nil {
		t.Fatalf("StartProject() error: %v", err)
	}

	if len(projects.statusUpdates) == 0 {
		t.Fatal("expected project status update, got none")
	}

	last := projects.statusUpdates[len(projects.statusUpdates)-1]
	if last.status != "active" {
		t.Errorf("final project status = %q, want active", last.status)
	}
}

func TestProjectStarter_RejectsPlanningStatus(t *testing.T) {
	starter, projects, _, _, _, _, _ := setupStarter(t)

	for _, status := range []string{"draft", "planning", "estimated", "active", "paused"} {
		t.Run(status, func(t *testing.T) {
			p := projects.addProject(status)
			params := makeStartParams(p.ID)

			err := starter.StartProject(context.Background(), params)
			if err == nil {
				t.Errorf("expected error for status %q, got nil", status)
			}
		})
	}
}

// isValidationError checks if err is (or wraps) a *shared.ValidationError.
func isValidationError(err error, target **shared.ValidationError) bool {
	if err == nil {
		return false
	}
	if ve, ok := err.(*shared.ValidationError); ok {
		if target != nil {
			*target = ve
		}
		return true
	}
	return false
}

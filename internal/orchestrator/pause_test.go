package orchestrator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/cenron/foundry/internal/agent"
	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/project"
	"github.com/cenron/foundry/internal/shared"
)

// --- mocks ---

type mockAgentPauser struct {
	agents        []agent.Agent
	statusUpdates []agentStatusUpdate
}

type agentStatusUpdate struct {
	id     shared.ID
	status string
}

func (m *mockAgentPauser) ListByProject(_ context.Context, _ shared.ID) ([]agent.Agent, error) {
	return m.agents, nil
}

func (m *mockAgentPauser) UpdateStatus(_ context.Context, id shared.ID, status string) error {
	m.statusUpdates = append(m.statusUpdates, agentStatusUpdate{id, status})
	for i := range m.agents {
		if m.agents[i].ID == id {
			m.agents[i].Status = status
		}
	}
	return nil
}

// --- helpers ---

func newActiveAgent(projectID shared.ID) agent.Agent {
	return agent.Agent{
		ID:        shared.NewID(),
		ProjectID: projectID,
		Role:      "backend-developer",
		Status:    string(shared.AgentStatusActive),
	}
}

func newPausedAgent(projectID shared.ID) agent.Agent {
	a := newActiveAgent(projectID)
	a.Status = string(shared.AgentStatusPaused)
	return a
}

func setupPauseManager(t *testing.T) (
	*orchestrator.PauseManager,
	*mockProjectReader,
	*mockAgentPauser,
	*mockTaskStateStore,
	*recordingPublisher,
) {
	t.Helper()

	projects := newMockProjectReader()
	agentPauser := &mockAgentPauser{}
	taskStore := newMockTaskStateStore()

	pub := &recordingPublisher{}
	sm := orchestrator.NewStateMachine(taskStore, pub)

	// PauseManager uses *TaskStore only for access via StateMachine internally.
	// Pass nil for the direct *TaskStore reference since we drive transitions through sm.
	pm := orchestrator.NewPauseManager(projects, agentPauser, nil, sm, pub)

	return pm, projects, agentPauser, taskStore, pub
}

// --- tests ---

func TestPauseManager_PauseAgent_SendsCommandAndUpdatesStatus(t *testing.T) {
	pm, _, agentPauser, _, pub := setupPauseManager(t)

	projectID := shared.NewID()
	a := newActiveAgent(projectID)
	agentPauser.agents = []agent.Agent{a}

	err := pm.PauseAgent(context.Background(), a)
	if err != nil {
		t.Fatalf("PauseAgent() error: %v", err)
	}

	// Status updated to paused
	if len(agentPauser.statusUpdates) != 1 {
		t.Fatalf("expected 1 status update, got %d", len(agentPauser.statusUpdates))
	}
	if agentPauser.statusUpdates[0].status != "paused" {
		t.Errorf("agent status = %q, want paused", agentPauser.statusUpdates[0].status)
	}

	// Pause command published
	pub.mu.Lock()
	defer pub.mu.Unlock()
	if !hasCommandOfType(pub.messages, "pause_agent") {
		t.Error("expected pause_agent command to be published")
	}
}

func TestPauseManager_PauseAgent_WithCurrentTask_AlsoPausesTask(t *testing.T) {
	pm, _, agentPauser, taskStore, _ := setupPauseManager(t)

	projectID := shared.NewID()
	task := taskStore.addTask("in_progress")
	task.ProjectID = projectID
	taskStore.tasks[task.ID] = task

	a := newActiveAgent(projectID)
	a.CurrentTaskID = &task.ID
	agentPauser.agents = []agent.Agent{a}

	err := pm.PauseAgent(context.Background(), a)
	if err != nil {
		t.Fatalf("PauseAgent() error: %v", err)
	}

	got, _ := taskStore.GetByID(context.Background(), task.ID)
	if got.Status != "paused" {
		t.Errorf("task status = %q, want paused", got.Status)
	}
}

func TestPauseManager_PauseProject_PausesAllActiveAgents(t *testing.T) {
	pm, projects, agentPauser, _, _ := setupPauseManager(t)

	projectID := shared.NewID()
	projects.projects[projectID] = &project.Project{ID: projectID, Status: "active"}

	a1 := newActiveAgent(projectID)
	a2 := newActiveAgent(projectID)
	inactive := newPausedAgent(projectID) // already paused — should be skipped
	agentPauser.agents = []agent.Agent{a1, a2, inactive}

	err := pm.PauseProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("PauseProject() error: %v", err)
	}

	// Two active agents should have been paused
	pausedCount := 0
	for _, upd := range agentPauser.statusUpdates {
		if upd.status == "paused" {
			pausedCount++
		}
	}
	if pausedCount != 2 {
		t.Errorf("paused %d agents, want 2", pausedCount)
	}

	// Project status updated to paused
	if !hasProjectStatusUpdate(projects.statusUpdates, projectID, "paused") {
		t.Error("expected project status to be updated to paused")
	}
}

func TestPauseManager_ResumeAgent_SendsCommandAndUpdatesStatus(t *testing.T) {
	pm, _, agentPauser, _, pub := setupPauseManager(t)

	projectID := shared.NewID()
	a := newPausedAgent(projectID)
	agentPauser.agents = []agent.Agent{a}

	err := pm.ResumeAgent(context.Background(), a)
	if err != nil {
		t.Fatalf("ResumeAgent() error: %v", err)
	}

	// Status updated to active
	if len(agentPauser.statusUpdates) != 1 {
		t.Fatalf("expected 1 status update, got %d", len(agentPauser.statusUpdates))
	}
	if agentPauser.statusUpdates[0].status != "active" {
		t.Errorf("agent status = %q, want active", agentPauser.statusUpdates[0].status)
	}

	// Resume command published
	pub.mu.Lock()
	defer pub.mu.Unlock()
	if !hasCommandOfType(pub.messages, "resume_agent") {
		t.Error("expected resume_agent command to be published")
	}
}

func TestPauseManager_ResumeAgent_WithPausedTask_TransitionsToAssigned(t *testing.T) {
	pm, _, agentPauser, taskStore, _ := setupPauseManager(t)

	projectID := shared.NewID()
	task := taskStore.addTask("paused")
	task.ProjectID = projectID
	taskStore.tasks[task.ID] = task

	a := newPausedAgent(projectID)
	a.CurrentTaskID = &task.ID
	agentPauser.agents = []agent.Agent{a}

	err := pm.ResumeAgent(context.Background(), a)
	if err != nil {
		t.Fatalf("ResumeAgent() error: %v", err)
	}

	got, _ := taskStore.GetByID(context.Background(), task.ID)
	if got.Status != "assigned" {
		t.Errorf("task status = %q, want assigned", got.Status)
	}
}

func TestPauseManager_ResumeProject_ResumesAllPausedAgents(t *testing.T) {
	pm, projects, agentPauser, _, _ := setupPauseManager(t)

	projectID := shared.NewID()
	projects.projects[projectID] = &project.Project{ID: projectID, Status: "paused"}

	a1 := newPausedAgent(projectID)
	a2 := newPausedAgent(projectID)
	active := newActiveAgent(projectID) // already active — should be skipped
	agentPauser.agents = []agent.Agent{a1, a2, active}

	err := pm.ResumeProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("ResumeProject() error: %v", err)
	}

	// Two paused agents should have been resumed
	activeCount := 0
	for _, upd := range agentPauser.statusUpdates {
		if upd.status == "active" {
			activeCount++
		}
	}
	if activeCount != 2 {
		t.Errorf("resumed %d agents, want 2", activeCount)
	}

	// Project status updated to active
	if !hasProjectStatusUpdate(projects.statusUpdates, projectID, "active") {
		t.Error("expected project status to be updated to active")
	}
}

// --- assertion helpers ---

func hasCommandOfType(messages []publishedEvent, cmdType string) bool {
	needle := `"type":"` + cmdType + `"`
	for _, msg := range messages {
		if strings.Contains(string(msg.body), needle) {
			return true
		}
	}
	return false
}

func hasProjectStatusUpdate(updates []projectStatusUpdate, id shared.ID, status string) bool {
	for _, u := range updates {
		if u.id == id && u.status == status {
			return true
		}
	}
	return false
}

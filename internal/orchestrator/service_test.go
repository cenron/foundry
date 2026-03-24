package orchestrator_test

import (
	"context"
	"sync"
	"testing"

	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/shared"
	"github.com/lib/pq"
)

type mockAgentFinder struct {
	agents      []orchestrator.AvailableAgent
	taskUpdates []agentTaskUpdate
}

type agentTaskUpdate struct {
	agentID shared.ID
	taskID  *shared.ID
}

func (m *mockAgentFinder) ListAvailableByProject(_ context.Context, _ shared.ID) ([]orchestrator.AvailableAgent, error) {
	return m.agents, nil
}

func (m *mockAgentFinder) UpdateCurrentTask(_ context.Context, agentID shared.ID, taskID *shared.ID) error {
	m.taskUpdates = append(m.taskUpdates, agentTaskUpdate{agentID, taskID})
	return nil
}

type mockUnblockedFinder struct {
	tasks []orchestrator.Task
}

func (m *mockUnblockedFinder) GetUnblockedTasks(_ context.Context, _ shared.ID) ([]orchestrator.Task, error) {
	return m.tasks, nil
}

type recordingPublisher struct {
	mu       sync.Mutex
	messages []publishedEvent
}

func (r *recordingPublisher) Publish(_ context.Context, exchange, routingKey string, body []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, publishedEvent{exchange, routingKey, body})
	return nil
}

func setupService(t *testing.T) (
	*orchestrator.Service,
	*mockTaskStateStore,
	*mockAgentFinder,
	*mockUnblockedFinder,
	*recordingPublisher,
) {
	t.Helper()

	store := newMockTaskStateStore()
	pub := &recordingPublisher{}
	sm := orchestrator.NewStateMachine(store, pub)

	unblocker := &mockUnblockedFinder{}
	dag := orchestrator.NewDAGResolver(unblocker)

	agents := &mockAgentFinder{}

	svc := orchestrator.NewService(store, dag, sm, agents, pub)
	return svc, store, agents, unblocker, pub
}

func TestService_HandleTaskCompleted_AssignsUnblocked(t *testing.T) {
	svc, store, agents, unblocker, pub := setupService(t)

	projectID := shared.NewID()

	// Task A is in_progress, about to complete
	taskA := store.addTask("in_progress")
	taskA.ProjectID = projectID
	store.tasks[taskA.ID] = taskA

	// Task B is pending and unblocked after A completes
	taskB := &orchestrator.Task{
		ID:           shared.NewID(),
		ProjectID:    projectID,
		Title:        "Task B",
		Status:       "pending",
		AssignedRole: "backend-developer",
		DependsOn:    pq.StringArray{taskA.ID.String()},
	}
	store.tasks[taskB.ID] = taskB
	unblocker.tasks = []orchestrator.Task{*taskB}

	// Available agent matching the role
	agentID := shared.NewID()
	agents.agents = []orchestrator.AvailableAgent{
		{ID: agentID, Role: "backend-developer"},
	}

	err := svc.HandleTaskCompleted(context.Background(), projectID, taskA.ID)
	if err != nil {
		t.Fatalf("HandleTaskCompleted() error: %v", err)
	}

	// Task A should be done
	gotA, _ := store.GetByID(context.Background(), taskA.ID)
	if gotA.Status != "done" {
		t.Errorf("Task A status = %q, want done", gotA.Status)
	}

	// Task B should be assigned
	gotB, _ := store.GetByID(context.Background(), taskB.ID)
	if gotB.Status != "assigned" {
		t.Errorf("Task B status = %q, want assigned", gotB.Status)
	}

	// Agent should have task B assigned
	if len(agents.taskUpdates) != 1 || *agents.taskUpdates[0].taskID != taskB.ID {
		t.Errorf("expected agent to be assigned task B")
	}

	// Should have published assignment command
	pub.mu.Lock()
	defer pub.mu.Unlock()
	hasCommand := false
	for _, msg := range pub.messages {
		if msg.exchange == "foundry.commands" {
			hasCommand = true
			break
		}
	}
	if !hasCommand {
		t.Error("expected assignment command to be published")
	}
}

func TestService_HandleTaskCompleted_NoAvailableAgent(t *testing.T) {
	svc, store, agents, unblocker, _ := setupService(t)

	projectID := shared.NewID()
	taskA := store.addTask("in_progress")
	taskA.ProjectID = projectID
	store.tasks[taskA.ID] = taskA

	taskB := &orchestrator.Task{
		ID:           shared.NewID(),
		ProjectID:    projectID,
		Title:        "Task B",
		Status:       "pending",
		AssignedRole: "frontend-developer",
		DependsOn:    pq.StringArray{},
	}
	store.tasks[taskB.ID] = taskB
	unblocker.tasks = []orchestrator.Task{*taskB}

	// No agents available
	agents.agents = nil

	err := svc.HandleTaskCompleted(context.Background(), projectID, taskA.ID)
	if err != nil {
		t.Fatalf("HandleTaskCompleted() error: %v", err)
	}

	// Task B should remain pending (no agent to assign)
	gotB, _ := store.GetByID(context.Background(), taskB.ID)
	if gotB.Status != "pending" {
		t.Errorf("Task B status = %q, want pending (no agent available)", gotB.Status)
	}
}

func TestService_HandleAgentUnhealthy(t *testing.T) {
	svc, store, _, _, _ := setupService(t)

	task := store.addTask("in_progress")

	err := svc.HandleAgentUnhealthy(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("HandleAgentUnhealthy() error: %v", err)
	}

	got, _ := store.GetByID(context.Background(), task.ID)
	if got.Status != "paused" {
		t.Errorf("status = %q, want paused", got.Status)
	}
}

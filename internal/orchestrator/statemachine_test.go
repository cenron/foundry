package orchestrator_test

import (
	"context"
	"sync"
	"testing"

	"github.com/cenron/foundry/internal/orchestrator"
	"github.com/cenron/foundry/internal/shared"
)

type mockTaskStateStore struct {
	mu     sync.Mutex
	tasks  map[shared.ID]*orchestrator.Task
	updates []statusUpdate
}

type statusUpdate struct {
	id     shared.ID
	status string
}

func newMockTaskStateStore() *mockTaskStateStore {
	return &mockTaskStateStore{
		tasks: make(map[shared.ID]*orchestrator.Task),
	}
}

func (m *mockTaskStateStore) addTask(status string) *orchestrator.Task {
	id := shared.NewID()
	t := &orchestrator.Task{
		ID:        id,
		ProjectID: shared.NewID(),
		Title:     "Test task",
		Status:    status,
	}
	m.tasks[id] = t
	return t
}

func (m *mockTaskStateStore) GetByID(_ context.Context, id shared.ID) (*orchestrator.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, &shared.NotFoundError{Resource: "task", ID: id.String()}
	}
	// Return a copy so the caller sees the current status
	cp := *t
	return &cp, nil
}

func (m *mockTaskStateStore) UpdateStatus(_ context.Context, id shared.ID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return &shared.NotFoundError{Resource: "task", ID: id.String()}
	}
	t.Status = status
	m.updates = append(m.updates, statusUpdate{id, status})
	return nil
}

type mockEventPublisher struct {
	mu       sync.Mutex
	messages []publishedEvent
}

type publishedEvent struct {
	exchange   string
	routingKey string
	body       []byte
}

func (m *mockEventPublisher) Publish(_ context.Context, exchange, routingKey string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, publishedEvent{exchange, routingKey, body})
	return nil
}

func TestStateMachine_ValidTransitions(t *testing.T) {
	tests := []struct {
		from string
		to   string
	}{
		{"pending", "assigned"},
		{"assigned", "in_progress"},
		{"in_progress", "paused"},
		{"in_progress", "review"},
		{"in_progress", "done"},
		{"paused", "assigned"},
		{"review", "done"},
		{"review", "in_progress"},
		{"assigned", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.from+"→"+tt.to, func(t *testing.T) {
			store := newMockTaskStateStore()
			pub := &mockEventPublisher{}
			sm := orchestrator.NewStateMachine(store, pub)

			task := store.addTask(tt.from)

			err := sm.Transition(context.Background(), task.ID, tt.to)
			if err != nil {
				t.Fatalf("Transition(%s→%s) error: %v", tt.from, tt.to, err)
			}

			// Verify status was persisted
			got, _ := store.GetByID(context.Background(), task.ID)
			if got.Status != tt.to {
				t.Errorf("status = %q, want %q", got.Status, tt.to)
			}

			// Verify event was published
			pub.mu.Lock()
			if len(pub.messages) != 1 {
				t.Errorf("expected 1 event, got %d", len(pub.messages))
			}
			pub.mu.Unlock()
		})
	}
}

func TestStateMachine_InvalidTransitions(t *testing.T) {
	tests := []struct {
		from string
		to   string
	}{
		{"pending", "in_progress"},
		{"pending", "done"},
		{"assigned", "done"},
		{"assigned", "review"},
		{"done", "pending"},
		{"done", "in_progress"},
		{"paused", "done"},
		{"review", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.from+"→"+tt.to, func(t *testing.T) {
			store := newMockTaskStateStore()
			pub := &mockEventPublisher{}
			sm := orchestrator.NewStateMachine(store, pub)

			task := store.addTask(tt.from)

			err := sm.Transition(context.Background(), task.ID, tt.to)
			if err == nil {
				t.Fatalf("expected error for invalid transition %s→%s", tt.from, tt.to)
			}

			// Verify status was NOT changed
			got, _ := store.GetByID(context.Background(), task.ID)
			if got.Status != tt.from {
				t.Errorf("status should remain %q, got %q", tt.from, got.Status)
			}
		})
	}
}

func TestStateMachine_TaskNotFound(t *testing.T) {
	store := newMockTaskStateStore()
	pub := &mockEventPublisher{}
	sm := orchestrator.NewStateMachine(store, pub)

	err := sm.Transition(context.Background(), shared.NewID(), "assigned")
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
}

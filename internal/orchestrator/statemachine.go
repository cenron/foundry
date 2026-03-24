package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cenron/foundry/internal/shared"
)

// validTransitions defines the allowed task state transitions.
var validTransitions = map[string][]string{
	"pending":     {"assigned"},
	"assigned":    {"in_progress", "pending"},
	"in_progress": {"paused", "review", "done"},
	"paused":      {"assigned"},
	"review":      {"done", "in_progress"},
}

// EventPublisher publishes events to the message broker.
type EventPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

// TaskStateStore persists task state changes.
type TaskStateStore interface {
	UpdateStatus(ctx context.Context, id shared.ID, status string) error
	GetByID(ctx context.Context, id shared.ID) (*Task, error)
}

type StateMachine struct {
	store     TaskStateStore
	publisher EventPublisher
}

func NewStateMachine(store TaskStateStore, publisher EventPublisher) *StateMachine {
	return &StateMachine{store: store, publisher: publisher}
}

func (sm *StateMachine) Transition(ctx context.Context, taskID shared.ID, newStatus string) error {
	task, err := sm.store.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("loading task: %w", err)
	}

	if !isValidTransition(task.Status, newStatus) {
		return &shared.ConflictError{
			Resource: "task",
			Message:  fmt.Sprintf("invalid transition from %q to %q", task.Status, newStatus),
		}
	}

	if err := sm.store.UpdateStatus(ctx, taskID, newStatus); err != nil {
		return fmt.Errorf("persisting transition: %w", err)
	}

	sm.publishTransitionEvent(ctx, task, newStatus)
	return nil
}

func isValidTransition(from, to string) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

func (sm *StateMachine) publishTransitionEvent(ctx context.Context, task *Task, newStatus string) {
	payload, _ := json.Marshal(map[string]string{
		"task_id":    task.ID.String(),
		"project_id": task.ProjectID.String(),
		"from":       task.Status,
		"to":         newStatus,
		"title":      task.Title,
	})

	routingKey := fmt.Sprintf("events.%s.task_%s", task.ProjectID, newStatus)
	if err := sm.publisher.Publish(ctx, "foundry.events", routingKey, payload); err != nil {
		log.Printf("state machine: publishing transition event: %v", err)
	}
}

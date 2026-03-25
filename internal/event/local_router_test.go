package event_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/cenron/foundry/internal/event"
)

// localMockBroadcaster is a test double for event.Broadcaster.
// Uses a different name from mockBroadcaster in router_test.go.
type localMockBroadcaster struct {
	mu       sync.Mutex
	messages [][]byte
}

func (m *localMockBroadcaster) Broadcast(msg []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *localMockBroadcaster) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}

func (m *localMockBroadcaster) last() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.messages) == 0 {
		return nil
	}
	return m.messages[len(m.messages)-1]
}

// --- StreamAgentOutput ---

func TestLocalRouter_StreamAgentOutput_ForwardsLines(t *testing.T) {
	hub := &localMockBroadcaster{}
	lr := event.NewLocalRouter(nil, hub, nil)

	projectID := "00000000-0000-0000-0000-000000000001"
	agentID := "00000000-0000-0000-0000-000000000002"

	lines := `{"type":"assistant","content":"hello"}
{"type":"assistant","content":"world"}
`
	reader := strings.NewReader(lines)

	ctx := context.Background()
	lr.StreamAgentOutput(ctx, projectID, agentID, reader)

	if hub.count() != 2 {
		t.Errorf("Broadcast called %d times, want 2", hub.count())
	}
}

func TestLocalRouter_StreamAgentOutput_MalformedJSONSkipped(t *testing.T) {
	hub := &localMockBroadcaster{}
	lr := event.NewLocalRouter(nil, hub, nil)

	projectID := "00000000-0000-0000-0000-000000000001"
	agentID := "00000000-0000-0000-0000-000000000002"

	lines := "not json at all\n{\"type\":\"assistant\"}\n"
	reader := strings.NewReader(lines)

	lr.StreamAgentOutput(context.Background(), projectID, agentID, reader)

	// Only the valid JSON line is broadcast.
	if hub.count() != 1 {
		t.Errorf("Broadcast called %d times, want 1 (malformed line skipped)", hub.count())
	}
}

func TestLocalRouter_StreamAgentOutput_ContextCancel(t *testing.T) {
	hub := &localMockBroadcaster{}
	lr := event.NewLocalRouter(nil, hub, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Even with data, context is already cancelled so we should exit early.
	reader := strings.NewReader(`{"type":"assistant"}` + "\n")
	lr.StreamAgentOutput(ctx, "proj-1", "agent-1", reader)

	// With an already-cancelled context the select in the loop may or may not
	// process the single line depending on scheduling. We just verify it doesn't hang.
}

// --- ForwardLogLine ---

func TestLocalRouter_ForwardLogLine_EmptyLineSkipped(t *testing.T) {
	hub := &localMockBroadcaster{}
	lr := event.NewLocalRouter(nil, hub, nil)

	lr.ForwardLogLine(context.Background(), "proj-1", "agent-1", "")
	lr.ForwardLogLine(context.Background(), "proj-1", "agent-1", "   ")

	if hub.count() != 0 {
		t.Errorf("Broadcast called %d times for empty lines, want 0", hub.count())
	}
}

func TestLocalRouter_ForwardLogLine_NoTypeDefaultsToAgentOutput(t *testing.T) {
	hub := &localMockBroadcaster{}
	lr := event.NewLocalRouter(nil, hub, nil)

	// Valid JSON but no "type" field.
	lr.ForwardLogLine(context.Background(), "proj-1", "agent-1", `{"content":"something"}`)

	if hub.count() != 1 {
		t.Fatalf("Broadcast called %d times, want 1", hub.count())
	}

	last := string(hub.last())
	if !strings.Contains(last, `"agent.output"`) {
		t.Errorf("expected default type agent.output in broadcast, got: %s", last)
	}
}

func TestLocalRouter_ForwardLogLine_TypeFieldExtracted(t *testing.T) {
	hub := &localMockBroadcaster{}
	lr := event.NewLocalRouter(nil, hub, nil)

	lr.ForwardLogLine(context.Background(), "proj-1", "agent-1", `{"type":"task.completed","task_id":"t-1"}`)

	if hub.count() != 1 {
		t.Fatalf("Broadcast called %d times, want 1", hub.count())
	}

	last := string(hub.last())
	if !strings.Contains(last, `"task.completed"`) {
		t.Errorf("expected type task.completed in broadcast, got: %s", last)
	}
}

func TestLocalRouter_ForwardLogLine_InvalidJSONSkipped(t *testing.T) {
	hub := &localMockBroadcaster{}
	lr := event.NewLocalRouter(nil, hub, nil)

	lr.ForwardLogLine(context.Background(), "proj-1", "agent-1", "definitely not json")

	if hub.count() != 0 {
		t.Errorf("Broadcast called %d times for invalid JSON, want 0", hub.count())
	}
}

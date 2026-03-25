package e2e_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cenron/foundry/internal/event"
	"github.com/cenron/foundry/internal/po"
	"github.com/cenron/foundry/internal/runtime"
)

// TestLocalSmoke_SetupWorkspace verifies that LocalRuntime.Setup creates
// the expected directory tree.
func TestLocalSmoke_SetupWorkspace(t *testing.T) {
	r := runtime.NewLocalRuntime(4)
	base := t.TempDir()

	if err := r.Setup(context.Background(), runtime.SetupOpts{
		ProjectID: "smoke-proj-1",
		WorkDir:   base,
	}); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	dirs := []string{
		filepath.Join(base, "projects", "smoke-proj-1", "workspace"),
		filepath.Join(base, "projects", "smoke-proj-1", "worktrees"),
		filepath.Join(base, "projects", "smoke-proj-1", "shared", "status"),
		filepath.Join(base, "projects", "smoke-proj-1", "shared", "messages"),
		filepath.Join(base, "projects", "smoke-proj-1", "state"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %q to exist", dir)
		}
	}
}

// TestLocalSmoke_WatchEvents_StatusFile verifies that fsnotify fires when a
// file is written into the shared/status directory.
func TestLocalSmoke_WatchEvents_StatusFile(t *testing.T) {
	r := runtime.NewLocalRuntime(4)
	base := t.TempDir()

	// Create the directories manually for this test.
	statusDir := filepath.Join(base, "projects", "proj-watch", "shared", "status")
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := r.WatchEvents(ctx, "proj-watch")
	if err != nil {
		t.Fatalf("WatchEvents() error: %v", err)
	}
	defer func() { _ = r.Cleanup(context.Background(), "proj-watch") }()

	if err := r.AddWatchPath("proj-watch", statusDir); err != nil {
		t.Fatalf("AddWatchPath() error: %v", err)
	}

	// Write a JSON status file.
	payload := `{"status":"done","agent_id":"agent-1"}`
	statusFile := filepath.Join(statusDir, "agent-1.json")
	if err := os.WriteFile(statusFile, []byte(payload), 0644); err != nil {
		t.Fatalf("writing status file: %v", err)
	}

	select {
	case evt := <-ch:
		if evt.Type != "agent.status" {
			t.Errorf("Type = %q, want %q", evt.Type, "agent.status")
		}
		if evt.ProjectID != "proj-watch" {
			t.Errorf("ProjectID = %q, want %q", evt.ProjectID, "proj-watch")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for fsnotify event")
	}
}

// TestLocalSmoke_WatchEvents_MessageFile verifies that a message file written
// to shared/messages produces an agent.message event.
func TestLocalSmoke_WatchEvents_MessageFile(t *testing.T) {
	r := runtime.NewLocalRuntime(4)
	base := t.TempDir()

	messagesDir := filepath.Join(base, "shared", "messages")
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := r.WatchEvents(ctx, "proj-messages")
	if err != nil {
		t.Fatalf("WatchEvents() error: %v", err)
	}
	defer func() { _ = r.Cleanup(context.Background(), "proj-messages") }()

	if err := r.AddWatchPath("proj-messages", messagesDir); err != nil {
		t.Fatalf("AddWatchPath() error: %v", err)
	}

	// Write a JSON message file.
	msgPayload := `{"text":"task complete","from":"backend-developer"}`
	if err := os.WriteFile(filepath.Join(messagesDir, "msg-1.json"), []byte(msgPayload), 0644); err != nil {
		t.Fatalf("writing message file: %v", err)
	}

	select {
	case evt := <-ch:
		if evt.Type != "agent.message" {
			t.Errorf("Type = %q, want %q", evt.Type, "agent.message")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for message event")
	}
}

// TestLocalSmoke_LocalRouter_Forwarding verifies that LocalRouter.ForwardLogLine
// broadcasts stream-json output to the hub.
func TestLocalSmoke_LocalRouter_Forwarding(t *testing.T) {
	hub := &smokeHub{}
	lr := event.NewLocalRouter(nil, hub, nil)

	projectID := "00000000-0000-0000-0000-000000000001"
	agentID := "00000000-0000-0000-0000-000000000002"

	lines := []string{
		`{"type":"assistant","content":"Starting implementation"}`,
		`{"type":"tool_use","name":"bash","input":{"command":"go build ./..."}}`,
		`{"type":"result","content":"Build successful"}`,
	}

	for _, line := range lines {
		lr.ForwardLogLine(context.Background(), projectID, agentID, line)
	}

	if hub.count() != 3 {
		t.Errorf("expected 3 broadcasts, got %d", hub.count())
	}
}

// TestLocalSmoke_LocalPOCommandBuilder verifies that NewLocalSessionManager
// produces commands without --bare or --max-budget-usd but with --verbose.
func TestLocalSmoke_LocalPOCommandBuilder(t *testing.T) {
	m := po.NewLocalSessionManager(t.TempDir(), "latest")

	sessionTypes := []struct {
		name    string
		trigger string
	}{
		{po.SessionTypePlanning, "user"},
		{po.SessionTypeEstimation, "system"},
		{po.SessionTypeReview, "system"},
		{po.SessionTypeExecutionChat, "user"},
		{po.SessionTypeEscalation, "system"},
		{po.SessionTypePhaseTransition, "system"},
	}

	for _, st := range sessionTypes {
		t.Run(st.name, func(t *testing.T) {
			opts := po.POSessionOpts{
				Type:    st.name,
				Project: "smoke-proj",
				Trigger: st.trigger,
				Message: "smoke test",
			}

			args := m.BuildCommand(context.Background(), opts).Args

			// Must have --verbose.
			found := false
			for _, a := range args {
				if a == "--verbose" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("args missing --verbose: %v", args)
			}

			// Must NOT have --bare or --max-budget-usd.
			for _, a := range args {
				if a == "--bare" {
					t.Errorf("args unexpectedly contains --bare: %v", args)
				}
				if a == "--max-budget-usd" {
					t.Errorf("args unexpectedly contains --max-budget-usd: %v", args)
				}
			}
		})
	}
}

// TestLocalSmoke_StreamAgentOutput_JSON verifies that stream-json lines are
// forwarded and non-JSON is silently dropped.
func TestLocalSmoke_StreamAgentOutput_JSON(t *testing.T) {
	hub := &smokeHub{}
	lr := event.NewLocalRouter(nil, hub, nil)

	input := strings.Join([]string{
		`{"type":"assistant","content":"hello"}`,
		`not json at all`,
		`{"type":"result","content":"done"}`,
	}, "\n") + "\n"

	reader := strings.NewReader(input)
	lr.StreamAgentOutput(context.Background(), "proj-1", "agent-1", reader)

	// Only 2 valid JSON lines should be broadcast.
	if hub.count() != 2 {
		t.Errorf("expected 2 broadcasts, got %d", hub.count())
	}
}

// TestLocalSmoke_Cleanup verifies that Cleanup releases the watcher without error.
func TestLocalSmoke_Cleanup(t *testing.T) {
	r := runtime.NewLocalRuntime(4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := r.WatchEvents(ctx, "proj-cleanup"); err != nil {
		t.Fatalf("WatchEvents() error: %v", err)
	}

	if err := r.Cleanup(context.Background(), "proj-cleanup"); err != nil {
		t.Errorf("Cleanup() error: %v", err)
	}
}

// TestLocalSmoke_ParseEventFile verifies parsing a valid JSON event file.
func TestLocalSmoke_ParseEventFile(t *testing.T) {
	f, err := os.CreateTemp("", "smoke-event-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	data := map[string]interface{}{
		"type":      "agent.status",
		"agent_id":  "agent-1",
		"status":    "running",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	b, _ := json.Marshal(data)
	_, _ = f.Write(b)
	_ = f.Close()

	// Read back the file and verify it's valid JSON.
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("parsing JSON: %v", err)
	}

	if result["type"] != "agent.status" {
		t.Errorf("type = %v, want agent.status", result["type"])
	}
}

// smokeHub is a minimal Broadcaster for smoke tests.
type smokeHub struct {
	mu       sync.Mutex
	messages [][]byte
}

func (h *smokeHub) Broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = append(h.messages, msg)
}

func (h *smokeHub) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.messages)
}

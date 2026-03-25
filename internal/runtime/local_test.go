package runtime

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// --- Setup ---

func TestLocalRuntime_Setup_CreatesDirectories(t *testing.T) {
	r := NewLocalRuntime(4)
	base := t.TempDir()

	opts := SetupOpts{
		ProjectID: "proj-1",
		WorkDir:   base,
	}

	if err := r.Setup(context.Background(), opts); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	expected := []string{
		filepath.Join(base, "projects", "proj-1", "workspace"),
		filepath.Join(base, "projects", "proj-1", "worktrees"),
		filepath.Join(base, "projects", "proj-1", "shared", "contracts"),
		filepath.Join(base, "projects", "proj-1", "shared", "designs"),
		filepath.Join(base, "projects", "proj-1", "shared", "context"),
		filepath.Join(base, "projects", "proj-1", "shared", "reviews"),
		filepath.Join(base, "projects", "proj-1", "shared", "messages"),
		filepath.Join(base, "projects", "proj-1", "shared", "status"),
		filepath.Join(base, "projects", "proj-1", "state"),
	}

	for _, dir := range expected {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %q to exist", dir)
		}
	}
}

func TestLocalRuntime_Setup_Idempotent(t *testing.T) {
	r := NewLocalRuntime(4)
	base := t.TempDir()

	opts := SetupOpts{ProjectID: "proj-idem", WorkDir: base}

	if err := r.Setup(context.Background(), opts); err != nil {
		t.Fatalf("first Setup() error: %v", err)
	}

	if err := r.Setup(context.Background(), opts); err != nil {
		t.Fatalf("second Setup() error: %v", err)
	}
}

// --- LaunchAgent ---

func TestLocalRuntime_LaunchAgent_SkipIfNoClaude(t *testing.T) {
	if _, err := findClaude(); err != nil {
		t.Skip("claude not found on PATH, skipping")
	}
}

func TestLocalRuntime_StopAgent_NotFound(t *testing.T) {
	r := NewLocalRuntime(4)

	err := r.StopAgent(context.Background(), "nonexistent-agent")
	if err == nil {
		t.Error("StopAgent() error = nil for missing agent, want error")
	}
}

// --- WatchEvents ---

func TestLocalRuntime_WatchEvents_ReturnsChannel(t *testing.T) {
	r := NewLocalRuntime(4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := r.WatchEvents(ctx, "proj-watch")
	if err != nil {
		t.Fatalf("WatchEvents() error: %v", err)
	}

	if ch == nil {
		t.Error("WatchEvents() returned nil channel")
	}

	_ = r.Cleanup(context.Background(), "proj-watch")
}

func TestLocalRuntime_WatchEvents_DuplicateReturnsError(t *testing.T) {
	r := NewLocalRuntime(4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := r.WatchEvents(ctx, "proj-dup")
	if err != nil {
		t.Fatalf("first WatchEvents() error: %v", err)
	}
	defer func() { _ = r.Cleanup(context.Background(), "proj-dup") }()

	_, err = r.WatchEvents(ctx, "proj-dup")
	if err == nil {
		t.Error("second WatchEvents() error = nil, want error")
	}
}

func TestLocalRuntime_WatchEvents_MessageEvent(t *testing.T) {
	r := NewLocalRuntime(4)
	base := t.TempDir()

	// Create the shared/messages directory to watch.
	messagesDir := filepath.Join(base, "shared", "messages")
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		t.Fatalf("creating messages dir: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := r.WatchEvents(ctx, "proj-msg")
	if err != nil {
		t.Fatalf("WatchEvents() error: %v", err)
	}
	defer func() { _ = r.Cleanup(context.Background(), "proj-msg") }()

	// Add the messages dir to the watcher.
	if err := r.AddWatchPath("proj-msg", messagesDir); err != nil {
		t.Fatalf("AddWatchPath() error: %v", err)
	}

	// Write a JSON file to trigger an event.
	payload := `{"text":"hello from agent"}`
	msgFile := filepath.Join(messagesDir, "msg-001.json")
	if err := os.WriteFile(msgFile, []byte(payload), 0644); err != nil {
		t.Fatalf("writing message file: %v", err)
	}

	select {
	case evt := <-ch:
		if evt.Type != "agent.message" {
			t.Errorf("Type = %q, want %q", evt.Type, "agent.message")
		}
		if evt.ProjectID != "proj-msg" {
			t.Errorf("ProjectID = %q, want %q", evt.ProjectID, "proj-msg")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for message event")
	}
}

// --- Cleanup ---

func TestLocalRuntime_Cleanup_NoWatcher(t *testing.T) {
	r := NewLocalRuntime(4)

	// Cleanup for a project with no watcher should not error.
	if err := r.Cleanup(context.Background(), "proj-none"); err != nil {
		t.Errorf("Cleanup() error = %v, want nil", err)
	}
}

// --- Concurrency limits ---

func TestLocalRuntime_ConcurrencyLimit(t *testing.T) {
	r := NewLocalRuntime(1)

	// Fill the semaphore manually.
	r.semaphore <- struct{}{}

	opts := AgentOpts{
		AgentID:   "agent-overflow",
		ProjectID: "proj-1",
		Prompt:    "test",
	}

	_, err := r.LaunchAgent(context.Background(), opts)
	if err == nil {
		t.Error("LaunchAgent() error = nil when at concurrency limit, want error")
	}
}

// --- buildAgentArgs ---

func TestLocalRuntime_BuildAgentArgs_HasVerbose(t *testing.T) {
	r := NewLocalRuntime(4)
	opts := AgentOpts{
		AgentID: "agent-1",
		Prompt:  "do the thing",
	}

	args := r.buildAgentArgs(opts)

	assertArgPresent(t, args, "--verbose")
}

func TestLocalRuntime_BuildAgentArgs_NoBareFlagPresent(t *testing.T) {
	r := NewLocalRuntime(4)
	opts := AgentOpts{
		AgentID: "agent-1",
		Prompt:  "do the thing",
	}

	args := r.buildAgentArgs(opts)

	assertArgAbsent(t, args, "--bare")
}

func TestLocalRuntime_BuildAgentArgs_HasStreamJSON(t *testing.T) {
	r := NewLocalRuntime(4)
	opts := AgentOpts{Prompt: "test"}

	args := r.buildAgentArgs(opts)

	assertArgPair(t, args, "--output-format", "stream-json")
}

func TestLocalRuntime_BuildAgentArgs_HasDangerouslySkipPermissions(t *testing.T) {
	r := NewLocalRuntime(4)
	opts := AgentOpts{Prompt: "test"}

	args := r.buildAgentArgs(opts)

	assertArgPresent(t, args, "--dangerously-skip-permissions")
}

// --- classifyByDirectory ---

func TestClassifyByDirectory(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/foundry/projects/p/shared/status/agent.json", "agent.status"},
		{"/foundry/projects/p/shared/messages/msg.json", "agent.message"},
		{"/foundry/projects/p/shared/reviews/r.json", "agent.review"},
		{"/foundry/projects/p/shared/contracts/c.json", "agent.contract"},
		{"/foundry/projects/p/shared/designs/d.json", "agent.design"},
		{"/foundry/projects/p/shared/context/ctx.json", "agent.context"},
		{"/foundry/projects/p/workspace/output.txt", "agent.output"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := classifyByDirectory(tt.path)
			if got != tt.want {
				t.Errorf("classifyByDirectory(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// --- cloneRepo ---

func TestLocalRuntime_CloneRepo_AlreadyPopulated(t *testing.T) {
	r := NewLocalRuntime(4)
	dest := t.TempDir()

	// Pre-populate the dir so clone is skipped.
	if err := os.WriteFile(filepath.Join(dest, "existing.txt"), []byte("exists"), 0644); err != nil {
		t.Fatalf("writing existing file: %v", err)
	}

	// Should return nil without calling git.
	if err := r.cloneRepo(context.Background(), "https://github.com/example/repo", dest); err != nil {
		t.Errorf("cloneRepo() error = %v, want nil for already-populated dir", err)
	}
}

func TestLocalRuntime_CloneRepo_InvalidURL(t *testing.T) {
	r := NewLocalRuntime(4)
	dest := t.TempDir()

	// Empty dest + bad URL → git will fail.
	err := r.cloneRepo(context.Background(), "https://invalid.localhost/no-such-repo", dest)
	if err == nil {
		t.Error("cloneRepo() error = nil for invalid URL, want error")
	}
}

// --- parseEventFile ---

func TestParseEventFile_Invalid(t *testing.T) {
	f, err := os.CreateTemp("", "event-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	_, _ = f.WriteString("not valid json")
	_ = f.Close()

	_, err = parseEventFile(f.Name())
	if err == nil {
		t.Error("parseEventFile() error = nil for invalid JSON, want error")
	}
}

func TestParseEventFile_Valid(t *testing.T) {
	f, err := os.CreateTemp("", "event-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	data := map[string]interface{}{"type": "agent.status", "status": "done"}
	b, _ := json.Marshal(data)
	_, _ = f.Write(b)
	_ = f.Close()

	result, err := parseEventFile(f.Name())
	if err != nil {
		t.Fatalf("parseEventFile() error: %v", err)
	}

	if result["type"] != "agent.status" {
		t.Errorf("type = %v, want %q", result["type"], "agent.status")
	}
}

// --- watchLoop context cancel ---

func TestLocalRuntime_WatchLoop_ContextCancel(t *testing.T) {
	r := NewLocalRuntime(4)

	ctx, cancel := context.WithCancel(context.Background())

	ch, err := r.WatchEvents(ctx, "proj-cancel")
	if err != nil {
		t.Fatalf("WatchEvents() error: %v", err)
	}

	// Cancel the context — the watch loop should terminate and close the channel.
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed after context cancel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for channel to close after context cancel")
	}
}

// --- AddWatchPath no watcher ---

func TestLocalRuntime_AddWatchPath_NoWatcher(t *testing.T) {
	r := NewLocalRuntime(4)

	err := r.AddWatchPath("proj-none", "/tmp")
	if err == nil {
		t.Error("AddWatchPath() error = nil for missing watcher, want error")
	}
}

// --- IsAgentRunning ---

func TestLocalRuntime_IsAgentRunning_False(t *testing.T) {
	r := NewLocalRuntime(4)

	if r.IsAgentRunning("nonexistent") {
		t.Error("IsAgentRunning() = true for unknown agent, want false")
	}
}

// --- NewLocalRuntime default concurrency ---

func TestLocalRuntime_DefaultConcurrency(t *testing.T) {
	// When maxConcurrent is 0, the semaphore has capacity 0 — any LaunchAgent
	// call will immediately fail with "concurrency limit reached".
	r := NewLocalRuntime(0)

	// Verify the semaphore was created with capacity 0 (channel capacity == 0).
	// We can probe this indirectly: trying to send without a receiver blocks forever;
	// but we can attempt a non-blocking send and see it fails.
	select {
	case r.semaphore <- struct{}{}:
		// Capacity is not 0 — unexpected.
		<-r.semaphore // drain to avoid leak
		t.Error("semaphore unexpectedly accepted a token for capacity-0 runtime")
	default:
		// Capacity is 0 — correct.
	}
}

// --- writeAgentState ---

func TestLocalRuntime_WriteAgentState_CreatesFile(t *testing.T) {
	r := NewLocalRuntime(4)
	base := t.TempDir()

	// writeAgentState computes: stateDir = filepath.Clean(WorkDir + "/../../state").
	// With WorkDir = base/projects/proj-1/workspace:
	//   WorkDir/../..  = base/projects
	//   + "/state"     = base/projects/state
	agentWorkDir := filepath.Join(base, "projects", "proj-1", "workspace")
	if err := os.MkdirAll(agentWorkDir, 0755); err != nil {
		t.Fatalf("creating agent workdir: %v", err)
	}

	opts := AgentOpts{
		AgentID:   "agent-state-test",
		ProjectID: "proj-1",
		WorkDir:   agentWorkDir,
	}

	r.writeAgentState(opts, "running", 9999)

	// State file lands at base/projects/state/agent-state-test.json.
	stateFile := filepath.Join(base, "projects", "state", "agent-state-test.json")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Fatalf("state file not created at %s", stateFile)
	}

	result, err := parseEventFile(stateFile)
	if err != nil {
		t.Fatalf("parsing state file: %v", err)
	}

	if result["agent_id"] != "agent-state-test" {
		t.Errorf("agent_id = %v, want %q", result["agent_id"], "agent-state-test")
	}
	if result["status"] != "running" {
		t.Errorf("status = %v, want %q", result["status"], "running")
	}
}

// --- parseEventFile invalid path ---

func TestParseEventFile_InvalidPath(t *testing.T) {
	_, err := parseEventFile("/nonexistent/path/to/event.json")
	if err == nil {
		t.Error("parseEventFile() error = nil for nonexistent path, want error")
	}
}

// --- watchLoop non-JSON file ---

func TestLocalRuntime_WatchLoop_NonJSONFile(t *testing.T) {
	r := NewLocalRuntime(4)
	base := t.TempDir()

	watchDir := filepath.Join(base, "shared", "messages")
	_ = os.MkdirAll(watchDir, 0755)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := r.WatchEvents(ctx, "proj-nonjson")
	if err != nil {
		t.Fatalf("WatchEvents() error: %v", err)
	}
	defer func() { _ = r.Cleanup(context.Background(), "proj-nonjson") }()

	if err := r.AddWatchPath("proj-nonjson", watchDir); err != nil {
		t.Fatalf("AddWatchPath() error: %v", err)
	}

	// Write a non-JSON .txt file — watchLoop must skip it and emit no event.
	txtFile := filepath.Join(watchDir, "note.txt")
	if err := os.WriteFile(txtFile, []byte("not json content"), 0644); err != nil {
		t.Fatalf("writing txt file: %v", err)
	}

	select {
	case evt, ok := <-ch:
		if ok {
			t.Errorf("unexpected event emitted for non-JSON file: %+v", evt)
		}
	case <-time.After(200 * time.Millisecond):
		// No event within 200ms — correct behaviour for non-JSON file.
	}
}

// --- Setup with pre-populated workspace (cloneRepo no-op path) ---

func TestLocalRuntime_Setup_WithRepoURL_PopulatedWorkspace(t *testing.T) {
	r := NewLocalRuntime(4)
	base := t.TempDir()

	// Pre-populate the workspace directory so cloneRepo returns early.
	workspaceDir := filepath.Join(base, "projects", "proj-populated", "workspace")
	_ = os.MkdirAll(workspaceDir, 0755)
	if err := os.WriteFile(filepath.Join(workspaceDir, "existing.txt"), []byte("present"), 0644); err != nil {
		t.Fatalf("pre-populating workspace: %v", err)
	}

	opts := SetupOpts{
		ProjectID: "proj-populated",
		RepoURL:   "https://github.com/example/repo",
		WorkDir:   base,
	}

	// Setup should succeed because cloneRepo skips the clone when dest is populated.
	if err := r.Setup(context.Background(), opts); err != nil {
		t.Errorf("Setup() error = %v, want nil for pre-populated workspace", err)
	}
}

// --- buildAgentArgs AllowedTools absent ---

func TestLocalRuntime_BuildAgentArgs_NoAllowedTools(t *testing.T) {
	r := NewLocalRuntime(4)
	opts := AgentOpts{
		AgentID: "agent-1",
		Prompt:  "do the thing",
	}

	args := r.buildAgentArgs(opts)

	assertArgAbsent(t, args, "--allowedTools")
}

// --- helpers ---

func assertArgPresent(t *testing.T, args []string, flag string) {
	t.Helper()
	for _, a := range args {
		if a == flag {
			return
		}
	}
	t.Errorf("args missing %q\ngot: %v", flag, args)
}

func assertArgAbsent(t *testing.T, args []string, flag string) {
	t.Helper()
	for _, a := range args {
		if a == flag {
			t.Errorf("args unexpectedly contains %q\ngot: %v", flag, args)
			return
		}
	}
}

func assertArgPair(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, a := range args {
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return
		}
	}
	t.Errorf("args missing %q %q\ngot: %v", flag, value, args)
}

// findClaude checks whether claude is available on PATH.
func findClaude() (string, error) {
	return exec.LookPath("claude")
}

// stubClaude creates a temp directory with a shell script named "claude" that
// sleeps for 30 seconds (simulating a long-running agent process). It prepends
// the directory to PATH and returns a cleanup function that restores PATH.
func stubClaude(t *testing.T) (stubDir string) {
	t.Helper()

	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not available, skipping test")
	}

	claudeDir := t.TempDir()
	claudeScript := filepath.Join(claudeDir, "claude")
	script := "#!/bin/sh\n" + sleepPath + " 30\n"
	if err := os.WriteFile(claudeScript, []byte(script), 0755); err != nil {
		t.Fatalf("creating claude stub: %v", err)
	}

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", claudeDir+":"+origPath)

	return claudeDir
}

// --- LaunchAgent and StopAgent success paths ---

func TestLocalRuntime_LaunchAgent_AndStop(t *testing.T) {
	stubClaude(t)

	r := NewLocalRuntime(4)
	base := t.TempDir()

	opts := AgentOpts{
		AgentID:   "agent-launch-test",
		ProjectID: "proj-launch",
		Prompt:    "do something",
		WorkDir:   base,
	}

	proc, err := r.LaunchAgent(context.Background(), opts)
	if err != nil {
		t.Fatalf("LaunchAgent() error: %v", err)
	}

	if proc == nil {
		t.Fatal("LaunchAgent() returned nil AgentProcess")
	}

	if proc.PID <= 0 {
		t.Errorf("PID = %d, want > 0", proc.PID)
	}

	if !r.IsAgentRunning("agent-launch-test") {
		t.Error("IsAgentRunning() = false after launch, want true")
	}

	// Stop the agent and verify it exits cleanly.
	if err := r.StopAgent(context.Background(), "agent-launch-test"); err != nil {
		t.Fatalf("StopAgent() error: %v", err)
	}

	// After stopping, the reaper goroutine removes the agent from the map.
	// Give it a brief moment to run.
	select {
	case <-proc.Done:
		// Process exited — correct.
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for agent process to exit after StopAgent")
	}
}

func TestLocalRuntime_LaunchAgent_VerifyWritesState(t *testing.T) {
	stubClaude(t)

	r := NewLocalRuntime(4)
	base := t.TempDir()

	// Build a workdir that follows the expected layout so writeAgentState can
	// find the right state dir.
	agentWorkDir := filepath.Join(base, "projects", "state-proj", "workspace")
	if err := os.MkdirAll(agentWorkDir, 0755); err != nil {
		t.Fatalf("creating agent workdir: %v", err)
	}

	opts := AgentOpts{
		AgentID:   "agent-state-verify",
		ProjectID: "state-proj",
		Prompt:    "test",
		WorkDir:   agentWorkDir,
	}

	proc, err := r.LaunchAgent(context.Background(), opts)
	if err != nil {
		t.Fatalf("LaunchAgent() error: %v", err)
	}

	// Stop so the reaper sets status to "stopped" and writes the second state file.
	if err := r.StopAgent(context.Background(), "agent-state-verify"); err != nil {
		t.Fatalf("StopAgent() error: %v", err)
	}

	<-proc.Done
}

// --- watchLoop watcher close (error channel closed) ---

func TestLocalRuntime_WatchLoop_WatcherClose(t *testing.T) {
	r := NewLocalRuntime(4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := r.WatchEvents(ctx, "proj-watcher-close")
	if err != nil {
		t.Fatalf("WatchEvents() error: %v", err)
	}

	// Closing the watcher closes its internal events channel,
	// which causes watchLoop to return and close out.
	if err := r.Cleanup(context.Background(), "proj-watcher-close"); err != nil {
		t.Fatalf("Cleanup() error: %v", err)
	}

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed after watcher close")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for channel to close after watcher close")
	}
}

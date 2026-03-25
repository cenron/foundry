package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// agentEntry tracks a running agent process and its cancellation.
type agentEntry struct {
	cancel context.CancelFunc
	done   chan struct{}
	pid    int
}

// LocalRuntime implements Runtime using local claude CLI processes.
type LocalRuntime struct {
	mu            sync.Mutex
	agents        map[string]*agentEntry
	watchers      map[string]*fsnotify.Watcher
	semaphore     chan struct{}
	maxConcurrent int
}

// NewLocalRuntime creates a LocalRuntime with bounded concurrency.
func NewLocalRuntime(maxConcurrent int) *LocalRuntime {
	return &LocalRuntime{
		agents:        make(map[string]*agentEntry),
		watchers:      make(map[string]*fsnotify.Watcher),
		semaphore:     make(chan struct{}, maxConcurrent),
		maxConcurrent: maxConcurrent,
	}
}

// Setup creates the standard workspace directory layout for a project.
func (r *LocalRuntime) Setup(ctx context.Context, opts SetupOpts) error {
	base := filepath.Join(opts.WorkDir, "projects", opts.ProjectID)

	dirs := []string{
		filepath.Join(base, "workspace"),
		filepath.Join(base, "worktrees"),
		filepath.Join(base, "shared", "contracts"),
		filepath.Join(base, "shared", "designs"),
		filepath.Join(base, "shared", "context"),
		filepath.Join(base, "shared", "reviews"),
		filepath.Join(base, "shared", "messages"),
		filepath.Join(base, "shared", "status"),
		filepath.Join(base, "state"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating workspace dir %s: %w", dir, err)
		}
	}

	if opts.RepoURL != "" {
		dest := filepath.Join(base, "workspace")
		if err := r.cloneRepo(ctx, opts.RepoURL, dest); err != nil {
			return fmt.Errorf("cloning repo: %w", err)
		}
	}

	return nil
}

// LaunchAgent starts a claude CLI process for an agent.
// CRITICAL: uses context.Background() so the agent outlives the HTTP request.
func (r *LocalRuntime) LaunchAgent(_ context.Context, opts AgentOpts) (*AgentProcess, error) {
	// Acquire concurrency slot before starting.
	select {
	case r.semaphore <- struct{}{}:
	default:
		return nil, fmt.Errorf("concurrency limit reached (%d agents running)", r.maxConcurrent)
	}

	agentCtx, cancel := context.WithCancel(context.Background())

	args := r.buildAgentArgs(opts)
	cmd := exec.CommandContext(agentCtx, "claude", args...)
	cmd.Dir = opts.WorkDir
	cmd.Env = append(os.Environ(), opts.Env...)

	if err := cmd.Start(); err != nil {
		cancel()
		<-r.semaphore
		return nil, fmt.Errorf("launching agent %s: %w", opts.AgentID, err)
	}

	done := make(chan struct{})

	entry := &agentEntry{
		cancel: cancel,
		done:   done,
		pid:    cmd.Process.Pid,
	}

	r.mu.Lock()
	r.agents[opts.AgentID] = entry
	r.mu.Unlock()

	r.writeAgentState(opts, "running", cmd.Process.Pid)

	// Reaper goroutine: clean up when the process exits.
	go func() {
		defer func() {
			cancel()
			<-r.semaphore
			close(done)

			r.mu.Lock()
			delete(r.agents, opts.AgentID)
			r.mu.Unlock()

			r.writeAgentState(opts, "stopped", cmd.Process.Pid)
		}()

		_ = cmd.Wait()
	}()

	return &AgentProcess{
		AgentID: opts.AgentID,
		PID:     cmd.Process.Pid,
		Done:    done,
	}, nil
}

// StopAgent cancels the context of a running agent and waits up to 10 seconds.
func (r *LocalRuntime) StopAgent(_ context.Context, agentID string) error {
	r.mu.Lock()
	entry, ok := r.agents[agentID]
	r.mu.Unlock()

	if !ok {
		return fmt.Errorf("agent %s not found", agentID)
	}

	entry.cancel()

	select {
	case <-entry.done:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timed out waiting for agent %s to stop", agentID)
	}

	return nil
}

// IsAgentRunning reports whether an agent is currently tracked.
func (r *LocalRuntime) IsAgentRunning(agentID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.agents[agentID]
	return ok
}

// WatchEvents sets up an fsnotify watcher on the project's shared/ directory
// and returns a channel of classified events.
func (r *LocalRuntime) WatchEvents(ctx context.Context, projectID string) (<-chan Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.watchers[projectID]; exists {
		return nil, fmt.Errorf("watcher already active for project %s", projectID)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	r.watchers[projectID] = watcher

	events := make(chan Event, 64)

	go r.watchLoop(ctx, projectID, watcher, events)

	return events, nil
}

// AddWatchPath adds a path to an existing project watcher.
func (r *LocalRuntime) AddWatchPath(projectID, path string) error {
	r.mu.Lock()
	watcher, ok := r.watchers[projectID]
	r.mu.Unlock()

	if !ok {
		return fmt.Errorf("no active watcher for project %s", projectID)
	}

	return watcher.Add(path)
}

// Cleanup closes the watcher for a project, if any.
func (r *LocalRuntime) Cleanup(_ context.Context, projectID string) error {
	r.mu.Lock()
	watcher, ok := r.watchers[projectID]
	if ok {
		delete(r.watchers, projectID)
	}
	r.mu.Unlock()

	if ok {
		return watcher.Close()
	}

	return nil
}

// watchLoop reads fsnotify events and classifies them into Event values.
func (r *LocalRuntime) watchLoop(ctx context.Context, projectID string, watcher *fsnotify.Watcher, out chan<- Event) {
	defer close(out)

	for {
		select {
		case <-ctx.Done():
			return

		case fsEvent, ok := <-watcher.Events:
			if !ok {
				return
			}

			if fsEvent.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			payload, err := os.ReadFile(fsEvent.Name)
			if err != nil {
				continue
			}

			// Skip empty files and non-JSON content.
			if len(payload) == 0 {
				continue
			}

			if !json.Valid(payload) {
				continue
			}

			evt := Event{
				ProjectID: projectID,
				Type:      classifyByDirectory(fsEvent.Name),
				Path:      fsEvent.Name,
				Payload:   payload,
			}

			select {
			case out <- evt:
			case <-ctx.Done():
				return
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("local runtime watcher error (project %s): %v", projectID, err)
		}
	}
}

// classifyByDirectory maps a file path's parent directory to an event type.
func classifyByDirectory(path string) string {
	dir := filepath.Base(filepath.Dir(path))
	switch dir {
	case "status":
		return "agent.status"
	case "messages":
		return "agent.message"
	case "reviews":
		return "agent.review"
	case "contracts":
		return "agent.contract"
	case "designs":
		return "agent.design"
	case "context":
		return "agent.context"
	default:
		return "agent.output"
	}
}

// buildAgentArgs constructs the claude CLI argument slice.
// Always includes --verbose (required with -p + stream-json).
// Never includes --bare or --max-budget-usd for local mode.
func (r *LocalRuntime) buildAgentArgs(opts AgentOpts) []string {
	return []string{
		"--output-format", "stream-json",
		"--verbose",
		"--dangerously-skip-permissions",
		"-p", opts.Prompt,
	}
}

// cloneRepo runs git clone into dest. Skips if dest is already populated.
func (r *LocalRuntime) cloneRepo(ctx context.Context, repoURL, dest string) error {
	entries, err := os.ReadDir(dest)
	if err == nil && len(entries) > 0 {
		// Directory already has content — skip clone.
		return nil
	}

	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone %s: %w\n%s", repoURL, err, string(out))
	}

	return nil
}

// writeAgentState persists agent state as a JSON file in the project's state/ dir.
func (r *LocalRuntime) writeAgentState(opts AgentOpts, status string, pid int) {
	// We don't know the WorkDir layout relative to foundryHome here, so we
	// write state into the WorkDir itself if it looks like a project dir.
	stateDir := filepath.Join(opts.WorkDir, "..", "..", "state")
	stateDir = filepath.Clean(stateDir)

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return
	}

	state := map[string]interface{}{
		"agent_id":   opts.AgentID,
		"project_id": opts.ProjectID,
		"status":     status,
		"pid":        pid,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	path := filepath.Join(stateDir, opts.AgentID+".json")
	_ = os.WriteFile(path, data, 0644)
}

// parseEventFile reads and JSON-decodes a file into a map. Used for testing.
func parseEventFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading event file: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing event file: %w", err)
	}

	return result, nil
}

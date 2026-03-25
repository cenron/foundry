package runtime

import "context"

// Runtime abstracts where and how agents are launched.
type Runtime interface {
	Setup(ctx context.Context, opts SetupOpts) error
	LaunchAgent(ctx context.Context, opts AgentOpts) (*AgentProcess, error)
	StopAgent(ctx context.Context, agentID string) error
	WatchEvents(ctx context.Context, projectID string) (<-chan Event, error)
	Cleanup(ctx context.Context, projectID string) error
	IsAgentRunning(agentID string) bool
}

// SetupOpts holds parameters for setting up a project workspace.
type SetupOpts struct {
	ProjectID string
	RepoURL   string
	WorkDir   string // absolute path to foundry home
}

// AgentOpts holds parameters for launching an agent process.
type AgentOpts struct {
	AgentID   string
	ProjectID string
	Role      string
	Prompt    string
	WorkDir   string // agent working directory
	Env       []string
}

// AgentProcess represents a running agent.
type AgentProcess struct {
	AgentID string
	PID     int
	Done    <-chan struct{}
}

// Event is a filesystem event from the shared workspace directory.
type Event struct {
	ProjectID string
	Type      string
	Path      string
	Payload   []byte
}

package po

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/google/uuid"
)

// SessionType constants for PO session types.
const (
	SessionTypePlanning        = "planning"
	SessionTypeEstimation      = "estimation"
	SessionTypeReview          = "review"
	SessionTypeExecutionChat   = "execution-chat"
	SessionTypeEscalation      = "escalation"
	SessionTypePhaseTransition = "phase-transition"
)

// SessionStatus constants for PO session lifecycle.
const (
	SessionStatusStarting = "starting"
	SessionStatusActive   = "active"
	SessionStatusStopped  = "stopped"
)

// POSessionOpts holds the parameters for starting a PO session.
type POSessionOpts struct {
	Type    string // planning, estimation, review, execution-chat, escalation, phase-transition
	Project string // project name
	Trigger string // user, system
	Message string // user message or system prompt

	// System-triggered fields:
	TaskID    string
	TaskTitle string
	AgentRole string
	RiskLevel string
	Branch    string
}

// POSession represents a running Claude Code PO process.
type POSession struct {
	ID          string
	ProjectName string
	Type        string
	PID         int
	Status      string
	cmd         *exec.Cmd
	stdout      io.ReadCloser
	cancel      context.CancelFunc
}

// sessionTierConfig holds the model tier and behavioral flags for a session type.
type sessionTierConfig struct {
	model          string
	maxTurns       int
	maxBudgetUSD   float64 // 0 means no budget limit
	systemTriggered bool
}

var sessionTiers = map[string]sessionTierConfig{
	SessionTypePlanning: {
		model:           "opus",
		maxTurns:        50,
		systemTriggered: false,
	},
	SessionTypeEstimation: {
		model:           "opus",
		maxTurns:        30,
		maxBudgetUSD:    5.00,
		systemTriggered: true,
	},
	SessionTypeReview: {
		model:           "sonnet",
		maxTurns:        20,
		maxBudgetUSD:    2.00,
		systemTriggered: true,
	},
	SessionTypeExecutionChat: {
		model:           "sonnet",
		maxTurns:        30,
		systemTriggered: false,
	},
	SessionTypeEscalation: {
		model:           "sonnet",
		maxTurns:        20,
		maxBudgetUSD:    2.00,
		systemTriggered: true,
	},
	SessionTypePhaseTransition: {
		model:           "sonnet",
		maxTurns:        20,
		maxBudgetUSD:    2.00,
		systemTriggered: true,
	},
}

// SessionManager launches and tracks PO Claude Code sessions.
type SessionManager struct {
	foundryHome   string
	apiKey        string
	claudeVersion string
	mu            sync.Mutex
	sessions      map[string]*POSession // projectName -> active session
}

// NewSessionManager creates a SessionManager for the given foundry home directory.
func NewSessionManager(foundryHome, apiKey, claudeVersion string) *SessionManager {
	return &SessionManager{
		foundryHome:   foundryHome,
		apiKey:        apiKey,
		claudeVersion: claudeVersion,
		sessions:      make(map[string]*POSession),
	}
}

// StartSession launches a PO Claude Code session for the given project.
func (m *SessionManager) StartSession(ctx context.Context, opts POSessionOpts) (*POSession, error) {
	sessionCtx, cancel := context.WithCancel(ctx)

	cmd := m.BuildCommand(sessionCtx, opts)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("starting PO session for project %q: %w", opts.Project, err)
	}

	session := &POSession{
		ID:          uuid.New().String(),
		ProjectName: opts.Project,
		Type:        opts.Type,
		PID:         cmd.Process.Pid,
		Status:      SessionStatusActive,
		cmd:         cmd,
		stdout:      stdout,
		cancel:      cancel,
	}

	m.mu.Lock()
	m.sessions[opts.Project] = session
	m.mu.Unlock()

	return session, nil
}

// StopSession gracefully stops the active PO session for a project.
func (m *SessionManager) StopSession(projectName string) error {
	m.mu.Lock()
	session, ok := m.sessions[projectName]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("no active session for project %q", projectName)
	}
	delete(m.sessions, projectName)
	m.mu.Unlock()

	session.Status = SessionStatusStopped

	if session.cancel != nil {
		session.cancel()
	}

	if session.cmd != nil && session.cmd.Process != nil {
		_ = session.cmd.Process.Kill()
	}

	return nil
}

// GetSession returns the active session for a project, or nil if none exists.
func (m *SessionManager) GetSession(projectName string) *POSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[projectName]
}

// IsActive reports whether a project has an active PO session.
func (m *SessionManager) IsActive(projectName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.sessions[projectName]
	return ok
}

// BuildCommand constructs the exec.Cmd for a PO session without starting it.
// Exported for testability — callers can inspect args, env, and dir.
func (m *SessionManager) BuildCommand(ctx context.Context, opts POSessionOpts) *exec.Cmd {
	tier := sessionTiers[opts.Type]
	sessionCtx := BuildSessionContext(opts)

	args := []string{
		"--output-format", "stream-json",
		"--model", tier.model,
		"--max-turns", fmt.Sprintf("%d", tier.maxTurns),
	}

	if tier.maxBudgetUSD > 0 {
		args = append(args, "--max-budget-usd", fmt.Sprintf("%.2f", tier.maxBudgetUSD))
	}

	if tier.systemTriggered {
		args = append(args, "--bare", "--dangerously-skip-permissions")
	}

	args = append(args, "--append-system-prompt", sessionCtx)
	args = append(args, "-p", opts.Message)

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = m.foundryHome
	cmd.Env = append(cmd.Environ(), "ANTHROPIC_API_KEY="+m.apiKey)

	return cmd
}

// InjectSession inserts a session directly into the manager's tracking map.
// Intended for testing only — allows verifying IsActive, GetSession, and StopSession
// without launching a real process.
func (m *SessionManager) InjectSession(projectName string, session *POSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[projectName] = session
}

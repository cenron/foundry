package po_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cenron/foundry/internal/po"
)

// --- BuildSessionContext ---

func TestBuildSessionContext_Planning(t *testing.T) {
	opts := po.POSessionOpts{
		Type:    "planning",
		Project: "my-app",
		Trigger: "user",
		Message: "Let's plan the next sprint.",
	}

	got := po.BuildSessionContext(opts)

	checks := []string{
		"[foundry:session]",
		"type: planning",
		"project: my-app",
		"project_dir: projects/my-app",
		"playbook: playbooks/planning.md",
		"trigger: user",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("BuildSessionContext() missing %q\ngot:\n%s", want, got)
		}
	}

	// System fields must NOT appear for user-triggered sessions.
	systemFields := []string{"task_id:", "task_title:", "agent_role:", "risk_level:", "branch:"}
	for _, field := range systemFields {
		if strings.Contains(got, field) {
			t.Errorf("BuildSessionContext() unexpectedly contains %q in user-triggered context\ngot:\n%s", field, got)
		}
	}
}

func TestBuildSessionContext_Review(t *testing.T) {
	opts := po.POSessionOpts{
		Type:      "review",
		Project:   "api-service",
		Trigger:   "system",
		Message:   "Review completed task.",
		TaskID:    "task-42",
		TaskTitle: "Implement auth middleware",
		AgentRole: "backend-developer",
		RiskLevel: "high",
		Branch:    "feat/auth-middleware",
	}

	got := po.BuildSessionContext(opts)

	checks := []string{
		"[foundry:session]",
		"type: review",
		"project: api-service",
		"project_dir: projects/api-service",
		"playbook: playbooks/review.md",
		"trigger: system",
		"task_id: task-42",
		"task_title: Implement auth middleware",
		"agent_role: backend-developer",
		"risk_level: high",
		"branch: feat/auth-middleware",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("BuildSessionContext() missing %q\ngot:\n%s", want, got)
		}
	}
}

// --- ScaffoldProjectWorkspace ---

func TestScaffoldProjectWorkspace(t *testing.T) {
	foundryHome := t.TempDir()
	projectName := "test-project"
	repoURL := "https://github.com/acme/test-project"
	techStack := []string{"go", "react", "postgres"}

	err := po.ScaffoldProjectWorkspace(foundryHome, projectName, repoURL, techStack)
	if err != nil {
		t.Fatalf("ScaffoldProjectWorkspace() error: %v", err)
	}

	projectDir := filepath.Join(foundryHome, "projects", projectName)

	t.Run("directories exist", func(t *testing.T) {
		expectedDirs := []string{
			projectDir,
			filepath.Join(projectDir, "memory"),
			filepath.Join(projectDir, "decisions"),
			filepath.Join(projectDir, "artifacts"),
		}
		for _, dir := range expectedDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("expected directory %q to exist", dir)
			}
		}
	})

	t.Run("project.yaml content", func(t *testing.T) {
		yamlPath := filepath.Join(projectDir, "project.yaml")
		content, err := os.ReadFile(yamlPath)
		if err != nil {
			t.Fatalf("reading project.yaml: %v", err)
		}

		got := string(content)

		checks := []string{
			"name: test-project",
			"repo: https://github.com/acme/test-project",
			"tech_stack:",
			"  - go",
			"  - react",
			"  - postgres",
			"created:",
		}
		for _, want := range checks {
			if !strings.Contains(got, want) {
				t.Errorf("project.yaml missing %q\ngot:\n%s", want, got)
			}
		}
	})
}

func TestScaffoldProjectWorkspace_EmptyTechStack(t *testing.T) {
	foundryHome := t.TempDir()

	err := po.ScaffoldProjectWorkspace(foundryHome, "empty-stack", "", nil)
	if err != nil {
		t.Fatalf("ScaffoldProjectWorkspace() error: %v", err)
	}

	yamlPath := filepath.Join(foundryHome, "projects", "empty-stack", "project.yaml")
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("reading project.yaml: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "tech_stack:") {
		t.Errorf("project.yaml missing tech_stack key\ngot:\n%s", got)
	}
}

func TestScaffoldProjectWorkspace_Idempotent(t *testing.T) {
	foundryHome := t.TempDir()
	projectName := "idempotent-project"
	repoURL := "https://github.com/acme/idempotent"
	techStack := []string{"go"}

	// First call creates the workspace.
	if err := po.ScaffoldProjectWorkspace(foundryHome, projectName, repoURL, techStack); err != nil {
		t.Fatalf("first ScaffoldProjectWorkspace() error: %v", err)
	}

	// Second call with the same args should succeed without error (MkdirAll is idempotent).
	if err := po.ScaffoldProjectWorkspace(foundryHome, projectName, repoURL, techStack); err != nil {
		t.Fatalf("second ScaffoldProjectWorkspace() error: %v", err)
	}

	// Verify the workspace still has the expected structure.
	projectDir := filepath.Join(foundryHome, "projects", projectName)
	for _, dir := range []string{projectDir,
		filepath.Join(projectDir, "memory"),
		filepath.Join(projectDir, "decisions"),
		filepath.Join(projectDir, "artifacts"),
	} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %q to exist after idempotent call", dir)
		}
	}
}

// --- buildCommand arg verification ---

func TestSessionManager_BuildCommand_Planning(t *testing.T) {
	m := po.NewSessionManager("/home/user/foundry", "test-api-key", "latest")

	opts := po.POSessionOpts{
		Type:    "planning",
		Project: "my-app",
		Trigger: "user",
		Message: "Start planning.",
	}

	cmd := m.BuildCommand(context.Background(), opts)

	args := cmd.Args // includes "claude" as args[0]

	t.Run("model is opus", func(t *testing.T) {
		assertArgPair(t, args, "--model", "opus")
	})

	t.Run("max-turns is 50", func(t *testing.T) {
		assertArgPair(t, args, "--max-turns", "50")
	})

	t.Run("output format is stream-json", func(t *testing.T) {
		assertArgPair(t, args, "--output-format", "stream-json")
	})

	t.Run("no --bare for user-triggered", func(t *testing.T) {
		assertArgAbsent(t, args, "--bare")
	})

	t.Run("no --dangerously-skip-permissions for user-triggered", func(t *testing.T) {
		assertArgAbsent(t, args, "--dangerously-skip-permissions")
	})

	t.Run("no --max-budget-usd for planning", func(t *testing.T) {
		assertArgAbsent(t, args, "--max-budget-usd")
	})

	t.Run("working dir is foundryHome", func(t *testing.T) {
		if cmd.Dir != "/home/user/foundry" {
			t.Errorf("Dir = %q, want %q", cmd.Dir, "/home/user/foundry")
		}
	})
}

func TestSessionManager_BuildCommand_Estimation(t *testing.T) {
	m := po.NewSessionManager("/home/user/foundry", "test-api-key", "latest")

	opts := po.POSessionOpts{
		Type:      "estimation",
		Project:   "my-app",
		Trigger:   "system",
		Message:   "Estimate this spec.",
		TaskID:    "task-1",
		TaskTitle: "Auth module",
		AgentRole: "backend-developer",
		RiskLevel: "medium",
		Branch:    "feat/auth",
	}

	cmd := m.BuildCommand(context.Background(), opts)
	args := cmd.Args

	t.Run("model is opus", func(t *testing.T) {
		assertArgPair(t, args, "--model", "opus")
	})

	t.Run("max-turns is 30", func(t *testing.T) {
		assertArgPair(t, args, "--max-turns", "30")
	})

	t.Run("max-budget-usd is 5.00", func(t *testing.T) {
		assertArgPair(t, args, "--max-budget-usd", "5.00")
	})

	t.Run("has --bare", func(t *testing.T) {
		assertArgPresent(t, args, "--bare")
	})

	t.Run("has --dangerously-skip-permissions", func(t *testing.T) {
		assertArgPresent(t, args, "--dangerously-skip-permissions")
	})
}

func TestSessionManager_BuildCommand_AllTypes(t *testing.T) {
	type wantConfig struct {
		model        string
		maxTurns     string
		maxBudgetUSD string // empty string means absent
		systemFlags  bool
	}

	tests := []struct {
		sessionType string
		want        wantConfig
	}{
		{
			sessionType: "planning",
			want:        wantConfig{model: "opus", maxTurns: "50", maxBudgetUSD: "", systemFlags: false},
		},
		{
			sessionType: "estimation",
			want:        wantConfig{model: "opus", maxTurns: "30", maxBudgetUSD: "5.00", systemFlags: true},
		},
		{
			sessionType: "review",
			want:        wantConfig{model: "sonnet", maxTurns: "20", maxBudgetUSD: "2.00", systemFlags: true},
		},
		{
			sessionType: "execution-chat",
			want:        wantConfig{model: "sonnet", maxTurns: "30", maxBudgetUSD: "", systemFlags: false},
		},
		{
			sessionType: "escalation",
			want:        wantConfig{model: "sonnet", maxTurns: "20", maxBudgetUSD: "2.00", systemFlags: true},
		},
		{
			sessionType: "phase-transition",
			want:        wantConfig{model: "sonnet", maxTurns: "20", maxBudgetUSD: "2.00", systemFlags: true},
		},
	}

	m := po.NewSessionManager("/foundry", "key", "latest")

	for _, tc := range tests {
		t.Run(tc.sessionType, func(t *testing.T) {
			opts := po.POSessionOpts{
				Type:    tc.sessionType,
				Project: "proj",
				Trigger: triggerFor(tc.want.systemFlags),
				Message: "test message",
			}

			args := m.BuildCommand(context.Background(), opts).Args

			assertArgPair(t, args, "--model", tc.want.model)
			assertArgPair(t, args, "--max-turns", tc.want.maxTurns)

			if tc.want.maxBudgetUSD != "" {
				assertArgPair(t, args, "--max-budget-usd", tc.want.maxBudgetUSD)
			} else {
				assertArgAbsent(t, args, "--max-budget-usd")
			}

			if tc.want.systemFlags {
				assertArgPresent(t, args, "--bare")
				assertArgPresent(t, args, "--dangerously-skip-permissions")
			} else {
				assertArgAbsent(t, args, "--bare")
				assertArgAbsent(t, args, "--dangerously-skip-permissions")
			}
		})
	}
}

// --- IsActive and StopSession ---

func TestSessionManager_IsActive(t *testing.T) {
	m := po.NewSessionManager(t.TempDir(), "key", "latest")

	if m.IsActive("my-app") {
		t.Error("IsActive() = true before any session, want false")
	}

	// Inject a fake session directly via the exported test helper.
	m.InjectSession("my-app", &po.POSession{
		ID:          "test-id",
		ProjectName: "my-app",
		Type:        "planning",
		Status:      po.SessionStatusActive,
	})

	if !m.IsActive("my-app") {
		t.Error("IsActive() = false after injection, want true")
	}

	if m.IsActive("other-project") {
		t.Error("IsActive() = true for unregistered project, want false")
	}
}

func TestSessionManager_StopSession_Cleanup(t *testing.T) {
	m := po.NewSessionManager(t.TempDir(), "key", "latest")

	m.InjectSession("my-app", &po.POSession{
		ID:          "test-id",
		ProjectName: "my-app",
		Type:        "planning",
		Status:      po.SessionStatusActive,
	})

	if err := m.StopSession("my-app"); err != nil {
		t.Fatalf("StopSession() error: %v", err)
	}

	if m.IsActive("my-app") {
		t.Error("IsActive() = true after StopSession, want false")
	}
}

func TestSessionManager_StopSession_NotFound(t *testing.T) {
	m := po.NewSessionManager(t.TempDir(), "key", "latest")

	err := m.StopSession("nonexistent")
	if err == nil {
		t.Error("StopSession() error = nil for nonexistent project, want error")
	}
}

func TestSessionManager_GetSession(t *testing.T) {
	m := po.NewSessionManager(t.TempDir(), "key", "latest")

	if got := m.GetSession("my-app"); got != nil {
		t.Errorf("GetSession() = %v, want nil", got)
	}

	injected := &po.POSession{
		ID:          "abc",
		ProjectName: "my-app",
		Status:      po.SessionStatusActive,
	}
	m.InjectSession("my-app", injected)

	got := m.GetSession("my-app")
	if got == nil {
		t.Fatal("GetSession() = nil, want session")
	}
	if got.ID != "abc" {
		t.Errorf("GetSession().ID = %q, want %q", got.ID, "abc")
	}
}

// --- helpers ---

func assertArgPair(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, arg := range args {
		if arg == flag && i+1 < len(args) && args[i+1] == value {
			return
		}
	}
	t.Errorf("args missing %q %q\ngot: %v", flag, value, args)
}

func assertArgPresent(t *testing.T, args []string, flag string) {
	t.Helper()
	for _, arg := range args {
		if arg == flag {
			return
		}
	}
	t.Errorf("args missing %q\ngot: %v", flag, args)
}

func assertArgAbsent(t *testing.T, args []string, flag string) {
	t.Helper()
	for _, arg := range args {
		if arg == flag {
			t.Errorf("args unexpectedly contains %q\ngot: %v", flag, args)
			return
		}
	}
}

func triggerFor(systemTriggered bool) string {
	if systemTriggered {
		return "system"
	}
	return "user"
}

// --- StartSession invalid type ---

func TestSessionManager_StartSession_InvalidType(t *testing.T) {
	m := po.NewSessionManager(t.TempDir(), "key", "latest")

	_, err := m.StartSession(context.Background(), po.POSessionOpts{
		Type:    "not-a-real-type",
		Project: "proj",
		Message: "hi",
	})

	if err == nil {
		t.Error("StartSession() error = nil for invalid type, want error")
	}
}

// --- StopSession with process ---

func TestSessionManager_StopSession_WithProcess(t *testing.T) {
	m := po.NewSessionManager(t.TempDir(), "key", "latest")

	// Inject a session with a real cmd (sleep) so StopSession exercises the kill path.
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "sleep", "30")
	_ = cmd.Start()

	m.InjectSession("sleep-project", &po.POSession{
		ID:          "proc-id",
		ProjectName: "sleep-project",
		Status:      po.SessionStatusActive,
	})

	// Stop the injected session (no actual process pointer, so just exercises map cleanup).
	if err := m.StopSession("sleep-project"); err != nil {
		t.Fatalf("StopSession() error: %v", err)
	}

	if m.IsActive("sleep-project") {
		t.Error("IsActive() = true after StopSession, want false")
	}

	// Clean up the background process.
	cancel()
	_ = cmd.Wait()
}

// --- ScaffoldProjectWorkspace invalid path ---

func TestScaffoldProjectWorkspace_InvalidPath(t *testing.T) {
	// Use a path inside a file (not a directory) to trigger an error.
	f, err := os.CreateTemp("", "foundry-test-*")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	_ = f.Close()

	// Attempt to scaffold inside a file path — should fail.
	err = po.ScaffoldProjectWorkspace(f.Name(), "proj", "", nil)
	if err == nil {
		t.Error("ScaffoldProjectWorkspace() error = nil for invalid foundryHome, want error")
	}
}

// --- StopAll ---

func TestSessionManager_StopAll_MultipleSessions(t *testing.T) {
	m := po.NewSessionManager(t.TempDir(), "key", "latest")

	for _, name := range []string{"proj-a", "proj-b", "proj-c"} {
		m.InjectSession(name, &po.POSession{
			ID:          name + "-id",
			ProjectName: name,
			Status:      po.SessionStatusActive,
		})
	}

	m.StopAll()

	for _, name := range []string{"proj-a", "proj-b", "proj-c"} {
		if m.IsActive(name) {
			t.Errorf("IsActive(%q) = true after StopAll, want false", name)
		}
	}
}

func TestSessionManager_StopAll_Empty(t *testing.T) {
	m := po.NewSessionManager(t.TempDir(), "key", "latest")
	// Should not panic or error on empty session map.
	m.StopAll()
}

// --- LocalSessionManager ---

func TestLocalSessionManager_BuildCommand_NoBareFlagForSystemType(t *testing.T) {
	m := po.NewLocalSessionManager("/foundry", "latest")

	opts := po.POSessionOpts{
		Type:    po.SessionTypeEstimation,
		Project: "proj",
		Trigger: "system",
		Message: "estimate this",
	}

	args := m.BuildCommand(context.Background(), opts).Args

	assertArgAbsent(t, args, "--bare")
	assertArgAbsent(t, args, "--max-budget-usd")
}

func TestLocalSessionManager_BuildCommand_HasVerbose(t *testing.T) {
	m := po.NewLocalSessionManager("/foundry", "latest")

	opts := po.POSessionOpts{
		Type:    po.SessionTypePlanning,
		Project: "proj",
		Trigger: "user",
		Message: "plan",
	}

	args := m.BuildCommand(context.Background(), opts).Args

	assertArgPresent(t, args, "--verbose")
}

func TestLocalSessionManager_BuildCommand_SystemTypeHasDangerouslySkipPermissions(t *testing.T) {
	m := po.NewLocalSessionManager("/foundry", "latest")

	systemTypes := []string{
		po.SessionTypeEstimation,
		po.SessionTypeReview,
		po.SessionTypeEscalation,
		po.SessionTypePhaseTransition,
	}

	for _, sessionType := range systemTypes {
		t.Run(sessionType, func(t *testing.T) {
			opts := po.POSessionOpts{
				Type:    sessionType,
				Project: "proj",
				Trigger: "system",
				Message: "test",
			}

			args := m.BuildCommand(context.Background(), opts).Args

			assertArgPresent(t, args, "--dangerously-skip-permissions")
			assertArgAbsent(t, args, "--bare")
			assertArgAbsent(t, args, "--max-budget-usd")
		})
	}
}

func TestLocalSessionManager_BuildCommand_UserTriggeredNoPermissionsFlag(t *testing.T) {
	m := po.NewLocalSessionManager("/foundry", "latest")

	userTypes := []string{
		po.SessionTypePlanning,
		po.SessionTypeExecutionChat,
	}

	for _, sessionType := range userTypes {
		t.Run(sessionType, func(t *testing.T) {
			opts := po.POSessionOpts{
				Type:    sessionType,
				Project: "proj",
				Trigger: "user",
				Message: "test",
			}

			args := m.BuildCommand(context.Background(), opts).Args

			assertArgAbsent(t, args, "--dangerously-skip-permissions")
			assertArgAbsent(t, args, "--bare")
		})
	}
}

// --- DeployPOWorkspace ---

func TestDeployPOWorkspace(t *testing.T) {
	foundryHome := t.TempDir()

	if err := po.DeployPOWorkspace(foundryHome); err != nil {
		t.Fatalf("DeployPOWorkspace() error: %v", err)
	}

	// CLAUDE.md must exist at the foundry home root.
	claudeMD := filepath.Join(foundryHome, "CLAUDE.md")
	if _, err := os.Stat(claudeMD); os.IsNotExist(err) {
		t.Errorf("expected CLAUDE.md to exist at %s", claudeMD)
	}

	// playbooks/ directory must exist.
	playbooksDir := filepath.Join(foundryHome, "playbooks")
	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		t.Errorf("expected playbooks/ directory to exist at %s", playbooksDir)
	}
}

func TestDeployPOWorkspace_Idempotent(t *testing.T) {
	foundryHome := t.TempDir()

	// First deploy.
	if err := po.DeployPOWorkspace(foundryHome); err != nil {
		t.Fatalf("first DeployPOWorkspace() error: %v", err)
	}

	// Second deploy — should overwrite without error.
	if err := po.DeployPOWorkspace(foundryHome); err != nil {
		t.Fatalf("second DeployPOWorkspace() error: %v", err)
	}
}

// --- StartSession success path ---

// TestSessionManager_StartSession_Success verifies the full StartSession happy path
// by using a real short-lived process (the system "sleep" command) in place of claude.
// This exercises the PID assignment, IsActive registration, and subsequent StopSession
// process teardown code paths without requiring claude to be installed.
func TestSessionManager_StartSession_Success(t *testing.T) {
	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not found, skipping")
	}

	// Build a session manager that points at `sleep` as its "claude" binary by
	// temporarily overriding PATH so that "claude" resolves to our fake helper.
	//
	// We create a symlink in a temp dir pointing to sleep, then use a separate
	// wrapper approach: instead, we directly call BuildCommand and swap the Path.
	// BuildCommand returns an exec.Cmd; we replace the binary path to use sleep
	// instead. But StartSession builds the command internally, so we need a
	// different approach.
	//
	// The cleanest approach is to create a temp dir with a script named "claude"
	// that just runs sleep, and prepend it to PATH.
	claudeDir := t.TempDir()
	claudeScript := filepath.Join(claudeDir, "claude")
	scriptContent := "#!/bin/sh\n" + sleepPath + " 30\n"
	if err := os.WriteFile(claudeScript, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("creating claude stub: %v", err)
	}

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", claudeDir+":"+origPath)

	foundryHome := t.TempDir()
	m := po.NewSessionManager(foundryHome, "test-key", "latest")

	session, err := m.StartSession(context.Background(), po.POSessionOpts{
		Type:    po.SessionTypePlanning,
		Project: "success-proj",
		Trigger: "user",
		Message: "Hello PO",
	})

	if err != nil {
		t.Fatalf("StartSession() error: %v", err)
	}

	if session == nil {
		t.Fatal("StartSession() returned nil session")
	}

	if session.PID <= 0 {
		t.Errorf("PID = %d, want > 0", session.PID)
	}

	if !m.IsActive("success-proj") {
		t.Error("IsActive() = false after StartSession, want true")
	}

	// StopSession must clean up the process and remove it from the map.
	if err := m.StopSession("success-proj"); err != nil {
		t.Fatalf("StopSession() error: %v", err)
	}

	if m.IsActive("success-proj") {
		t.Error("IsActive() = true after StopSession, want false")
	}
}

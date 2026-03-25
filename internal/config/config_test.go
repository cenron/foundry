package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cenron/foundry/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := config.Load()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"DatabaseURL", cfg.DatabaseURL, "postgres://foundry:foundry@localhost:5433/foundry?sslmode=disable"},
		{"RedisURL", cfg.RedisURL, "redis://localhost:6379"},
		{"RabbitMQURL", cfg.RabbitMQURL, "amqp://guest:guest@localhost:5672/"},
		{"APIPort", cfg.APIPort, "8080"},
		{"AgentLibraryPath", cfg.AgentLibraryPath, "./agents"},
		{"SSHKeyPath", cfg.SSHKeyPath, "~/.ssh"},
		{"FoundryHome", cfg.FoundryHome, "~/foundry"},
		{"ClaudeVersion", cfg.ClaudeVersion, "latest"},
		{"RuntimeMode", cfg.RuntimeMode, "docker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}

	if cfg.MaxConcurrentAgents != 4 {
		t.Errorf("MaxConcurrentAgents = %d, want 4", cfg.MaxConcurrentAgents)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://custom:custom@db:5432/custom")
	t.Setenv("REDIS_URL", "redis://custom:6380")
	t.Setenv("API_PORT", "9090")

	cfg := config.Load()

	if cfg.DatabaseURL != "postgres://custom:custom@db:5432/custom" {
		t.Errorf("DatabaseURL = %q, want custom override", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis://custom:6380" {
		t.Errorf("RedisURL = %q, want custom override", cfg.RedisURL)
	}
	if cfg.APIPort != "9090" {
		t.Errorf("APIPort = %q, want 9090", cfg.APIPort)
	}
}

func TestLoad_LocalMode(t *testing.T) {
	t.Setenv("FOUNDRY_RUNTIME_MODE", "local")

	cfg := config.Load()

	if !cfg.IsLocalMode() {
		t.Error("IsLocalMode() = false, want true when FOUNDRY_RUNTIME_MODE=local")
	}
}

func TestLoad_DockerModeIsNotLocal(t *testing.T) {
	t.Setenv("FOUNDRY_RUNTIME_MODE", "docker")

	cfg := config.Load()

	if cfg.IsLocalMode() {
		t.Error("IsLocalMode() = true for docker mode, want false")
	}
}

func TestLoad_MaxConcurrentAgents_Override(t *testing.T) {
	t.Setenv("FOUNDRY_MAX_CONCURRENT_AGENTS", "8")

	cfg := config.Load()

	if cfg.MaxConcurrentAgents != 8 {
		t.Errorf("MaxConcurrentAgents = %d, want 8", cfg.MaxConcurrentAgents)
	}
}

func TestLoad_MaxConcurrentAgents_InvalidFallback(t *testing.T) {
	t.Setenv("FOUNDRY_MAX_CONCURRENT_AGENTS", "not-a-number")

	cfg := config.Load()

	if cfg.MaxConcurrentAgents != 4 {
		t.Errorf("MaxConcurrentAgents = %d, want default 4 for invalid value", cfg.MaxConcurrentAgents)
	}
}

func TestLoad_ExpandHome_EnvVar(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	customPath := "~/my-foundry"
	t.Setenv("FOUNDRY_HOME", customPath)

	cfg := config.Load()

	want := filepath.Join(home, "my-foundry")
	if cfg.FoundryHome != want {
		t.Errorf("FoundryHome = %q, want %q (home expanded)", cfg.FoundryHome, want)
	}
}

func TestLoad_ExpandHome_NoTildePrefix_IsReturnedAsIs(t *testing.T) {
	customPath := "/absolute/path/to/foundry"
	t.Setenv("FOUNDRY_HOME", customPath)

	cfg := config.Load()

	if cfg.FoundryHome != customPath {
		t.Errorf("FoundryHome = %q, want %q (no tilde expansion needed)", cfg.FoundryHome, customPath)
	}
}

func TestLoad_ExpandHome_HomeDirError_ReturnsTildePathUnchanged(t *testing.T) {
	// Unset HOME and USERPROFILE so that os.UserHomeDir() returns an error.
	// This exercises the err != nil branch in expandHome.
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("FOUNDRY_HOME", "~/foundry-error-path")

	cfg := config.Load()

	// When home dir cannot be resolved, expandHome returns the tilde path unchanged.
	if cfg.FoundryHome != "~/foundry-error-path" {
		// On some systems (e.g. macOS via Directory Services), UserHomeDir may
		// still succeed even with HOME unset. In that case the path gets expanded
		// and we cannot reliably trigger the error branch — skip gracefully.
		if _, err := os.UserHomeDir(); err == nil {
			t.Skip("os.UserHomeDir() succeeded despite HOME being unset; cannot test error branch")
		}
		t.Errorf("FoundryHome = %q, want tilde path unchanged", cfg.FoundryHome)
	}
}

func TestLoad_EnvOrInt_ValidValue(t *testing.T) {
	t.Setenv("FOUNDRY_MAX_CONCURRENT_AGENTS", "16")

	cfg := config.Load()

	if cfg.MaxConcurrentAgents != 16 {
		t.Errorf("MaxConcurrentAgents = %d, want 16", cfg.MaxConcurrentAgents)
	}
}

func TestLoad_SSHKeyPath_DefaultNotExpanded(t *testing.T) {
	// When SSH_KEY_PATH env var is not set, the default "~/.ssh" is returned as-is
	// (expandHome is only called when the env var is non-empty).
	cfg := config.Load()

	if cfg.SSHKeyPath != "~/.ssh" {
		t.Errorf("SSHKeyPath = %q, want default %q (unexpanded)", cfg.SSHKeyPath, "~/.ssh")
	}
}

func TestLoad_SSHKeyPath_EnvExpanded(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	t.Setenv("FOUNDRY_SSH_KEY_PATH", "~/.ssh/foundry_rsa")

	cfg := config.Load()

	want := filepath.Join(home, ".ssh", "foundry_rsa")
	if cfg.SSHKeyPath != want {
		t.Errorf("SSHKeyPath = %q, want %q", cfg.SSHKeyPath, want)
	}
}

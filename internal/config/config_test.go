package config_test

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
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

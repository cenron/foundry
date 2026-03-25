package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL         string
	RedisURL            string
	RabbitMQURL         string
	APIPort             string
	AnthropicAPIKey     string
	GitToken            string
	AgentLibraryPath    string
	SSHKeyPath          string
	FoundryHome         string
	ClaudeVersion       string
	RuntimeMode         string
	MaxConcurrentAgents int
}

func Load() Config {
	return Config{
		DatabaseURL:         envOr("DATABASE_URL", "postgres://foundry:foundry@localhost:5433/foundry?sslmode=disable"),
		RedisURL:            envOr("REDIS_URL", "redis://localhost:6379"),
		RabbitMQURL:         envOr("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		APIPort:             envOr("API_PORT", "8080"),
		AnthropicAPIKey:     envOr("ANTHROPIC_API_KEY", ""),
		GitToken:            envOr("GIT_TOKEN", ""),
		AgentLibraryPath:    envOr("FOUNDRY_AGENT_LIBRARY", "./agents"),
		SSHKeyPath:          envOr("FOUNDRY_SSH_KEY_PATH", "~/.ssh"),
		FoundryHome:         envOr("FOUNDRY_HOME", "~/foundry"),
		ClaudeVersion:       envOr("FOUNDRY_CLAUDE_VERSION", "latest"),
		RuntimeMode:         envOr("FOUNDRY_RUNTIME_MODE", "docker"),
		MaxConcurrentAgents: envOrInt("FOUNDRY_MAX_CONCURRENT_AGENTS", 4),
	}
}

// IsLocalMode reports whether the runtime is configured for local mode.
func (c Config) IsLocalMode() bool {
	return c.RuntimeMode == "local"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return expandHome(v)
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// expandHome replaces a leading ~/ with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}

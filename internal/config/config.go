package config

import "os"

type Config struct {
	DatabaseURL      string
	RedisURL         string
	RabbitMQURL      string
	APIPort          string
	AnthropicAPIKey  string
	GitToken         string
	AgentLibraryPath string
	SSHKeyPath       string
	FoundryHome      string
	ClaudeVersion    string
}

func Load() Config {
	return Config{
		DatabaseURL:      envOr("DATABASE_URL", "postgres://foundry:foundry@localhost:5433/foundry?sslmode=disable"),
		RedisURL:         envOr("REDIS_URL", "redis://localhost:6379"),
		RabbitMQURL:      envOr("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		APIPort:          envOr("API_PORT", "8080"),
		AnthropicAPIKey:  envOr("ANTHROPIC_API_KEY", ""),
		GitToken:         envOr("GIT_TOKEN", ""),
		AgentLibraryPath: envOr("FOUNDRY_AGENT_LIBRARY", "./agents"),
		SSHKeyPath:       envOr("FOUNDRY_SSH_KEY_PATH", "~/.ssh"),
		FoundryHome:      envOr("FOUNDRY_HOME", "~/foundry"),
		ClaudeVersion:    envOr("FOUNDRY_CLAUDE_VERSION", "latest"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

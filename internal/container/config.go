package container

import "time"

type TeamContainerConfig struct {
	ProjectID       string
	RepoURL         string
	AnthropicKey    string
	RabbitMQURL     string
	GitToken        string
	NetworkName     string
	SharedVolPath   string   // host path: ~/foundry/projects/<name>/shared/
	AgentLibPath    string   // host path: FOUNDRY_AGENT_LIBRARY
	SSHKeyPath      string   // host path: FOUNDRY_SSH_KEY_PATH
	TeamComposition []string // agent roles
	ClaudeVersion   string
}

type ContainerStatus struct {
	ID      string
	State   string // running, exited, paused, created, etc.
	Health  string
	Started time.Time
}

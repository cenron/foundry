package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	networktypes "github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/docker/docker/pkg/stdcopy"
)

// DockerClient is the interface for Docker API operations used by Manager.
// Matches the signatures from github.com/docker/docker v28.
type DockerClient interface {
	ContainerCreate(ctx context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, networkingConfig *networktypes.NetworkingConfig, platform *ocispec.Platform, containerName string) (containertypes.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options containertypes.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options containertypes.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options containertypes.RemoveOptions) error
	ContainerInspect(ctx context.Context, containerID string) (containertypes.InspectResponse, error)
	ContainerLogs(ctx context.Context, containerID string, options containertypes.LogsOptions) (io.ReadCloser, error)
	ContainerExecCreate(ctx context.Context, containerID string, options containertypes.ExecOptions) (containertypes.ExecCreateResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config containertypes.ExecAttachOptions) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (containertypes.ExecInspect, error)
}

const (
	teamImageDefault = "foundry-team:dev"
	containerPrefix  = "foundry-team-"
)

type Manager struct {
	client DockerClient
	image  string
}

func NewManager(client DockerClient, image string) *Manager {
	if image == "" {
		image = teamImageDefault
	}
	return &Manager{client: client, image: image}
}

func (m *Manager) CreateTeam(ctx context.Context, cfg TeamContainerConfig) (string, error) {
	containerName := containerPrefix + cfg.ProjectID

	config := &containertypes.Config{
		Image: m.image,
		Env: []string{
			"ANTHROPIC_API_KEY=" + cfg.AnthropicKey,
			"RABBITMQ_URL=" + cfg.RabbitMQURL,
			"GIT_TOKEN=" + cfg.GitToken,
			"PROJECT_ID=" + cfg.ProjectID,
			"REPO_URL=" + cfg.RepoURL,
			"CLAUDE_VERSION=" + cfg.ClaudeVersion,
			"TEAM_COMPOSITION=" + strings.Join(cfg.TeamComposition, ","),
		},
		Labels: map[string]string{
			"foundry.project_id": cfg.ProjectID,
			"foundry.type":       "team",
		},
		WorkingDir: "/workspace",
	}

	hostConfig := &containertypes.HostConfig{
		Mounts: buildMounts(cfg),
		RestartPolicy: containertypes.RestartPolicy{
			Name: "no",
		},
	}

	var networkConfig *networktypes.NetworkingConfig
	if cfg.NetworkName != "" {
		networkConfig = &networktypes.NetworkingConfig{
			EndpointsConfig: map[string]*networktypes.EndpointSettings{
				cfg.NetworkName: {},
			},
		}
	}

	resp, err := m.client.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("creating team container: %w", err)
	}

	return resp.ID, nil
}

func buildMounts(cfg TeamContainerConfig) []mount.Mount {
	mounts := []mount.Mount{
		{
			Type:   mount.TypeVolume,
			Source: "foundry-team-" + cfg.ProjectID,
			Target: "/workspace",
		},
	}

	if cfg.SharedVolPath != "" {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: cfg.SharedVolPath,
			Target: "/shared",
		})
	}

	if cfg.AgentLibPath != "" {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   cfg.AgentLibPath,
			Target:   "/agents",
			ReadOnly: true,
		})
	}

	if cfg.SSHKeyPath != "" {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   cfg.SSHKeyPath,
			Target:   "/root/.ssh",
			ReadOnly: true,
		})
	}

	return mounts
}

func (m *Manager) StartTeam(ctx context.Context, containerID string) error {
	if err := m.client.ContainerStart(ctx, containerID, containertypes.StartOptions{}); err != nil {
		return fmt.Errorf("starting team container: %w", err)
	}
	return nil
}

func (m *Manager) StopTeam(ctx context.Context, containerID string) error {
	timeout := 30
	if err := m.client.ContainerStop(ctx, containerID, containertypes.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("stopping team container: %w", err)
	}
	return nil
}

func (m *Manager) RemoveTeam(ctx context.Context, containerID string) error {
	if err := m.client.ContainerRemove(ctx, containerID, containertypes.RemoveOptions{
		Force:         true,
		RemoveVolumes: false, // preserve workspace volume
	}); err != nil {
		return fmt.Errorf("removing team container: %w", err)
	}
	return nil
}

func (m *Manager) GetStatus(ctx context.Context, containerID string) (*ContainerStatus, error) {
	info, err := m.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}

	if info.State == nil {
		return nil, fmt.Errorf("inspecting container: State is nil for %s", containerID)
	}

	started, err := time.Parse(time.RFC3339Nano, info.State.StartedAt)
	if err != nil {
		started = time.Time{}
	}

	return &ContainerStatus{
		ID:      info.ID,
		State:   info.State.Status,
		Health:  containerHealthString(info.State),
		Started: started,
	}, nil
}

func containerHealthString(state *containertypes.State) string {
	if state.Health != nil {
		return string(state.Health.Status)
	}
	if state.Running {
		return "running"
	}
	return state.Status
}

func (m *Manager) StreamLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	reader, err := m.client.ContainerLogs(ctx, containerID, containertypes.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",
	})
	if err != nil {
		return nil, fmt.Errorf("streaming container logs: %w", err)
	}
	return reader, nil
}

func (m *Manager) ExecInTeam(ctx context.Context, containerID string, cmd []string) (string, error) {
	execResp, err := m.client.ContainerExecCreate(ctx, containerID, containertypes.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	})
	if err != nil {
		return "", fmt.Errorf("creating exec: %w", err)
	}

	attachResp, err := m.client.ContainerExecAttach(ctx, execResp.ID, containertypes.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("attaching to exec: %w", err)
	}
	defer attachResp.Close()

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
	if err != nil {
		return "", fmt.Errorf("reading exec output: %w", err)
	}

	inspect, err := m.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return "", fmt.Errorf("inspecting exec: %w", err)
	}

	if inspect.ExitCode != 0 {
		return "", fmt.Errorf("exec exited with code %d: %s", inspect.ExitCode, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

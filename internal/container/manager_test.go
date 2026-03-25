package container_test

import (
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	networktypes "github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/cenron/foundry/internal/container"
)

// mockConn implements net.Conn for testing HijackedResponse.
type mockConn struct {
	data   []byte
	offset int
}

func (c *mockConn) Read(b []byte) (int, error) {
	if c.offset >= len(c.data) {
		return 0, io.EOF
	}
	n := copy(b, c.data[c.offset:])
	c.offset += n
	return n, nil
}

func (c *mockConn) Write(b []byte) (int, error)        { return len(b), nil }
func (c *mockConn) Close() error                        { return nil }
func (c *mockConn) LocalAddr() net.Addr                 { return nil }
func (c *mockConn) RemoteAddr() net.Addr                { return nil }
func (c *mockConn) SetDeadline(_ time.Time) error       { return nil }
func (c *mockConn) SetReadDeadline(_ time.Time) error   { return nil }
func (c *mockConn) SetWriteDeadline(_ time.Time) error  { return nil }

// mockDockerClient records all calls for assertion.
type mockDockerClient struct {
	createConfig     *containertypes.Config
	createHostConfig *containertypes.HostConfig
	createNetwork    *networktypes.NetworkingConfig
	createName       string
	createResp       containertypes.CreateResponse

	startID string
	stopID  string
	stopOpts containertypes.StopOptions
	removeID string

	inspectResp containertypes.InspectResponse

	execOpts     containertypes.ExecOptions
	execOutput   string
	execExitCode int
}

func (m *mockDockerClient) ContainerCreate(_ context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, networkConfig *networktypes.NetworkingConfig, _ *ocispec.Platform, name string) (containertypes.CreateResponse, error) {
	m.createConfig = config
	m.createHostConfig = hostConfig
	m.createNetwork = networkConfig
	m.createName = name
	return m.createResp, nil
}

func (m *mockDockerClient) ContainerStart(_ context.Context, id string, _ containertypes.StartOptions) error {
	m.startID = id
	return nil
}

func (m *mockDockerClient) ContainerStop(_ context.Context, id string, opts containertypes.StopOptions) error {
	m.stopID = id
	m.stopOpts = opts
	return nil
}

func (m *mockDockerClient) ContainerRemove(_ context.Context, id string, _ containertypes.RemoveOptions) error {
	m.removeID = id
	return nil
}

func (m *mockDockerClient) ContainerInspect(_ context.Context, _ string) (containertypes.InspectResponse, error) {
	return m.inspectResp, nil
}

func (m *mockDockerClient) ContainerLogs(_ context.Context, _ string, _ containertypes.LogsOptions) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("log output")), nil
}

func (m *mockDockerClient) ContainerExecCreate(_ context.Context, _ string, opts containertypes.ExecOptions) (containertypes.ExecCreateResponse, error) {
	m.execOpts = opts
	return containertypes.ExecCreateResponse{ID: "exec-123"}, nil
}

func (m *mockDockerClient) ContainerExecAttach(_ context.Context, _ string, _ containertypes.ExecAttachOptions) (types.HijackedResponse, error) {
	// stdcopy format: 8-byte header + payload per frame
	// header: [stream_type, 0, 0, 0, size(4 bytes big-endian)]
	// stream_type: 1=stdout, 2=stderr
	output := m.execOutput
	size := uint32(len(output))
	header := []byte{
		1, 0, 0, 0,
		byte(size >> 24), byte(size >> 16), byte(size >> 8), byte(size),
	}
	data := append(header, []byte(output)...)

	conn := &mockConn{data: data}
	return types.NewHijackedResponse(conn, "application/vnd.docker.raw-stream"), nil
}

func (m *mockDockerClient) ContainerExecInspect(_ context.Context, _ string) (containertypes.ExecInspect, error) {
	return containertypes.ExecInspect{ExitCode: m.execExitCode}, nil
}

func testConfig() container.TeamContainerConfig {
	return container.TeamContainerConfig{
		ProjectID:       "proj-123",
		RepoURL:         "https://github.com/test/repo",
		AnthropicKey:    "sk-test-key",
		RabbitMQURL:     "amqp://guest:guest@rabbitmq:5672/",
		GitToken:        "ghp_test",
		NetworkName:     "foundry-net",
		SharedVolPath:   "/home/user/foundry/projects/test/shared",
		AgentLibPath:    "/home/user/agents",
		SSHKeyPath:      "/home/user/.ssh",
		TeamComposition: []string{"backend-developer", "frontend-developer"},
		ClaudeVersion:   "latest",
	}
}

func TestManager_CreateTeam_ContainerConfig(t *testing.T) {
	mock := &mockDockerClient{
		createResp: containertypes.CreateResponse{ID: "container-abc"},
	}
	mgr := container.NewManager(mock, "foundry-team:test")

	id, err := mgr.CreateTeam(context.Background(), testConfig())
	if err != nil {
		t.Fatalf("CreateTeam() error: %v", err)
	}
	if id != "container-abc" {
		t.Errorf("ID = %q, want %q", id, "container-abc")
	}

	// Verify container name
	if mock.createName != "foundry-team-proj-123" {
		t.Errorf("name = %q, want %q", mock.createName, "foundry-team-proj-123")
	}

	// Verify image
	if mock.createConfig.Image != "foundry-team:test" {
		t.Errorf("image = %q, want %q", mock.createConfig.Image, "foundry-team:test")
	}

	// Verify env vars
	envMap := make(map[string]string)
	for _, e := range mock.createConfig.Env {
		parts := strings.SplitN(e, "=", 2)
		envMap[parts[0]] = parts[1]
	}

	expectedEnv := map[string]string{
		"ANTHROPIC_API_KEY": "sk-test-key",
		"PROJECT_ID":        "proj-123",
		"REPO_URL":          "https://github.com/test/repo",
		"CLAUDE_VERSION":    "latest",
	}
	for k, v := range expectedEnv {
		if envMap[k] != v {
			t.Errorf("env %s = %q, want %q", k, envMap[k], v)
		}
	}

	// Verify labels
	if mock.createConfig.Labels["foundry.project_id"] != "proj-123" {
		t.Errorf("label foundry.project_id = %q", mock.createConfig.Labels["foundry.project_id"])
	}
	if mock.createConfig.Labels["foundry.type"] != "team" {
		t.Errorf("label foundry.type = %q", mock.createConfig.Labels["foundry.type"])
	}
}

func TestManager_CreateTeam_Mounts(t *testing.T) {
	mock := &mockDockerClient{
		createResp: containertypes.CreateResponse{ID: "c-1"},
	}
	mgr := container.NewManager(mock, "")

	_, _ = mgr.CreateTeam(context.Background(), testConfig())

	mounts := mock.createHostConfig.Mounts

	tests := []struct {
		name     string
		target   string
		typ      mount.Type
		readOnly bool
	}{
		{"workspace volume", "/workspace", mount.TypeVolume, false},
		{"shared bind", "/shared", mount.TypeBind, false},
		{"agents bind (ro)", "/agents", mount.TypeBind, true},
		{"ssh bind (ro)", "/root/.ssh", mount.TypeBind, true},
	}

	if len(mounts) != len(tests) {
		t.Fatalf("got %d mounts, want %d", len(mounts), len(tests))
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mounts[i]
			if m.Target != tt.target {
				t.Errorf("target = %q, want %q", m.Target, tt.target)
			}
			if m.Type != tt.typ {
				t.Errorf("type = %q, want %q", m.Type, tt.typ)
			}
			if m.ReadOnly != tt.readOnly {
				t.Errorf("readOnly = %v, want %v", m.ReadOnly, tt.readOnly)
			}
		})
	}
}

func TestManager_CreateTeam_Network(t *testing.T) {
	mock := &mockDockerClient{
		createResp: containertypes.CreateResponse{ID: "c-1"},
	}
	mgr := container.NewManager(mock, "")

	_, _ = mgr.CreateTeam(context.Background(), testConfig())

	if mock.createNetwork == nil {
		t.Fatal("expected network config")
	}
	if _, ok := mock.createNetwork.EndpointsConfig["foundry-net"]; !ok {
		t.Error("expected foundry-net in endpoints config")
	}
}

func TestManager_StartTeam(t *testing.T) {
	mock := &mockDockerClient{}
	mgr := container.NewManager(mock, "")

	err := mgr.StartTeam(context.Background(), "c-123")
	if err != nil {
		t.Fatalf("StartTeam() error: %v", err)
	}
	if mock.startID != "c-123" {
		t.Errorf("startID = %q, want %q", mock.startID, "c-123")
	}
}

func TestManager_StopTeam(t *testing.T) {
	mock := &mockDockerClient{}
	mgr := container.NewManager(mock, "")

	err := mgr.StopTeam(context.Background(), "c-123")
	if err != nil {
		t.Fatalf("StopTeam() error: %v", err)
	}
	if mock.stopID != "c-123" {
		t.Errorf("stopID = %q, want %q", mock.stopID, "c-123")
	}
	if mock.stopOpts.Timeout == nil || *mock.stopOpts.Timeout != 30 {
		t.Error("expected 30s timeout")
	}
}

func TestManager_RemoveTeam(t *testing.T) {
	mock := &mockDockerClient{}
	mgr := container.NewManager(mock, "")

	err := mgr.RemoveTeam(context.Background(), "c-123")
	if err != nil {
		t.Fatalf("RemoveTeam() error: %v", err)
	}
	if mock.removeID != "c-123" {
		t.Errorf("removeID = %q, want %q", mock.removeID, "c-123")
	}
}

func TestManager_GetStatus(t *testing.T) {
	mock := &mockDockerClient{
		inspectResp: containertypes.InspectResponse{
			ContainerJSONBase: &containertypes.ContainerJSONBase{
				ID: "c-123",
				State: &containertypes.State{
					Status:    "running",
					Running:   true,
					StartedAt: "2026-03-23T20:00:00.000000000Z",
				},
			},
		},
	}
	mgr := container.NewManager(mock, "")

	status, err := mgr.GetStatus(context.Background(), "c-123")
	if err != nil {
		t.Fatalf("GetStatus() error: %v", err)
	}

	if status.State != "running" {
		t.Errorf("State = %q, want %q", status.State, "running")
	}
	if status.Health != "running" {
		t.Errorf("Health = %q, want %q", status.Health, "running")
	}
}

func TestManager_GetStatus_WithDockerHealthCheck(t *testing.T) {
	mock := &mockDockerClient{
		inspectResp: containertypes.InspectResponse{
			ContainerJSONBase: &containertypes.ContainerJSONBase{
				ID: "c-123",
				State: &containertypes.State{
					Status:    "running",
					Running:   true,
					StartedAt: "2026-03-23T20:00:00.000000000Z",
					Health: &containertypes.Health{
						Status: "healthy",
					},
				},
			},
		},
	}
	mgr := container.NewManager(mock, "")

	status, err := mgr.GetStatus(context.Background(), "c-123")
	if err != nil {
		t.Fatalf("GetStatus() error: %v", err)
	}

	if status.Health != "healthy" {
		t.Errorf("Health = %q, want %q (should use docker healthcheck status)", status.Health, "healthy")
	}
}

func TestManager_GetStatus_StoppedContainer(t *testing.T) {
	mock := &mockDockerClient{
		inspectResp: containertypes.InspectResponse{
			ContainerJSONBase: &containertypes.ContainerJSONBase{
				ID: "c-123",
				State: &containertypes.State{
					Status:    "exited",
					Running:   false,
					StartedAt: "2026-03-23T20:00:00.000000000Z",
				},
			},
		},
	}
	mgr := container.NewManager(mock, "")

	status, err := mgr.GetStatus(context.Background(), "c-123")
	if err != nil {
		t.Fatalf("GetStatus() error: %v", err)
	}

	if status.State != "exited" {
		t.Errorf("State = %q, want %q", status.State, "exited")
	}
	if status.Health != "exited" {
		t.Errorf("Health = %q, want %q (stopped container reports state as health)", status.Health, "exited")
	}
}

func TestManager_StreamLogs(t *testing.T) {
	mock := &mockDockerClient{}
	mgr := container.NewManager(mock, "")

	reader, err := mgr.StreamLogs(context.Background(), "c-123")
	if err != nil {
		t.Fatalf("StreamLogs() error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading logs: %v", err)
	}

	if string(data) != "log output" {
		t.Errorf("logs = %q, want %q", string(data), "log output")
	}
}

func TestManager_ExecInTeam(t *testing.T) {
	mock := &mockDockerClient{
		execOutput:   "hello from exec",
		execExitCode: 0,
	}
	mgr := container.NewManager(mock, "")

	output, err := mgr.ExecInTeam(context.Background(), "c-123", []string{"echo", "hello from exec"})
	if err != nil {
		t.Fatalf("ExecInTeam() error: %v", err)
	}

	if output != "hello from exec" {
		t.Errorf("output = %q, want %q", output, "hello from exec")
	}

	if mock.execOpts.AttachStdout != true {
		t.Error("expected AttachStdout = true")
	}
	if mock.execOpts.AttachStderr != true {
		t.Error("expected AttachStderr = true")
	}
}

func TestManager_ExecInTeam_NonZeroExitCode(t *testing.T) {
	mock := &mockDockerClient{
		execOutput:   "command failed",
		execExitCode: 1,
	}
	mgr := container.NewManager(mock, "")

	_, err := mgr.ExecInTeam(context.Background(), "c-123", []string{"false"})
	if err == nil {
		t.Fatal("expected error for non-zero exit code")
	}
}

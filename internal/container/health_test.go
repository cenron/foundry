package container_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cenron/foundry/internal/container"
)

type mockExecutor struct {
	mu      sync.Mutex
	outputs map[string]string // containerID -> output
	errors  map[string]error
}

func (m *mockExecutor) ExecInTeam(_ context.Context, containerID string, _ []string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.errors[containerID]; ok {
		return "", err
	}
	return m.outputs[containerID], nil
}

type mockHealthUpdater struct {
	mu      sync.Mutex
	updates []healthUpdate
}

type healthUpdate struct {
	projectID string
	health    string
}

func (m *mockHealthUpdater) UpdateHealthByProject(_ context.Context, projectID, health string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updates = append(m.updates, healthUpdate{projectID, health})
	return nil
}

type mockPublisher struct {
	mu       sync.Mutex
	messages []publishedMsg
}

type publishedMsg struct {
	exchange   string
	routingKey string
}

func (m *mockPublisher) Publish(_ context.Context, exchange, routingKey string, _ []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, publishedMsg{exchange, routingKey})
	return nil
}

func newTestMonitor(executor *mockExecutor, updater *mockHealthUpdater, publisher *mockPublisher, now func() time.Time) *container.HealthMonitor {
	monitor := container.NewHealthMonitor(executor, updater, publisher)
	monitor.SetNowFunc(now)
	monitor.SetInterval(50 * time.Millisecond)
	monitor.SetThreshold(3)
	return monitor
}

func TestHealthMonitor_FreshBeatResetsCount(t *testing.T) {
	now := time.Unix(1000, 0)

	executor := &mockExecutor{
		outputs: map[string]string{"c-1": fmt.Sprintf("%d", now.Unix())},
	}
	updater := &mockHealthUpdater{}
	publisher := &mockPublisher{}

	monitor := newTestMonitor(executor, updater, publisher, func() time.Time { return now })
	monitor.Register("c-1", "proj-1")

	ctx, cancel := context.WithCancel(context.Background())
	go monitor.Start(ctx)

	time.Sleep(200 * time.Millisecond)
	cancel()

	updater.mu.Lock()
	defer updater.mu.Unlock()
	// No health updates should occur — everything is healthy
	if len(updater.updates) != 0 {
		t.Errorf("expected 0 health updates, got %d", len(updater.updates))
	}
}

func TestHealthMonitor_MissedBeatsMarkUnhealthy(t *testing.T) {
	now := time.Unix(1000, 0)

	executor := &mockExecutor{
		// Stale heartbeat — 5 minutes ago
		outputs: map[string]string{"c-1": fmt.Sprintf("%d", now.Add(-5*time.Minute).Unix())},
	}
	updater := &mockHealthUpdater{}
	publisher := &mockPublisher{}

	monitor := newTestMonitor(executor, updater, publisher, func() time.Time { return now })
	monitor.Register("c-1", "proj-1")

	ctx, cancel := context.WithCancel(context.Background())
	go monitor.Start(ctx)

	// Wait for at least 3 checks at 50ms interval
	time.Sleep(300 * time.Millisecond)
	cancel()

	updater.mu.Lock()
	defer updater.mu.Unlock()

	found := false
	for _, u := range updater.updates {
		if u.projectID == "proj-1" && u.health == "unhealthy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected unhealthy update for proj-1")
	}
}

func TestHealthMonitor_ExecErrorCountsAsMissed(t *testing.T) {
	now := time.Unix(1000, 0)

	executor := &mockExecutor{
		outputs: map[string]string{},
		errors:  map[string]error{"c-1": fmt.Errorf("connection refused")},
	}
	updater := &mockHealthUpdater{}
	publisher := &mockPublisher{}

	monitor := newTestMonitor(executor, updater, publisher, func() time.Time { return now })
	monitor.Register("c-1", "proj-1")

	ctx, cancel := context.WithCancel(context.Background())
	go monitor.Start(ctx)

	time.Sleep(300 * time.Millisecond)
	cancel()

	updater.mu.Lock()
	defer updater.mu.Unlock()

	found := false
	for _, u := range updater.updates {
		if u.projectID == "proj-1" && u.health == "unhealthy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected unhealthy update after exec errors")
	}
}

func TestHealthMonitor_RecoveryAfterHealthyBeat(t *testing.T) {
	now := time.Unix(1000, 0)
	staleTs := fmt.Sprintf("%d", now.Add(-5*time.Minute).Unix())
	freshTs := fmt.Sprintf("%d", now.Unix())

	executor := &mockExecutor{
		outputs: map[string]string{"c-1": staleTs},
	}
	updater := &mockHealthUpdater{}
	publisher := &mockPublisher{}

	monitor := newTestMonitor(executor, updater, publisher, func() time.Time { return now })
	monitor.Register("c-1", "proj-1")

	ctx, cancel := context.WithCancel(context.Background())
	go monitor.Start(ctx)

	// Wait for unhealthy
	time.Sleep(300 * time.Millisecond)

	// Now recover
	executor.mu.Lock()
	executor.outputs["c-1"] = freshTs
	executor.mu.Unlock()

	time.Sleep(200 * time.Millisecond)
	cancel()

	updater.mu.Lock()
	defer updater.mu.Unlock()

	// Should see both unhealthy and healthy updates
	hasUnhealthy := false
	hasHealthy := false
	for _, u := range updater.updates {
		if u.health == "unhealthy" {
			hasUnhealthy = true
		}
		if u.health == "healthy" {
			hasHealthy = true
		}
	}
	if !hasUnhealthy {
		t.Error("expected unhealthy update")
	}
	if !hasHealthy {
		t.Error("expected healthy recovery update")
	}
}

func TestHealthMonitor_DeregisteredNotChecked(t *testing.T) {
	now := time.Unix(1000, 0)

	executor := &mockExecutor{
		errors: map[string]error{"c-1": fmt.Errorf("should not be called")},
	}
	updater := &mockHealthUpdater{}
	publisher := &mockPublisher{}

	monitor := newTestMonitor(executor, updater, publisher, func() time.Time { return now })
	monitor.Register("c-1", "proj-1")
	monitor.Deregister("c-1")

	ctx, cancel := context.WithCancel(context.Background())
	go monitor.Start(ctx)

	time.Sleep(200 * time.Millisecond)
	cancel()

	updater.mu.Lock()
	defer updater.mu.Unlock()
	if len(updater.updates) != 0 {
		t.Errorf("expected 0 updates for deregistered container, got %d", len(updater.updates))
	}
}

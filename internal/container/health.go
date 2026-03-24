package container

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TeamExecutor executes commands inside a team container.
type TeamExecutor interface {
	ExecInTeam(ctx context.Context, containerID string, cmd []string) (string, error)
}

// AgentHealthUpdater updates agent health status in the data layer.
type AgentHealthUpdater interface {
	UpdateHealthByProject(ctx context.Context, projectID string, health string) error
}

// EventPublisher publishes events to the message broker.
type EventPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

type HealthMonitor struct {
	executor  TeamExecutor
	agents    AgentHealthUpdater
	publisher EventPublisher
	interval  time.Duration
	threshold int
	now       func() time.Time

	mu         sync.Mutex
	containers map[string]*containerHealth
}

type containerHealth struct {
	projectID   string
	missedBeats int
	healthy     bool
}

func NewHealthMonitor(executor TeamExecutor, agents AgentHealthUpdater, publisher EventPublisher) *HealthMonitor {
	return &HealthMonitor{
		executor:   executor,
		agents:     agents,
		publisher:  publisher,
		interval:   10 * time.Second,
		threshold:  3,
		now:        time.Now,
		containers: make(map[string]*containerHealth),
	}
}

func (h *HealthMonitor) SetNowFunc(fn func() time.Time)  { h.now = fn }
func (h *HealthMonitor) SetInterval(d time.Duration)     { h.interval = d }
func (h *HealthMonitor) SetThreshold(n int)              { h.threshold = n }

func (h *HealthMonitor) Register(containerID, projectID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.containers[containerID] = &containerHealth{
		projectID:   projectID,
		missedBeats: 0,
		healthy:     true,
	}
}

func (h *HealthMonitor) Deregister(containerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.containers, containerID)
}

// Start begins the monitoring loop. Blocks until ctx is cancelled.
func (h *HealthMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkAll(ctx)
		}
	}
}

func (h *HealthMonitor) checkAll(ctx context.Context) {
	h.mu.Lock()
	ids := make([]string, 0, len(h.containers))
	for id := range h.containers {
		ids = append(ids, id)
	}
	h.mu.Unlock()

	for _, id := range ids {
		h.checkOne(ctx, id)
	}
}

func (h *HealthMonitor) checkOne(ctx context.Context, containerID string) {
	h.mu.Lock()
	ch, ok := h.containers[containerID]
	if !ok {
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	output, err := h.executor.ExecInTeam(ctx, containerID, []string{"cat", "/foundry/state/heartbeat"})

	beatFresh := false
	if err == nil {
		beatFresh = h.isFreshBeat(output)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if beatFresh {
		if !ch.healthy {
			h.markHealthy(ctx, ch)
		}
		ch.missedBeats = 0
		ch.healthy = true
		return
	}

	ch.missedBeats++
	if ch.missedBeats >= h.threshold && ch.healthy {
		h.markUnhealthy(ctx, ch)
		ch.healthy = false
	}
}

func (h *HealthMonitor) isFreshBeat(output string) bool {
	ts, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
	if err != nil {
		return false
	}

	beatTime := time.Unix(ts, 0)
	return h.now().Sub(beatTime) < h.interval*2
}

func (h *HealthMonitor) markUnhealthy(ctx context.Context, ch *containerHealth) {
	if err := h.agents.UpdateHealthByProject(ctx, ch.projectID, "unhealthy"); err != nil {
		log.Printf("health monitor: updating agents unhealthy: %v", err)
	}

	h.publishHealthEvent(ctx, ch.projectID, "container_unhealthy")
}

func (h *HealthMonitor) markHealthy(ctx context.Context, ch *containerHealth) {
	if err := h.agents.UpdateHealthByProject(ctx, ch.projectID, "healthy"); err != nil {
		log.Printf("health monitor: updating agents healthy: %v", err)
	}

	h.publishHealthEvent(ctx, ch.projectID, "container_healthy")
}

func (h *HealthMonitor) publishHealthEvent(ctx context.Context, projectID, eventType string) {
	payload, _ := json.Marshal(map[string]string{
		"project_id": projectID,
		"type":       eventType,
	})

	routingKey := fmt.Sprintf("events.%s.%s", projectID, eventType)
	if err := h.publisher.Publish(ctx, "foundry.events", routingKey, payload); err != nil {
		log.Printf("health monitor: publishing %s event: %v", eventType, err)
	}
}

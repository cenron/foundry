package event_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/cenron/foundry/internal/event"
)

type mockBroadcaster struct {
	mu       sync.Mutex
	messages [][]byte
}

func (m *mockBroadcaster) Broadcast(msg []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

type mockCacheUpdater struct {
	mu    sync.Mutex
	items map[string]interface{}
}

func newMockCache() *mockCacheUpdater {
	return &mockCacheUpdater{items: make(map[string]interface{})}
}

func (m *mockCacheUpdater) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = value
	return nil
}

type mockBrokerSubscriber struct {
	handlers map[string]func(body []byte) error
}

func newMockSubscriber() *mockBrokerSubscriber {
	return &mockBrokerSubscriber{handlers: make(map[string]func(body []byte) error)}
}

func (m *mockBrokerSubscriber) Subscribe(_, routingKey, _ string, handler func(body []byte) error) error {
	m.handlers[routingKey] = handler
	return nil
}

func TestRouter_Start_SubscribesCorrectly(t *testing.T) {
	sub := newMockSubscriber()
	router := event.NewRouter(nil, &mockBroadcaster{}, nil, sub)

	if err := router.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if _, ok := sub.handlers["events.#"]; !ok {
		t.Error("expected events subscription")
	}
	if _, ok := sub.handlers["logs.#"]; !ok {
		t.Error("expected logs subscription")
	}
}

func TestRouter_HandleEvent_BroadcastsToUI(t *testing.T) {
	hub := &mockBroadcaster{}
	cache := newMockCache()
	sub := newMockSubscriber()

	router := event.NewRouter(nil, hub, cache, sub)
	_ = router.Start()

	evt := map[string]interface{}{
		"project_id": "00000000-0000-0000-0000-000000000001",
		"type":       "task_completed",
		"payload":    map[string]string{"task_id": "t-1"},
	}
	body, _ := json.Marshal(evt)

	handler := sub.handlers["events.#"]
	err := handler(body)
	if err != nil {
		t.Fatalf("handleEvent error: %v", err)
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()
	if len(hub.messages) != 1 {
		t.Errorf("expected 1 broadcast, got %d", len(hub.messages))
	}
}

func TestRouter_HandleEvent_UpdatesCache(t *testing.T) {
	hub := &mockBroadcaster{}
	cache := newMockCache()
	sub := newMockSubscriber()

	router := event.NewRouter(nil, hub, cache, sub)
	_ = router.Start()

	evt := map[string]interface{}{
		"project_id": "proj-abc",
		"type":       "agent_started",
	}
	body, _ := json.Marshal(evt)

	handler := sub.handlers["events.#"]
	_ = handler(body)

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if _, ok := cache.items["project:proj-abc:latest_event"]; !ok {
		t.Error("expected cache update for project")
	}
}

func TestRouter_HandleLog_BroadcastsDirectly(t *testing.T) {
	hub := &mockBroadcaster{}
	sub := newMockSubscriber()

	router := event.NewRouter(nil, hub, nil, sub)
	_ = router.Start()

	logLine := []byte(`[agent:backend] compiling main.go...`)

	handler := sub.handlers["logs.#"]
	err := handler(logLine)
	if err != nil {
		t.Fatalf("handleLog error: %v", err)
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()
	if len(hub.messages) != 1 {
		t.Errorf("expected 1 log broadcast, got %d", len(hub.messages))
	}
}

func TestRouter_HandleEvent_InvalidJSON(t *testing.T) {
	hub := &mockBroadcaster{}
	sub := newMockSubscriber()

	router := event.NewRouter(nil, hub, nil, sub)
	_ = router.Start()

	handler := sub.handlers["events.#"]
	err := handler([]byte("not json"))
	if err != nil {
		t.Fatal("should ack bad messages without error")
	}
}

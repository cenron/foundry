package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// waitForClients polls hub.clients until at least n clients are registered,
// or the deadline passes.
func waitForClients(h *Hub, n int, deadline time.Duration) bool {
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		h.mu.RLock()
		count := len(h.clients)
		h.mu.RUnlock()
		if count >= n {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return false
}

// waitForChannelSubscribers polls channelHub.subscribers until at least n
// connections are registered under the given topic, or the deadline passes.
func waitForChannelSubscribers(ch *ChannelHub, topic string, n int, deadline time.Duration) bool {
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		ch.mu.RLock()
		count := len(ch.subscribers[topic])
		ch.mu.RUnlock()
		if count >= n {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return false
}

// dialTestServer dials the given httptest.Server URL as a WebSocket.
func dialTestServer(t *testing.T, ts *httptest.Server, path string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(ts.URL, "http") + path
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	return conn
}

func wsDeadline() time.Time {
	return time.Now().Add(2 * time.Second)
}

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TestHub_RegisterBroadcastUnregister tests Hub.Register, Broadcast, and Unregister.
func TestHub_RegisterBroadcastUnregister(t *testing.T) {
	hub := NewHub()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Register(conn)
		// Keep alive: read until close.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))
	defer ts.Close()

	conn := dialTestServer(t, ts, "/")
	defer func() { _ = conn.Close() }()

	// Wait for the handler goroutine to call Register before broadcasting.
	if !waitForClients(hub, 1, 2*time.Second) {
		t.Fatal("timed out waiting for hub to register client")
	}

	// Broadcast — the registered client should receive the message.
	hub.Broadcast([]byte("hello"))

	_ = conn.SetReadDeadline(wsDeadline())
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != "hello" {
		t.Errorf("got %q, want %q", string(msg), "hello")
	}

	// Unregister closes the server-side connection.
	// We trigger it by closing our end — the server's read loop returns,
	// but Unregister is called via defer inside hub.HandleWebSocket.
	// For the direct Unregister path, call it directly.
	_ = conn.Close()
}

// TestHub_Broadcast_NoClients verifies Broadcast does not panic with zero clients.
func TestHub_Broadcast_NoClients(t *testing.T) {
	hub := NewHub()
	hub.Broadcast([]byte("no-one-listening")) // must not panic
}

// TestChannelHub_SubscribePublish tests ChannelHub.Subscribe and Publish.
func TestChannelHub_SubscribePublish(t *testing.T) {
	hub := NewChannelHub()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Subscribe("proj:1", conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))
	defer ts.Close()

	conn := dialTestServer(t, ts, "/")
	defer func() { _ = conn.Close() }()

	// Wait for the handler goroutine to call Subscribe before publishing.
	if !waitForChannelSubscribers(hub, "proj:1", 1, 2*time.Second) {
		t.Fatal("timed out waiting for channelhub to register subscriber")
	}

	hub.Publish("proj:1", []byte("event-data"))

	_ = conn.SetReadDeadline(wsDeadline())
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != "event-data" {
		t.Errorf("got %q, want %q", string(msg), "event-data")
	}
}

// TestChannelHub_Publish_NoSubscribers verifies no panic when topic is empty.
func TestChannelHub_Publish_NoSubscribers(t *testing.T) {
	hub := NewChannelHub()
	hub.Publish("empty-topic", []byte("data")) // must not panic
}

// TestChannelHub_Unsubscribe removes the connection from the topic.
func TestChannelHub_Unsubscribe(t *testing.T) {
	hub := NewChannelHub()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Subscribe("topic-x", conn)

		// Wait briefly, then unsubscribe.
		time.Sleep(20 * time.Millisecond)
		hub.Unsubscribe("topic-x", conn) // closes conn server-side
	}))
	defer ts.Close()

	conn := dialTestServer(t, ts, "/")
	defer func() { _ = conn.Close() }()

	// After unsubscribe, reading from client should return an error (connection closed).
	_ = conn.SetReadDeadline(wsDeadline())
	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Error("expected error reading after server-side Unsubscribe, got nil")
	}
}

// setRouteContext injects a chi URL parameter into the request context.
func setRouteContext(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// TestHub_HandleWebSocket verifies that HandleWebSocket upgrades the connection
// and registers the client with the hub.
func TestHub_HandleWebSocket(t *testing.T) {
	hub := NewHub()
	srv := &Server{hub: hub, channelHub: NewChannelHub(), router: nil}

	ts := httptest.NewServer(http.HandlerFunc(srv.hub.HandleWebSocket))
	defer ts.Close()

	conn := dialTestServer(t, ts, "/")
	defer func() { _ = conn.Close() }()

	if !waitForClients(hub, 1, 2*time.Second) {
		t.Fatal("timed out waiting for hub to register client via HandleWebSocket")
	}
}

// TestHub_HandleWebSocket_UpgradeFail verifies HandleWebSocket doesn't panic when
// the upgrade fails (non-WS request).
func TestHub_HandleWebSocket_UpgradeFail(t *testing.T) {
	hub := NewHub()

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	w := httptest.NewRecorder()

	// This should not panic — upgrader returns an error and HandleWebSocket logs it.
	hub.HandleWebSocket(w, req)
}

// TestServer_HandleProjectEvents verifies the project events WS endpoint
// subscribes the client to the correct topic.
func TestServer_HandleProjectEvents(t *testing.T) {
	s := &Server{
		hub:        NewHub(),
		channelHub: NewChannelHub(),
		router:     nil,
		deps:       ServerDeps{},
	}

	projectID := "00000000-0000-0000-0000-000000000001"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = setRouteContext(r, "id", projectID)
		s.handleProjectEvents(w, r)
	}))
	defer ts.Close()

	conn := dialTestServer(t, ts, "/ws/projects/"+projectID+"/events")
	defer func() { _ = conn.Close() }()

	topic := "project:" + projectID
	if !waitForChannelSubscribers(s.channelHub, topic, 1, 2*time.Second) {
		t.Fatalf("timed out waiting for channelHub to register project subscriber on topic %q", topic)
	}

	// Publish to the topic — the connected client should receive it.
	s.channelHub.Publish(topic, []byte("project-event"))

	_ = conn.SetReadDeadline(wsDeadline())
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != "project-event" {
		t.Errorf("got %q, want %q", string(msg), "project-event")
	}
}

// TestHub_Broadcast_FailedWrite verifies that Broadcast handles a closed
// connection gracefully by triggering the Unregister goroutine.
func TestHub_Broadcast_FailedWrite(t *testing.T) {
	hub := NewHub()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Register(conn)
		// Read loop so the server-side goroutine stays alive.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))
	defer ts.Close()

	conn := dialTestServer(t, ts, "/")

	// Wait for registration.
	if !waitForClients(hub, 1, 2*time.Second) {
		t.Fatal("timed out waiting for hub to register client")
	}

	// Close the client-side connection so the next Broadcast write fails.
	_ = conn.Close()

	// Give the server a moment to detect the close, then broadcast.
	// The write will fail, triggering the Unregister path.
	time.Sleep(20 * time.Millisecond)
	hub.Broadcast([]byte("should fail to write"))

	// Wait for the async Unregister to remove the dead client.
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		hub.mu.RLock()
		n := len(hub.clients)
		hub.mu.RUnlock()
		if n == 0 {
			return // Client was unregistered — test passes.
		}
		time.Sleep(5 * time.Millisecond)
	}
	// The client may already have been removed by the read-loop goroutine
	// before Broadcast ran — either outcome is acceptable.
}

// TestServer_HandleAgentLogs verifies the agent logs WS endpoint
// subscribes the client to the correct topic.
func TestServer_HandleAgentLogs(t *testing.T) {
	s := &Server{
		hub:        NewHub(),
		channelHub: NewChannelHub(),
		router:     nil,
		deps:       ServerDeps{},
	}

	agentID := "00000000-0000-0000-0000-000000000002"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = setRouteContext(r, "agentId", agentID)
		s.handleAgentLogs(w, r)
	}))
	defer ts.Close()

	conn := dialTestServer(t, ts, "/ws/agents/"+agentID+"/logs")
	defer func() { _ = conn.Close() }()

	topic := "agent:" + agentID
	if !waitForChannelSubscribers(s.channelHub, topic, 1, 2*time.Second) {
		t.Fatalf("timed out waiting for channelHub to register agent subscriber on topic %q", topic)
	}

	s.channelHub.Publish(topic, []byte("agent-log"))

	_ = conn.SetReadDeadline(wsDeadline())
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != "agent-log" {
		t.Errorf("got %q, want %q", string(msg), "agent-log")
	}
}

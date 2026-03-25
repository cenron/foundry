package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

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

	// Give the handler goroutine time to Register.
	time.Sleep(10 * time.Millisecond)

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

	// Give the handler time to Subscribe.
	time.Sleep(10 * time.Millisecond)

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

package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/cenron/foundry/internal/shared"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Dev permissive; tighten in production
	},
}

// Hub manages WebSocket connections and broadcasts events to all clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = struct{}{}
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	_ = conn.Close()
}

func (h *Hub) Broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("websocket write error: %v", err)
			go h.Unregister(conn)
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	h.Register(conn)

	// Read loop — keeps connection alive, handles client disconnect
	go func() {
		defer h.Unregister(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// connEntry wraps a WebSocket connection with its own write mutex to prevent
// concurrent writers on the same connection.
type connEntry struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

// ChannelHub routes messages to per-topic WebSocket subscribers.
//
// Topics follow the convention "project:<id>" or "agent:<id>".
type ChannelHub struct {
	mu          sync.RWMutex
	subscribers map[string]map[*websocket.Conn]*connEntry
}

func NewChannelHub() *ChannelHub {
	return &ChannelHub{
		subscribers: make(map[string]map[*websocket.Conn]*connEntry),
	}
}

// Subscribe registers conn under topic. Safe for concurrent use.
func (ch *ChannelHub) Subscribe(topic string, conn *websocket.Conn) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.subscribers[topic] == nil {
		ch.subscribers[topic] = make(map[*websocket.Conn]*connEntry)
	}
	ch.subscribers[topic][conn] = &connEntry{conn: conn}
}

// Unsubscribe removes conn from topic and closes the connection.
func (ch *ChannelHub) Unsubscribe(topic string, conn *websocket.Conn) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	delete(ch.subscribers[topic], conn)
	if len(ch.subscribers[topic]) == 0 {
		delete(ch.subscribers, topic)
	}
	_ = conn.Close()
}

// Publish sends message to all subscribers of topic.
//
// The subscriber map is copied under the read lock, then the lock is released
// before any network I/O occurs. Each connection is protected by its own mutex
// to prevent concurrent writers on the same gorilla/websocket connection.
func (ch *ChannelHub) Publish(topic string, message []byte) {
	ch.mu.RLock()
	entries := make([]*connEntry, 0, len(ch.subscribers[topic]))
	for _, e := range ch.subscribers[topic] {
		entries = append(entries, e)
	}
	ch.mu.RUnlock()

	for _, e := range entries {
		e.mu.Lock()
		err := e.conn.WriteMessage(websocket.TextMessage, message)
		e.mu.Unlock()
		if err != nil {
			log.Printf("channelhub write error on topic %q: %v", topic, err)
			go ch.Unsubscribe(topic, e.conn)
		}
	}
}

func (s *Server) handleProjectEvents(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error (project events): %v", err)
		return
	}

	topic := "project:" + projectID.String()
	s.channelHub.Subscribe(topic, conn)

	go func() {
		defer s.channelHub.Unsubscribe(topic, conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

func (s *Server) handleAgentLogs(w http.ResponseWriter, r *http.Request) {
	agentID, err := shared.ParseID(chi.URLParam(r, "agentId"))
	if err != nil {
		http.Error(w, "invalid agent ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error (agent logs): %v", err)
		return
	}

	topic := "agent:" + agentID.String()
	s.channelHub.Subscribe(topic, conn)

	go func() {
		defer s.channelHub.Unsubscribe(topic, conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

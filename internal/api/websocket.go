package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
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

// ChannelHub routes messages to per-topic WebSocket subscribers.
//
// Topics follow the convention "project:<id>" or "agent:<id>".
type ChannelHub struct {
	mu          sync.RWMutex
	subscribers map[string]map[*websocket.Conn]struct{}
}

func NewChannelHub() *ChannelHub {
	return &ChannelHub{
		subscribers: make(map[string]map[*websocket.Conn]struct{}),
	}
}

// Subscribe registers conn under topic. Safe for concurrent use.
func (ch *ChannelHub) Subscribe(topic string, conn *websocket.Conn) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.subscribers[topic] == nil {
		ch.subscribers[topic] = make(map[*websocket.Conn]struct{})
	}
	ch.subscribers[topic][conn] = struct{}{}
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
func (ch *ChannelHub) Publish(topic string, message []byte) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	for conn := range ch.subscribers[topic] {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("channelhub write error on topic %q: %v", topic, err)
			go ch.Unsubscribe(topic, conn)
		}
	}
}

func (s *Server) handleProjectEvents(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error (project events): %v", err)
		return
	}

	topic := "project:" + projectID
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
	agentID := chi.URLParam(r, "agentId")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error (agent logs): %v", err)
		return
	}

	topic := "agent:" + agentID
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

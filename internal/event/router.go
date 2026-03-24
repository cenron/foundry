package event

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cenron/foundry/internal/shared"
)

// Broadcaster sends messages to connected WebSocket clients.
type Broadcaster interface {
	Broadcast(message []byte)
}

// CacheUpdater updates cached state in Redis.
type CacheUpdater interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// BrokerSubscriber subscribes to message broker topics.
type BrokerSubscriber interface {
	Subscribe(exchange, routingKey, queueName string, handler func(body []byte) error) error
}

type Router struct {
	store      *Store
	hub        Broadcaster
	cache      CacheUpdater
	subscriber BrokerSubscriber
}

func NewRouter(store *Store, hub Broadcaster, cache CacheUpdater, subscriber BrokerSubscriber) *Router {
	return &Router{
		store:      store,
		hub:        hub,
		cache:      cache,
		subscriber: subscriber,
	}
}

// Start subscribes to RabbitMQ exchanges and begins routing events.
func (r *Router) Start() error {
	if err := r.subscriber.Subscribe("foundry.events", "events.#", "event-router-events", r.handleEvent); err != nil {
		return err
	}

	if err := r.subscriber.Subscribe("foundry.logs", "logs.#", "event-router-logs", r.handleLog); err != nil {
		return err
	}

	log.Println("event router: subscribed to events and logs")
	return nil
}

func (r *Router) handleEvent(body []byte) error {
	var envelope eventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		log.Printf("event router: invalid event JSON: %v", err)
		return nil // ack bad messages to avoid requeue loop
	}

	r.persistEvent(envelope)
	r.updateCache(envelope)
	r.broadcastToUI(body)

	return nil
}

func (r *Router) handleLog(body []byte) error {
	// Forward log lines directly to WebSocket — no persistence by default
	r.hub.Broadcast(body)
	return nil
}

func (r *Router) persistEvent(env eventEnvelope) {
	if r.store == nil {
		return
	}

	ctx := context.Background()

	var taskID *shared.ID
	if env.TaskID != "" {
		id, err := shared.ParseID(env.TaskID)
		if err == nil {
			taskID = &id
		}
	}

	var agentID *shared.ID
	if env.AgentID != "" {
		id, err := shared.ParseID(env.AgentID)
		if err == nil {
			agentID = &id
		}
	}

	projectID, err := shared.ParseID(env.ProjectID)
	if err != nil {
		log.Printf("event router: invalid project_id %q: %v", env.ProjectID, err)
		return
	}

	_, err = r.store.Create(ctx, CreateEventParams{
		ProjectID: projectID,
		TaskID:    taskID,
		AgentID:   agentID,
		Type:      env.Type,
		Payload:   env.Payload,
	})
	if err != nil {
		log.Printf("event router: persisting event: %v", err)
	}
}

func (r *Router) updateCache(env eventEnvelope) {
	if r.cache == nil {
		return
	}

	ctx := context.Background()

	// Cache latest event per project for quick dashboard reads
	cacheKey := "project:" + env.ProjectID + ":latest_event"
	if err := r.cache.Set(ctx, cacheKey, env, 5*time.Minute); err != nil {
		log.Printf("event router: cache update: %v", err)
	}
}

func (r *Router) broadcastToUI(body []byte) {
	r.hub.Broadcast(body)
}

// eventEnvelope is the standard shape of events flowing through RabbitMQ.
type eventEnvelope struct {
	ProjectID string          `json:"project_id"`
	TaskID    string          `json:"task_id,omitempty"`
	AgentID   string          `json:"agent_id,omitempty"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

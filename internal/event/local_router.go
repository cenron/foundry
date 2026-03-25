package event

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/cenron/foundry/internal/shared"
)

// LocalRouter reads stream-json output from a local agent process, persists
// events via the Store, and broadcasts them to WebSocket clients.
type LocalRouter struct {
	store Storer
	hub   Broadcaster
	cache CacheUpdater
}

// Storer persists events. Satisfied by *Store.
type Storer interface {
	Create(ctx context.Context, params CreateEventParams) (*Event, error)
}

// NewLocalRouter constructs a LocalRouter.
func NewLocalRouter(store Storer, hub Broadcaster, cache CacheUpdater) *LocalRouter {
	return &LocalRouter{
		store: store,
		hub:   hub,
		cache: cache,
	}
}

// StreamAgentOutput reads newline-delimited stream-json from the given reader
// until EOF or the context is cancelled, forwarding each line to ForwardLogLine.
func (lr *LocalRouter) StreamAgentOutput(ctx context.Context, projectID, agentID string, reader interface {
	Read(p []byte) (n int, err error)
}) {
	scanner := bufio.NewScanner(reader)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !scanner.Scan() {
			return
		}

		line := scanner.Text()
		lr.ForwardLogLine(ctx, projectID, agentID, line)
	}
}

// ForwardLogLine processes a single stream-json line from an agent.
// Empty lines are silently ignored.
// Lines that are not valid JSON are logged and skipped.
// The event type is extracted from the "type" field; if absent, defaults to "agent.output".
func (lr *LocalRouter) ForwardLogLine(ctx context.Context, projectID, agentID, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		log.Printf("local router: agent %s produced non-JSON line: %v", agentID, err)
		return
	}

	eventType := "agent.output"
	if t, ok := payload["type"].(string); ok && t != "" {
		eventType = t
	}

	envelope := eventEnvelope{
		ProjectID: projectID,
		AgentID:   agentID,
		Type:      eventType,
		Payload:   json.RawMessage(line),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	raw, err := json.Marshal(envelope)
	if err != nil {
		log.Printf("local router: marshaling envelope: %v", err)
		return
	}

	if lr.store != nil {
		lr.persistEnvelope(ctx, envelope)
	}

	lr.hub.Broadcast(raw)
}

func (lr *LocalRouter) persistEnvelope(ctx context.Context, env eventEnvelope) {
	projectID, err := shared.ParseID(env.ProjectID)
	if err != nil {
		log.Printf("local router: invalid project_id %q: %v", env.ProjectID, err)
		return
	}

	params := CreateEventParams{
		ProjectID: projectID,
		Type:      env.Type,
		Payload:   env.Payload,
	}

	if env.AgentID != "" {
		agentID, err := shared.ParseID(env.AgentID)
		if err == nil {
			params.AgentID = &agentID
		}
	}

	if _, err := lr.store.Create(ctx, params); err != nil {
		log.Printf("local router: persisting event: %v", err)
	}
}

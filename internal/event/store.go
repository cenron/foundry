package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cenron/foundry/internal/shared"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, params CreateEventParams) (*Event, error) {
	payloadJSON, err := json.Marshal(params.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling event payload: %w", err)
	}

	var e Event
	err = s.db.QueryRowxContext(ctx, `
		INSERT INTO events (project_id, task_id, agent_id, type, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *`,
		params.ProjectID, params.TaskID, params.AgentID, params.Type, payloadJSON,
	).StructScan(&e)
	if err != nil {
		return nil, fmt.Errorf("creating event: %w", err)
	}

	return &e, nil
}

func (s *Store) ListByProject(ctx context.Context, projectID shared.ID, page, pageSize int) ([]Event, int, error) {
	p := shared.PaginationParams{Page: page, PageSize: pageSize}

	var total int
	if err := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM events WHERE project_id = $1", projectID); err != nil {
		return nil, 0, fmt.Errorf("counting events: %w", err)
	}

	var events []Event
	err := s.db.SelectContext(ctx, &events,
		"SELECT * FROM events WHERE project_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		projectID, p.Limit(), p.Offset(),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing events: %w", err)
	}

	return events, total, nil
}

func (s *Store) ListByAgent(ctx context.Context, agentID shared.ID) ([]Event, error) {
	var events []Event
	err := s.db.SelectContext(ctx, &events,
		"SELECT * FROM events WHERE agent_id = $1 ORDER BY created_at DESC",
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing events for agent %s: %w", agentID, err)
	}
	return events, nil
}

type ArtifactStore struct {
	db *sqlx.DB
}

func NewArtifactStore(db *sqlx.DB) *ArtifactStore {
	return &ArtifactStore{db: db}
}

func (s *ArtifactStore) Create(ctx context.Context, params CreateArtifactParams) (*Artifact, error) {
	var a Artifact
	err := s.db.QueryRowxContext(ctx, `
		INSERT INTO artifacts (project_id, task_id, agent_id, type, path, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *`,
		params.ProjectID, params.TaskID, params.AgentID,
		params.Type, params.Path, params.Description,
	).StructScan(&a)
	if err != nil {
		return nil, fmt.Errorf("creating artifact: %w", err)
	}
	return &a, nil
}

func (s *ArtifactStore) ListByProject(ctx context.Context, projectID shared.ID) ([]Artifact, error) {
	var artifacts []Artifact
	err := s.db.SelectContext(ctx, &artifacts,
		"SELECT * FROM artifacts WHERE project_id = $1 ORDER BY created_at DESC",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing artifacts: %w", err)
	}
	return artifacts, nil
}

func (s *ArtifactStore) ListByTask(ctx context.Context, taskID shared.ID) ([]Artifact, error) {
	var artifacts []Artifact
	err := s.db.SelectContext(ctx, &artifacts,
		"SELECT * FROM artifacts WHERE task_id = $1 ORDER BY created_at DESC",
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing artifacts by task: %w", err)
	}
	return artifacts, nil
}

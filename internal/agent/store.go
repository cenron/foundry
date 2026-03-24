package agent

import (
	"context"
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

func (s *Store) Create(ctx context.Context, params CreateAgentParams) (*Agent, error) {
	var a Agent
	err := s.db.QueryRowxContext(ctx, `
		INSERT INTO agents (project_id, role, provider, container_id)
		VALUES ($1, $2, $3, $4)
		RETURNING *`,
		params.ProjectID, params.Role, params.Provider, params.ContainerID,
	).StructScan(&a)
	if err != nil {
		return nil, fmt.Errorf("creating agent: %w", err)
	}
	return &a, nil
}

func (s *Store) GetByID(ctx context.Context, id shared.ID) (*Agent, error) {
	var a Agent
	err := s.db.GetContext(ctx, &a, "SELECT * FROM agents WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("getting agent %s: %w", id, err)
	}
	return &a, nil
}

func (s *Store) ListByProject(ctx context.Context, projectID shared.ID) ([]Agent, error) {
	var agents []Agent
	err := s.db.SelectContext(ctx, &agents,
		"SELECT * FROM agents WHERE project_id = $1 ORDER BY created_at",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing agents for project %s: %w", projectID, err)
	}
	return agents, nil
}

func (s *Store) UpdateStatus(ctx context.Context, id shared.ID, status string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE agents SET status = $1, updated_at = now() WHERE id = $2",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("updating agent status: %w", err)
	}
	return nil
}

func (s *Store) UpdateHealth(ctx context.Context, id shared.ID, health string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE agents SET health = $1, updated_at = now() WHERE id = $2",
		health, id,
	)
	if err != nil {
		return fmt.Errorf("updating agent health: %w", err)
	}
	return nil
}

func (s *Store) UpdateCurrentTask(ctx context.Context, id shared.ID, taskID *shared.ID) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE agents SET current_task_id = $1, updated_at = now() WHERE id = $2",
		taskID, id,
	)
	if err != nil {
		return fmt.Errorf("updating agent current task: %w", err)
	}
	return nil
}

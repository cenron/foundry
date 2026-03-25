package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

func (s *Store) Create(ctx context.Context, params CreateProjectParams) (*Project, error) {
	teamJSON, err := json.Marshal(params.TeamComposition)
	if err != nil {
		return nil, fmt.Errorf("marshaling team composition: %w", err)
	}

	var p Project
	err = s.db.QueryRowxContext(ctx, `
		INSERT INTO projects (name, description, repo_url, team_composition)
		VALUES ($1, $2, $3, $4)
		RETURNING *`,
		params.Name, params.Description, params.RepoURL, teamJSON,
	).StructScan(&p)
	if err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	return &p, nil
}

func (s *Store) GetByID(ctx context.Context, id shared.ID) (*Project, error) {
	var p Project
	err := s.db.GetContext(ctx, &p, "SELECT * FROM projects WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Resource: "project", ID: id.String()}
		}
		return nil, fmt.Errorf("getting project %s: %w", id, err)
	}
	return &p, nil
}

func (s *Store) List(ctx context.Context, page, pageSize int) ([]Project, int, error) {
	params := shared.PaginationParams{Page: page, PageSize: pageSize}

	var total int
	if err := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM projects"); err != nil {
		return nil, 0, fmt.Errorf("counting projects: %w", err)
	}

	var projects []Project
	err := s.db.SelectContext(ctx, &projects,
		"SELECT * FROM projects ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		params.Limit(), params.Offset(),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing projects: %w", err)
	}

	return projects, total, nil
}

func (s *Store) Update(ctx context.Context, id shared.ID, name, description string) (*Project, error) {
	var p Project
	err := s.db.QueryRowxContext(ctx,
		"UPDATE projects SET name = $1, description = $2, updated_at = now() WHERE id = $3 RETURNING *",
		name, description, id,
	).StructScan(&p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Resource: "project", ID: id.String()}
		}
		return nil, fmt.Errorf("updating project %s: %w", id, err)
	}
	return &p, nil
}

func (s *Store) UpdateStatus(ctx context.Context, id shared.ID, status string) error {
	result, err := s.db.ExecContext(ctx,
		"UPDATE projects SET status = $1, updated_at = now() WHERE id = $2",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("updating project status: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return &shared.NotFoundError{Resource: "project", ID: id.String()}
	}
	return nil
}

type SpecStore struct {
	db *sqlx.DB
}

func NewSpecStore(db *sqlx.DB) *SpecStore {
	return &SpecStore{db: db}
}

func (s *SpecStore) Create(ctx context.Context, params CreateSpecParams) (*Spec, error) {
	var spec Spec
	err := s.db.QueryRowxContext(ctx, `
		INSERT INTO specs (project_id, approved_content, execution_content, token_estimate, agent_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *`,
		params.ProjectID, params.ApprovedContent, params.ExecutionContent,
		params.TokenEstimate, params.AgentCount,
	).StructScan(&spec)
	if err != nil {
		return nil, fmt.Errorf("creating spec: %w", err)
	}
	return &spec, nil
}

func (s *SpecStore) GetByProjectID(ctx context.Context, projectID shared.ID) (*Spec, error) {
	var spec Spec
	err := s.db.GetContext(ctx, &spec,
		"SELECT * FROM specs WHERE project_id = $1 ORDER BY created_at DESC LIMIT 1",
		projectID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &shared.NotFoundError{Resource: "spec", ID: projectID.String()}
		}
		return nil, fmt.Errorf("getting spec for project %s: %w", projectID, err)
	}
	return &spec, nil
}

func (s *SpecStore) UpdateContent(ctx context.Context, projectID shared.ID, params CreateSpecParams) (*Spec, error) {
	existing, err := s.GetByProjectID(ctx, projectID)
	if err != nil {
		var notFound *shared.NotFoundError
		if !errors.As(err, &notFound) {
			return nil, err
		}
		// No existing spec — create one.
		return s.Create(ctx, params)
	}

	var spec Spec
	err = s.db.QueryRowxContext(ctx, `
		UPDATE specs
		SET    approved_content  = $1,
		       execution_content = $2,
		       token_estimate    = $3,
		       agent_count       = $4,
		       updated_at        = now()
		WHERE  id = $5
		RETURNING *`,
		params.ApprovedContent,
		params.ExecutionContent,
		params.TokenEstimate,
		params.AgentCount,
		existing.ID,
	).StructScan(&spec)
	if err != nil {
		return nil, fmt.Errorf("updating spec for project %s: %w", projectID, err)
	}
	return &spec, nil
}

func (s *SpecStore) UpdateApproval(ctx context.Context, id shared.ID, status string) error {
	query := "UPDATE specs SET approval_status = $1, updated_at = now()"
	if status == "approved" {
		query += ", approved_at = now()"
	}
	query += " WHERE id = $2"

	_, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("updating spec approval: %w", err)
	}
	return nil
}

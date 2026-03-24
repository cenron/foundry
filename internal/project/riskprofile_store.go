package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/cenron/foundry/internal/shared"
)

// RiskProfile holds risk classification criteria and model routing rules.
type RiskProfile struct {
	ID             shared.ID       `db:"id" json:"id"`
	ProjectID      *shared.ID      `db:"project_id" json:"project_id,omitempty"`
	Name           string          `db:"name" json:"name"`
	LowCriteria    json.RawMessage `db:"low_criteria" json:"low_criteria" swaggertype:"object"`
	MediumCriteria json.RawMessage `db:"medium_criteria" json:"medium_criteria" swaggertype:"object"`
	HighCriteria   json.RawMessage `db:"high_criteria" json:"high_criteria" swaggertype:"object"`
	ModelRouting   json.RawMessage `db:"model_routing" json:"model_routing" swaggertype:"object"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
}

// UpdateRiskProfileParams holds the fields that may be updated on a risk profile.
type UpdateRiskProfileParams struct {
	Name           string          `json:"name"`
	LowCriteria    json.RawMessage `json:"low_criteria"`
	MediumCriteria json.RawMessage `json:"medium_criteria"`
	HighCriteria   json.RawMessage `json:"high_criteria"`
	ModelRouting   json.RawMessage `json:"model_routing"`
}

// RiskProfileStore handles persistence for risk profiles.
type RiskProfileStore struct {
	db *sqlx.DB
}

func NewRiskProfileStore(db *sqlx.DB) *RiskProfileStore {
	return &RiskProfileStore{db: db}
}

// GetByProjectID returns the project-specific risk profile, or the global default when none exists.
func (s *RiskProfileStore) GetByProjectID(ctx context.Context, projectID shared.ID) (*RiskProfile, error) {
	var profile RiskProfile

	err := s.db.GetContext(ctx, &profile,
		"SELECT * FROM risk_profiles WHERE project_id = $1 ORDER BY created_at DESC LIMIT 1",
		projectID,
	)
	if err == nil {
		return &profile, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("getting risk profile for project %s: %w", projectID, err)
	}

	// Fall back to global default (project_id IS NULL).
	err = s.db.GetContext(ctx, &profile,
		"SELECT * FROM risk_profiles WHERE project_id IS NULL ORDER BY created_at ASC LIMIT 1",
	)
	if err != nil {
		return nil, fmt.Errorf("getting risk profile for project %s: %w", projectID, err)
	}

	return &profile, nil
}

// Create inserts a new project-specific risk profile.
func (s *RiskProfileStore) Create(ctx context.Context, projectID shared.ID, params UpdateRiskProfileParams) (*RiskProfile, error) {
	var profile RiskProfile

	err := s.db.QueryRowxContext(ctx, `
		INSERT INTO risk_profiles (project_id, name, low_criteria, medium_criteria, high_criteria, model_routing)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *`,
		projectID,
		params.Name,
		params.LowCriteria,
		params.MediumCriteria,
		params.HighCriteria,
		params.ModelRouting,
	).StructScan(&profile)
	if err != nil {
		return nil, fmt.Errorf("creating risk profile for project %s: %w", projectID, err)
	}

	return &profile, nil
}

// Update replaces criteria and routing on an existing risk profile.
func (s *RiskProfileStore) Update(ctx context.Context, id shared.ID, params UpdateRiskProfileParams) (*RiskProfile, error) {
	var profile RiskProfile

	err := s.db.QueryRowxContext(ctx, `
		UPDATE risk_profiles
		SET    name            = $1,
		       low_criteria    = $2,
		       medium_criteria = $3,
		       high_criteria   = $4,
		       model_routing   = $5,
		       updated_at      = now()
		WHERE  id = $6
		RETURNING *`,
		params.Name,
		params.LowCriteria,
		params.MediumCriteria,
		params.HighCriteria,
		params.ModelRouting,
		id,
	).StructScan(&profile)
	if err != nil {
		return nil, fmt.Errorf("updating risk profile %s: %w", id, err)
	}

	return &profile, nil
}

package project

import (
	"encoding/json"
	"time"

	"github.com/cenron/foundry/internal/shared"
)

type Project struct {
	ID              shared.ID       `db:"id" json:"id"`
	Name            string          `db:"name" json:"name"`
	Description     string          `db:"description" json:"description"`
	Status          string          `db:"status" json:"status"`
	RepoURL         string          `db:"repo_url" json:"repo_url"`
	TeamComposition json.RawMessage `db:"team_composition" json:"team_composition" swaggertype:"array,string"`
	ContainerID     *string         `db:"container_id" json:"container_id,omitempty"`
	RiskProfileID   *shared.ID      `db:"risk_profile_id" json:"risk_profile_id,omitempty"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

type CreateProjectParams struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	RepoURL         string   `json:"repo_url"`
	TeamComposition []string `json:"team_composition"`
}

type Spec struct {
	ID               shared.ID  `db:"id" json:"id"`
	ProjectID        shared.ID  `db:"project_id" json:"project_id"`
	ApprovedContent  string     `db:"approved_content" json:"approved_content"`
	ExecutionContent string     `db:"execution_content" json:"execution_content"`
	TokenEstimate    int        `db:"token_estimate" json:"token_estimate"`
	AgentCount       int        `db:"agent_count" json:"agent_count"`
	ApprovalStatus   string     `db:"approval_status" json:"approval_status"`
	ApprovedAt       *time.Time `db:"approved_at" json:"approved_at,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateSpecParams struct {
	ProjectID        shared.ID `json:"project_id"`
	ApprovedContent  string    `json:"approved_content"`
	ExecutionContent string    `json:"execution_content"`
	TokenEstimate    int       `json:"token_estimate"`
	AgentCount       int       `json:"agent_count"`
}

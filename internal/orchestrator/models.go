package orchestrator

import (
	"time"

	"github.com/lib/pq"

	"github.com/cenron/foundry/internal/shared"
)

type Task struct {
	ID                 shared.ID       `db:"id" json:"id"`
	ProjectID          shared.ID       `db:"project_id" json:"project_id"`
	SpecID             *shared.ID      `db:"spec_id" json:"spec_id,omitempty"`
	Title              string          `db:"title" json:"title"`
	Description        string          `db:"description" json:"description"`
	Status             string          `db:"status" json:"status"`
	RiskLevel          string          `db:"risk_level" json:"risk_level"`
	AssignedRole       string          `db:"assigned_role" json:"assigned_role"`
	AssignedAgentID    *shared.ID      `db:"assigned_agent_id" json:"assigned_agent_id,omitempty"`
	DependsOn          pq.StringArray  `db:"depends_on" json:"depends_on" swaggertype:"array,string"`
	AutomationEligible bool           `db:"automation_eligible" json:"automation_eligible"`
	ModelTier          string          `db:"model_tier" json:"model_tier"`
	TokenUsage         int             `db:"token_usage" json:"token_usage"`
	ContextSummary     *string         `db:"context_summary" json:"context_summary,omitempty"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at" json:"updated_at"`
}

type CreateTaskParams struct {
	ProjectID    shared.ID   `json:"project_id"`
	SpecID       *shared.ID  `json:"spec_id,omitempty"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	RiskLevel    string      `json:"risk_level"`
	AssignedRole string      `json:"assigned_role"`
	DependsOn    []shared.ID `json:"depends_on"`
}

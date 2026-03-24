package event

import (
	"encoding/json"
	"time"

	"github.com/cenron/foundry/internal/shared"
)

type Event struct {
	ID        shared.ID       `db:"id" json:"id"`
	ProjectID shared.ID       `db:"project_id" json:"project_id"`
	TaskID    *shared.ID      `db:"task_id" json:"task_id,omitempty"`
	AgentID   *shared.ID      `db:"agent_id" json:"agent_id,omitempty"`
	Type      string          `db:"type" json:"type"`
	Payload   json.RawMessage `db:"payload" json:"payload"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

type CreateEventParams struct {
	ProjectID shared.ID   `json:"project_id"`
	TaskID    *shared.ID  `json:"task_id,omitempty"`
	AgentID   *shared.ID  `json:"agent_id,omitempty"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
}

type Artifact struct {
	ID          shared.ID  `db:"id" json:"id"`
	ProjectID   shared.ID  `db:"project_id" json:"project_id"`
	TaskID      *shared.ID `db:"task_id" json:"task_id,omitempty"`
	AgentID     *shared.ID `db:"agent_id" json:"agent_id,omitempty"`
	Type        string     `db:"type" json:"type"`
	Path        string     `db:"path" json:"path"`
	Description string     `db:"description" json:"description"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

type CreateArtifactParams struct {
	ProjectID   shared.ID  `json:"project_id"`
	TaskID      *shared.ID `json:"task_id,omitempty"`
	AgentID     *shared.ID `json:"agent_id,omitempty"`
	Type        string     `json:"type"`
	Path        string     `json:"path"`
	Description string     `json:"description"`
}

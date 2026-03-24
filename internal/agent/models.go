package agent

import (
	"time"

	"github.com/cenron/foundry/internal/shared"
)

type Agent struct {
	ID            shared.ID  `db:"id" json:"id"`
	ProjectID     shared.ID  `db:"project_id" json:"project_id"`
	Role          string     `db:"role" json:"role"`
	Provider      string     `db:"provider" json:"provider"`
	ContainerID   string     `db:"container_id" json:"container_id"`
	ProcessID     *string    `db:"process_id" json:"process_id,omitempty"`
	WorktreePath  *string    `db:"worktree_path" json:"worktree_path,omitempty"`
	BranchName    *string    `db:"branch_name" json:"branch_name,omitempty"`
	Status        string     `db:"status" json:"status"`
	CurrentTaskID *shared.ID `db:"current_task_id" json:"current_task_id,omitempty"`
	Health        string     `db:"health" json:"health"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateAgentParams struct {
	ProjectID   shared.ID `json:"project_id"`
	Role        string    `json:"role"`
	Provider    string    `json:"provider"`
	ContainerID string    `json:"container_id"`
}

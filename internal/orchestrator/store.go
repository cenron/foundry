package orchestrator

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"github.com/cenron/foundry/internal/shared"
	"github.com/jmoiron/sqlx"
)

type TaskStore struct {
	db *sqlx.DB
}

func NewTaskStore(db *sqlx.DB) *TaskStore {
	return &TaskStore{db: db}
}

func (s *TaskStore) Create(ctx context.Context, params CreateTaskParams) (*Task, error) {
	depStrings := make([]string, len(params.DependsOn))
	for i, id := range params.DependsOn {
		depStrings[i] = id.String()
	}

	var task Task
	err := s.db.QueryRowxContext(ctx, `
		INSERT INTO tasks (project_id, spec_id, title, description, risk_level, assigned_role, depends_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *`,
		params.ProjectID, params.SpecID, params.Title, params.Description,
		params.RiskLevel, params.AssignedRole, pq.Array(depStrings),
	).StructScan(&task)
	if err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}

	return &task, nil
}

func (s *TaskStore) GetByID(ctx context.Context, id shared.ID) (*Task, error) {
	var task Task
	err := s.db.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("getting task %s: %w", id, err)
	}
	return &task, nil
}

func (s *TaskStore) ListByProject(ctx context.Context, projectID shared.ID) ([]Task, error) {
	var tasks []Task
	err := s.db.SelectContext(ctx, &tasks,
		"SELECT * FROM tasks WHERE project_id = $1 ORDER BY created_at",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing tasks for project %s: %w", projectID, err)
	}
	return tasks, nil
}

func (s *TaskStore) UpdateStatus(ctx context.Context, id shared.ID, status string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE tasks SET status = $1, updated_at = now() WHERE id = $2",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("updating task status: %w", err)
	}
	return nil
}

func (s *TaskStore) UpdateAssignment(ctx context.Context, id shared.ID, agentID shared.ID) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE tasks SET assigned_agent_id = $1, status = 'assigned', updated_at = now() WHERE id = $2",
		agentID, id,
	)
	if err != nil {
		return fmt.Errorf("updating task assignment: %w", err)
	}
	return nil
}

// GetUnblockedTasks returns pending tasks where all dependencies are done.
func (s *TaskStore) GetUnblockedTasks(ctx context.Context, projectID shared.ID) ([]Task, error) {
	var tasks []Task
	err := s.db.SelectContext(ctx, &tasks, `
		SELECT t.*
		FROM tasks t
		WHERE t.project_id = $1
		  AND t.status = 'pending'
		  AND NOT EXISTS (
		    SELECT 1
		    FROM unnest(t.depends_on) AS dep_id
		    JOIN tasks d ON d.id = dep_id::uuid
		    WHERE d.status != 'done'
		  )
		ORDER BY t.created_at`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting unblocked tasks: %w", err)
	}
	return tasks, nil
}

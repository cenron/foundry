CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    provider TEXT NOT NULL DEFAULT 'claude',
    container_id TEXT NOT NULL DEFAULT '',
    process_id TEXT,
    worktree_path TEXT,
    branch_name TEXT,
    status TEXT NOT NULL DEFAULT 'starting',
    current_task_id UUID,
    health TEXT NOT NULL DEFAULT 'healthy',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agents_project_id ON agents (project_id);
CREATE INDEX idx_agents_status ON agents (status);

-- Add foreign key from tasks.assigned_agent_id to agents now that both tables exist
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_assigned_agent FOREIGN KEY (assigned_agent_id) REFERENCES agents(id) ON DELETE SET NULL;

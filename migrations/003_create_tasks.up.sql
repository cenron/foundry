CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    spec_id UUID REFERENCES specs(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    risk_level TEXT NOT NULL DEFAULT 'medium',
    assigned_role TEXT NOT NULL DEFAULT '',
    assigned_agent_id UUID,
    depends_on UUID[] NOT NULL DEFAULT '{}',
    automation_eligible BOOLEAN NOT NULL DEFAULT false,
    model_tier TEXT NOT NULL DEFAULT 'sonnet',
    token_usage INTEGER NOT NULL DEFAULT 0,
    context_summary TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_project_id ON tasks (project_id);
CREATE INDEX idx_tasks_status ON tasks (status);
CREATE INDEX idx_tasks_project_status ON tasks (project_id, status);

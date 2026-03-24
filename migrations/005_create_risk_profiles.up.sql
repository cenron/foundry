CREATE TABLE risk_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    low_criteria JSONB NOT NULL DEFAULT '{}',
    medium_criteria JSONB NOT NULL DEFAULT '{}',
    high_criteria JSONB NOT NULL DEFAULT '{}',
    model_routing JSONB NOT NULL DEFAULT '{"claude": {"low": "haiku", "medium": "sonnet", "high": "opus"}}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_risk_profiles_project_id ON risk_profiles (project_id);

-- Default global risk profile (project_id NULL)
INSERT INTO risk_profiles (name, low_criteria, medium_criteria, high_criteria)
VALUES (
    'Default',
    '{"patterns": ["CRUD", "config", "boilerplate", "documentation"]}',
    '{"patterns": ["new features", "integrations", "API changes"]}',
    '{"patterns": ["auth", "payments", "data migrations", "security"]}'
);

-- Add foreign key from projects.risk_profile_id now that risk_profiles exists
ALTER TABLE projects ADD CONSTRAINT fk_projects_risk_profile FOREIGN KEY (risk_profile_id) REFERENCES risk_profiles(id) ON DELETE SET NULL;

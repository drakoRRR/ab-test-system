CREATE TABLE flags (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key        TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    enabled    BOOLEAN     NOT NULL DEFAULT FALSE,
    rules      JSONB       NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_flags_project_key UNIQUE (project_id, key)
);

CREATE INDEX idx_flags_project_id ON flags(project_id);

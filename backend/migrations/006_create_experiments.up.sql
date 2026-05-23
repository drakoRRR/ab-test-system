CREATE TABLE experiments (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    flag_id         UUID        REFERENCES flags(id) ON DELETE SET NULL,
    name            TEXT        NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    status          TEXT        NOT NULL DEFAULT 'draft'
                                CHECK (status IN ('draft', 'running', 'paused', 'completed')),
    traffic_percent FLOAT       NOT NULL DEFAULT 0
                                CHECK (traffic_percent >= 0 AND traffic_percent <= 100),
    variants        JSONB       NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    ended_at        TIMESTAMPTZ
);

CREATE INDEX idx_experiments_project_id ON experiments(project_id);
CREATE INDEX idx_experiments_flag_id    ON experiments(flag_id);
CREATE INDEX idx_experiments_status     ON experiments(status);

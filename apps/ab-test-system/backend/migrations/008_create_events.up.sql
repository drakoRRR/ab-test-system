CREATE TABLE events (
    id            UUID        PRIMARY KEY,
    project_id    UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id       TEXT        NOT NULL,
    experiment_id UUID        NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    variant_id    UUID        NOT NULL,
    type          TEXT        NOT NULL CHECK (type IN ('exposure', 'conversion')),
    name          TEXT        NOT NULL DEFAULT '',
    value         DOUBLE PRECISION NOT NULL DEFAULT 0,
    ts            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_experiment_id ON events(experiment_id);
CREATE INDEX idx_events_project_id    ON events(project_id);
CREATE INDEX idx_events_ts            ON events(ts);

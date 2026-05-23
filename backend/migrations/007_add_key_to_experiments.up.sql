ALTER TABLE experiments ADD COLUMN key TEXT NOT NULL DEFAULT '';
ALTER TABLE experiments ADD CONSTRAINT uq_experiments_project_key UNIQUE (project_id, key);
CREATE INDEX idx_experiments_key ON experiments(key);

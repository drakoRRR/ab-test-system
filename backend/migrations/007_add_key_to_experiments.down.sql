DROP INDEX IF EXISTS idx_experiments_key;
ALTER TABLE experiments DROP CONSTRAINT IF EXISTS uq_experiments_project_key;
ALTER TABLE experiments DROP COLUMN IF EXISTS key;

DROP INDEX IF EXISTS idx_exams_created_by;
ALTER TABLE exams DROP COLUMN IF EXISTS created_by;

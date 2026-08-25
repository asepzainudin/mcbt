DROP INDEX IF EXISTS idx_answers_grading;
ALTER TABLE exam_answers
    DROP COLUMN IF EXISTS score,
    DROP COLUMN IF EXISTS is_correct,
    DROP COLUMN IF EXISTS feedback,
    DROP COLUMN IF EXISTS graded_at,
    DROP COLUMN IF EXISTS graded_via;

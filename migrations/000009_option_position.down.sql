DROP INDEX IF EXISTS idx_options_question_position;
ALTER TABLE question_options DROP CONSTRAINT IF EXISTS ck_options_label;
ALTER TABLE question_options DROP COLUMN IF EXISTS position;
ALTER TABLE question_options
    ADD CONSTRAINT ck_options_label CHECK (label IN ('A', 'B', 'C', 'D', 'E'));

UPDATE question_options
SET label = sub.new_label
FROM (
    SELECT id,
           CHR(64 + ROW_NUMBER() OVER (PARTITION BY question_id ORDER BY position)) AS new_label
    FROM question_options
) sub
WHERE question_options.id = sub.id;

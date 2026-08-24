ALTER TABLE question_options ADD COLUMN position SMALLINT NOT NULL DEFAULT 0;

UPDATE question_options
SET position = sub.rn
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY question_id ORDER BY label) AS rn
    FROM question_options
) sub
WHERE question_options.id = sub.id;

CREATE INDEX idx_options_question_position ON question_options (question_id, position);

ALTER TABLE question_options DROP CONSTRAINT ck_options_label;
ALTER TABLE question_options
    ADD CONSTRAINT ck_options_label CHECK (label IN ('A', 'B', 'C', 'D', 'E', '1', '2', '3', '4', '5'));

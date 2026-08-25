ALTER TABLE exam_answers
    ADD COLUMN score      NUMERIC(6,2),
    ADD COLUMN is_correct BOOLEAN,
    ADD COLUMN feedback   TEXT,
    ADD COLUMN graded_at  TIMESTAMPTZ,
    ADD COLUMN graded_via VARCHAR(10);

CREATE INDEX idx_answers_grading ON exam_answers (question_id, graded_at);

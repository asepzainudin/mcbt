ALTER TABLE questions DROP COLUMN IF EXISTS answer_keys;

ALTER TABLE questions DROP CONSTRAINT IF EXISTS ck_questions_type;
ALTER TABLE questions
    ADD CONSTRAINT ck_questions_type CHECK (question_type IN ('multiple_choice', 'true_false', 'essay'));

ALTER TABLE questions ALTER COLUMN score_weight TYPE INTEGER USING round(score_weight)::int;
ALTER TABLE questions ALTER COLUMN score_weight SET DEFAULT 1;
ALTER TABLE questions RENAME COLUMN score_weight TO points;

ALTER TABLE question_banks DROP CONSTRAINT IF EXISTS ck_question_banks_status;
ALTER TABLE question_banks DROP COLUMN IF EXISTS status;

DROP INDEX IF EXISTS uq_question_banks_code;
ALTER TABLE question_banks DROP COLUMN IF EXISTS code;

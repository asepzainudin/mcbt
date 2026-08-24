ALTER TABLE question_banks ADD COLUMN code VARCHAR(50);

UPDATE question_banks
SET code = 'BANK-' || upper(regexp_replace(left(title, 10), '[^A-Za-z0-9]', '', 'g')) || '-' || left(id::text, 4);

ALTER TABLE question_banks ALTER COLUMN code SET NOT NULL;
CREATE UNIQUE INDEX uq_question_banks_code ON question_banks (code);

ALTER TABLE question_banks ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft';
ALTER TABLE question_banks
    ADD CONSTRAINT ck_question_banks_status CHECK (status IN ('draft', 'published', 'archived'));

ALTER TABLE questions RENAME COLUMN points TO score_weight;
ALTER TABLE questions ALTER COLUMN score_weight TYPE NUMERIC(6,2) USING score_weight::numeric;
ALTER TABLE questions ALTER COLUMN score_weight SET DEFAULT 1.0;

ALTER TABLE questions DROP CONSTRAINT ck_questions_type;
ALTER TABLE questions
    ADD CONSTRAINT ck_questions_type CHECK (question_type IN (
        'multiple_choice', 'true_false', 'multiple_answer', 'essay', 'short_answer'
    ));

ALTER TABLE questions ADD COLUMN answer_keys TEXT;

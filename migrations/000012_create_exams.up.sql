CREATE TABLE exams (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title                  VARCHAR(150) NOT NULL,
    description            TEXT,
    subject_id             UUID NOT NULL REFERENCES subjects (id),
    academic_year_id       UUID REFERENCES academic_years (id),
    question_bank_id       UUID REFERENCES question_banks (id),
    status                 VARCHAR(20) NOT NULL DEFAULT 'draft',
    duration_minutes       INTEGER NOT NULL DEFAULT 60,
    max_attempts           INTEGER NOT NULL DEFAULT 1,
    passing_grade          NUMERIC(5,2) NOT NULL DEFAULT 75.00,
    randomize_questions    BOOLEAN NOT NULL DEFAULT FALSE,
    randomize_options      BOOLEAN NOT NULL DEFAULT FALSE,
    allow_backtrack        BOOLEAN NOT NULL DEFAULT TRUE,
    auto_submit            BOOLEAN NOT NULL DEFAULT TRUE,
    show_result_immediately BOOLEAN NOT NULL DEFAULT FALSE,
    negative_marking       BOOLEAN NOT NULL DEFAULT FALSE,
    negative_value         NUMERIC(4,2) NOT NULL DEFAULT 0.00,
    token_enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    exam_token             VARCHAR(10),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_exams_status CHECK (status IN ('draft', 'published', 'closed')),
    CONSTRAINT ck_exams_duration CHECK (duration_minutes BETWEEN 1 AND 600),
    CONSTRAINT ck_exams_attempts CHECK (max_attempts BETWEEN 1 AND 10),
    CONSTRAINT ck_exams_passing CHECK (passing_grade BETWEEN 0 AND 100),
    CONSTRAINT ck_exams_negative CHECK (negative_value BETWEEN 0 AND 100)
);

CREATE INDEX idx_exams_subject_id ON exams (subject_id);
CREATE INDEX idx_exams_academic_year_id ON exams (academic_year_id);
CREATE INDEX idx_exams_question_bank_id ON exams (question_bank_id);
CREATE UNIQUE INDEX uq_exams_token ON exams (exam_token) WHERE exam_token IS NOT NULL;

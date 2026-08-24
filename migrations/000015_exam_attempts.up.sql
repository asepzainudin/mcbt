CREATE TABLE exam_attempts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id      UUID        NOT NULL REFERENCES exams (id) ON DELETE CASCADE,
    student_id   UUID        NOT NULL REFERENCES students (id) ON DELETE CASCADE,
    attempt_no   INTEGER     NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    started_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL,
    submitted_at TIMESTAMPTZ,
    score        NUMERIC(5,2),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (exam_id, student_id, attempt_no),
    CONSTRAINT ck_attempt_status CHECK (status IN ('in_progress', 'submitted', 'expired'))
);

CREATE INDEX idx_attempts_exam_student ON exam_attempts (exam_id, student_id);
CREATE INDEX idx_attempts_student ON exam_attempts (student_id);

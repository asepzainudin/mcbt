CREATE TABLE question_reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id  UUID        NOT NULL REFERENCES exam_attempts (id) ON DELETE CASCADE,
    question_id UUID        NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    student_id  UUID        NOT NULL REFERENCES students (id) ON DELETE CASCADE,
    reason      TEXT        NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    resolution  TEXT,
    resolved_by UUID        REFERENCES users (id),
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (attempt_id, question_id),
    CONSTRAINT ck_report_status CHECK (status IN ('pending', 'reviewing', 'resolved', 'rejected'))
);

CREATE INDEX idx_reports_question ON question_reports (question_id);
CREATE INDEX idx_reports_status ON question_reports (status);

ALTER TABLE exams ADD COLUMN results_published BOOLEAN NOT NULL DEFAULT FALSE;

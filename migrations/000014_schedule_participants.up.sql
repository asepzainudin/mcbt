CREATE TABLE exam_schedules (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id    UUID        NOT NULL UNIQUE REFERENCES exams (id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    token      VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_schedules_range CHECK (end_time > start_time)
);

CREATE UNIQUE INDEX uq_exam_schedules_token ON exam_schedules (token);

CREATE TABLE exam_participants (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id      UUID        NOT NULL REFERENCES exams (id) ON DELETE CASCADE,
    student_id   UUID        NOT NULL UNIQUE REFERENCES students (id) ON DELETE CASCADE,
    assigned_via VARCHAR(10) NOT NULL DEFAULT 'individual',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_participants_via CHECK (assigned_via IN ('class', 'individual'))
);

CREATE INDEX idx_exam_participants_exam ON exam_participants (exam_id);
CREATE INDEX idx_exam_participants_student ON exam_participants (student_id);

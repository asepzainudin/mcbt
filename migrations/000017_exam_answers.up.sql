CREATE TABLE exam_answers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id       UUID        NOT NULL REFERENCES exam_attempts (id) ON DELETE CASCADE,
    question_id      UUID        NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    answer_value     TEXT        NOT NULL DEFAULT '',
    client_timestamp BIGINT      NOT NULL DEFAULT 0,
    is_flagged       BOOLEAN     NOT NULL DEFAULT FALSE,
    answered_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (attempt_id, question_id)
);

CREATE INDEX idx_answers_attempt ON exam_answers (attempt_id);

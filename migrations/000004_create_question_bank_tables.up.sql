-- =============================================================
-- Migration 000004: Media and question bank
-- =============================================================

CREATE TABLE media (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uploaded_by UUID         REFERENCES users (id) ON DELETE SET NULL,
    file_name   VARCHAR(255) NOT NULL,
    file_path   VARCHAR(500) NOT NULL,
    mime_type   VARCHAR(100) NOT NULL,
    file_size   BIGINT       NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_media_file_path ON media (file_path);
CREATE INDEX idx_media_uploaded_by ON media (uploaded_by);

CREATE TABLE question_banks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id       UUID         NOT NULL REFERENCES subjects (id),
    academic_year_id UUID         REFERENCES academic_years (id) ON DELETE SET NULL,
    created_by       UUID         REFERENCES users (id) ON DELETE SET NULL,
    title            VARCHAR(150) NOT NULL,
    description      TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_question_banks_subject_title ON question_banks (subject_id, title);
CREATE INDEX idx_question_banks_academic_year_id ON question_banks (academic_year_id);
CREATE INDEX idx_question_banks_created_by ON question_banks (created_by);

CREATE TABLE questions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_bank_id UUID         NOT NULL REFERENCES question_banks (id) ON DELETE CASCADE,
    media_id         UUID         REFERENCES media (id) ON DELETE SET NULL,
    question_type    VARCHAR(20)  NOT NULL DEFAULT 'multiple_choice',
    content          TEXT         NOT NULL,
    points           INTEGER      NOT NULL DEFAULT 1,
    explanation      TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT ck_questions_type CHECK (
        question_type IN ('multiple_choice', 'true_false', 'essay')
    ),
    CONSTRAINT ck_questions_points CHECK (points > 0)
);

CREATE INDEX idx_questions_bank_id ON questions (question_bank_id);
CREATE INDEX idx_questions_type ON questions (question_type);
CREATE INDEX idx_questions_media_id ON questions (media_id);

CREATE TABLE question_options (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID        NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    media_id    UUID        REFERENCES media (id) ON DELETE SET NULL,
    label       CHAR(1)     NOT NULL,
    content     TEXT        NOT NULL,
    is_correct  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_options_label CHECK (label IN ('A', 'B', 'C', 'D', 'E'))
);

CREATE UNIQUE INDEX uq_options_question_label ON question_options (question_id, label);
CREATE INDEX idx_options_media_id ON question_options (media_id);

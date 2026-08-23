-- =============================================================
-- Migration 000002: Academic years, subjects, teachers
-- =============================================================

CREATE TABLE academic_years (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(20) NOT NULL,
    start_date DATE        NOT NULL,
    end_date   DATE        NOT NULL,
    is_active  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_academic_years_dates CHECK (end_date > start_date)
);

CREATE UNIQUE INDEX uq_academic_years_name ON academic_years (name);
CREATE UNIQUE INDEX uq_academic_years_single_active ON academic_years (is_active) WHERE is_active;

CREATE TABLE subjects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_subjects_code ON subjects (code);

CREATE TABLE teachers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    nip        VARCHAR(30),
    phone      VARCHAR(20),
    address    VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_teachers_user_id ON teachers (user_id);
CREATE UNIQUE INDEX uq_teachers_nip ON teachers (nip) WHERE nip IS NOT NULL;

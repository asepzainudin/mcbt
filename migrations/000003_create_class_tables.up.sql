-- =============================================================
-- Migration 000003: Classes and students
-- =============================================================

CREATE TABLE classes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    academic_year_id    UUID         NOT NULL REFERENCES academic_years (id),
    homeroom_teacher_id UUID         REFERENCES teachers (id) ON DELETE SET NULL,
    name                VARCHAR(100) NOT NULL,
    grade_level         SMALLINT,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_classes_year_name ON classes (academic_year_id, name);
CREATE INDEX idx_classes_homeroom_teacher_id ON classes (homeroom_teacher_id);

CREATE TABLE students (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    class_id   UUID        REFERENCES classes (id) ON DELETE SET NULL,
    nis        VARCHAR(30) NOT NULL,
    phone      VARCHAR(20),
    address    VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_students_user_id ON students (user_id);
CREATE UNIQUE INDEX uq_students_nis ON students (nis);
CREATE INDEX idx_students_class_id ON students (class_id);

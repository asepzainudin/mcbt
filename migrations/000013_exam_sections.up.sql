CREATE TABLE exam_sections (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id    UUID         NOT NULL REFERENCES exams (id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    sequence   INTEGER      NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (exam_id, name),
    UNIQUE (exam_id, sequence)
);

CREATE INDEX idx_exam_sections_exam_id ON exam_sections (exam_id);

CREATE TABLE exam_section_questions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id UUID        NOT NULL REFERENCES exam_sections (id) ON DELETE CASCADE,
    question_id UUID       NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (section_id, question_id)
);

CREATE INDEX idx_section_questions_section ON exam_section_questions (section_id);
CREATE INDEX idx_section_questions_question ON exam_section_questions (question_id);

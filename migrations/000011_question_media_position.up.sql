ALTER TABLE questions ADD COLUMN media_position VARCHAR(10) NOT NULL DEFAULT 'after';
ALTER TABLE questions
    ADD CONSTRAINT ck_questions_media_position CHECK (media_position IN ('before', 'after'));

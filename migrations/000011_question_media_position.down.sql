ALTER TABLE questions DROP CONSTRAINT IF EXISTS ck_questions_media_position;
ALTER TABLE questions DROP COLUMN IF EXISTS media_position;

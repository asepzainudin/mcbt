ALTER TABLE exams ADD COLUMN created_by UUID REFERENCES users(id);
CREATE INDEX idx_exams_created_by ON exams(created_by);

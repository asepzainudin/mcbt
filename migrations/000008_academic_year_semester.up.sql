ALTER TABLE academic_years RENAME COLUMN name TO year;

ALTER TABLE academic_years ADD COLUMN semester VARCHAR(9) NOT NULL DEFAULT 'ODD';

DROP INDEX IF EXISTS uq_academic_years_name;
CREATE UNIQUE INDEX uq_academic_years_year_semester ON academic_years (year, semester);

ALTER TABLE academic_years ALTER COLUMN start_date DROP NOT NULL;
ALTER TABLE academic_years ALTER COLUMN end_date DROP NOT NULL;

UPDATE academic_years
SET start_date = COALESCE(start_date, DATE '2000-07-01'),
    end_date   = COALESCE(end_date, DATE '2001-06-30');

DROP INDEX IF EXISTS uq_academic_years_year_semester;

ALTER TABLE academic_years DROP COLUMN IF EXISTS semester;

ALTER TABLE academic_years RENAME COLUMN year TO name;

CREATE UNIQUE INDEX uq_academic_years_name ON academic_years (name);

ALTER TABLE academic_years ALTER COLUMN start_date SET NOT NULL;
ALTER TABLE academic_years ALTER COLUMN end_date SET NOT NULL;

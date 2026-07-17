DROP INDEX IF EXISTS idx_skills_tags;

ALTER TABLE skills
    DROP COLUMN IF EXISTS tags;

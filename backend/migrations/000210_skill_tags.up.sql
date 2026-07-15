ALTER TABLE skills
    ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

CREATE INDEX idx_skills_tags ON skills USING GIN (tags);

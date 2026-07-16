ALTER TABLE experts
    DROP CONSTRAINT IF EXISTS chk_experts_revision_positive;

ALTER TABLE experts
    DROP COLUMN IF EXISTS revision;

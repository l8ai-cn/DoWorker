DROP INDEX IF EXISTS idx_loops_model_resource_id;

ALTER TABLE loops DROP CONSTRAINT IF EXISTS loops_model_resource_id_fkey;

ALTER TABLE loops DROP COLUMN IF EXISTS model_resource_id;

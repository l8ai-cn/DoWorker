ALTER TABLE loops ADD COLUMN IF NOT EXISTS model_resource_id BIGINT;

ALTER TABLE loops
  ADD CONSTRAINT loops_model_resource_id_fkey
  FOREIGN KEY (model_resource_id) REFERENCES model_resources(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_loops_model_resource_id ON loops(model_resource_id);

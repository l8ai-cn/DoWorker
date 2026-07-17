ALTER TABLE agents
  DROP CONSTRAINT IF EXISTS agents_adapter_id_check,
  DROP COLUMN IF EXISTS adapter_id;

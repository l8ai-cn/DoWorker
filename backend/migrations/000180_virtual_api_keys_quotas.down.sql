DROP INDEX IF EXISTS idx_pods_virtual_api_key;
ALTER TABLE pods DROP COLUMN IF EXISTS virtual_api_key_id;

DROP TABLE IF EXISTS token_quotas;
DROP TABLE IF EXISTS virtual_api_keys;

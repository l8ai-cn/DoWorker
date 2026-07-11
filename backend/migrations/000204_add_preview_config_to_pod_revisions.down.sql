ALTER TABLE pod_config_revisions
  DROP CONSTRAINT IF EXISTS pod_config_revisions_preview_path_check;

ALTER TABLE pod_config_revisions
  DROP CONSTRAINT IF EXISTS pod_config_revisions_preview_port_check;

ALTER TABLE pods
  DROP CONSTRAINT IF EXISTS pods_preview_path_check;

ALTER TABLE pods
  DROP CONSTRAINT IF EXISTS pods_preview_port_check;

ALTER TABLE pod_config_revisions
  DROP COLUMN IF EXISTS preview_path;

ALTER TABLE pod_config_revisions
  DROP COLUMN IF EXISTS preview_port;

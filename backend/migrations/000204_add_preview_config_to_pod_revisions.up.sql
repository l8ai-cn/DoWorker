ALTER TABLE pod_config_revisions
  ADD COLUMN preview_port INTEGER NOT NULL DEFAULT 0;

ALTER TABLE pod_config_revisions
  ADD COLUMN preview_path VARCHAR(255) NOT NULL DEFAULT '/';

UPDATE pod_config_revisions
SET
  preview_port = pods.preview_port,
  preview_path = COALESCE(NULLIF(pods.preview_path, ''), '/')
FROM pods
WHERE pod_config_revisions.pod_id = pods.id;

ALTER TABLE pod_config_revisions
  ADD CONSTRAINT pod_config_revisions_preview_port_check
  CHECK (preview_port = 0 OR preview_port BETWEEN 1024 AND 65535);

ALTER TABLE pod_config_revisions
  ADD CONSTRAINT pod_config_revisions_preview_path_check
  CHECK (
    preview_path <> ''
    AND left(preview_path, 1) = '/'
    AND position('?' IN preview_path) = 0
    AND position('#' IN preview_path) = 0
    AND preview_path !~ '(^|/)\.\.(/|$)'
  );

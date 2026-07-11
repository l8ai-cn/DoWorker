ALTER TABLE pod_config_revisions
  DROP CONSTRAINT pod_config_revisions_status_check;

ALTER TABLE pod_config_revisions
  ADD CONSTRAINT pod_config_revisions_status_check
  CHECK (status IN ('draft', 'applying', 'active', 'failed', 'superseded'));

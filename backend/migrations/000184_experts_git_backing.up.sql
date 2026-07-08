ALTER TABLE experts ADD COLUMN git_repo_path  VARCHAR(255);
ALTER TABLE experts ADD COLUMN default_branch VARCHAR(255) NOT NULL DEFAULT 'main';
ALTER TABLE experts ADD COLUMN http_clone_url VARCHAR(1000);
ALTER TABLE experts ADD COLUMN metadata       JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN experts.git_repo_path  IS 'am-experts/org<ID>-<slug>; NULL = legacy row not yet git-backed (lazy provision on next update).';
COMMENT ON COLUMN experts.metadata       IS 'Derived cache of expert.json extras: avatar (形象, repo-relative path) + expertType (类型) + future non-column config.';

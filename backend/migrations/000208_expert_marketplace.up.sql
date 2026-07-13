CREATE TABLE expert_market_applications (
  id BIGSERIAL PRIMARY KEY,
  slug VARCHAR(100) NOT NULL,
  publisher_organization_id BIGINT NOT NULL
    REFERENCES organizations(id) ON DELETE CASCADE,
  publisher_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  is_operator_owned BOOLEAN NOT NULL DEFAULT FALSE,
  latest_published_release_id BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT expert_market_applications_slug_unique UNIQUE (slug),
  CONSTRAINT expert_market_applications_slug_check CHECK (
    slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
    AND char_length(slug) BETWEEN 2 AND 100
  )
);

CREATE TABLE expert_market_releases (
  id BIGSERIAL PRIMARY KEY,
  application_id BIGINT NOT NULL
    REFERENCES expert_market_applications(id) ON DELETE CASCADE,
  source_expert_id BIGINT NOT NULL,
  publisher_organization_id BIGINT NOT NULL
    REFERENCES organizations(id) ON DELETE CASCADE,
  publisher_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  version INTEGER NOT NULL CHECK (version > 0),
  status VARCHAR(32) NOT NULL CHECK (
    status IN ('draft', 'pending_review', 'published', 'rejected', 'withdrawn')
  ),
  name VARCHAR(255) NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  category VARCHAR(100) NOT NULL DEFAULT '',
  icon VARCHAR(100) NOT NULL DEFAULT '',
  tags TEXT[] NOT NULL DEFAULT '{}',
  outcomes TEXT[] NOT NULL DEFAULT '{}',
  featured BOOLEAN NOT NULL DEFAULT FALSE,
  expert_snapshot JSONB NOT NULL,
  worker_spec_snapshot JSONB NOT NULL,
  skill_dependencies JSONB NOT NULL,
  reviewer_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  rejection_reason TEXT,
  submitted_at TIMESTAMPTZ,
  reviewed_at TIMESTAMPTZ,
  published_at TIMESTAMPTZ,
  rejected_at TIMESTAMPTZ,
  withdrawn_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT expert_market_releases_application_version_unique
    UNIQUE (application_id, version),
  CONSTRAINT expert_market_releases_application_id_id_unique
    UNIQUE (application_id, id),
  CONSTRAINT expert_market_releases_expert_snapshot_check CHECK (
    jsonb_typeof(expert_snapshot) = 'object'
    AND (expert_snapshot->>'version') ~ '^[1-9][0-9]*$'
    AND (expert_snapshot->>'version')::NUMERIC <= 9223372036854775807
  ),
  CONSTRAINT expert_market_releases_worker_spec_snapshot_check CHECK (
    jsonb_typeof(worker_spec_snapshot) = 'object'
    AND (worker_spec_snapshot->>'version') ~ '^[1-9][0-9]*$'
    AND (worker_spec_snapshot->>'version')::NUMERIC <= 9223372036854775807
  ),
  CONSTRAINT expert_market_releases_skill_dependencies_check
    CHECK (jsonb_typeof(skill_dependencies) = 'array')
);

CREATE INDEX idx_expert_market_applications_publisher
  ON expert_market_applications(publisher_organization_id, created_at DESC);
CREATE INDEX idx_expert_market_releases_application
  ON expert_market_releases(application_id, version DESC);
CREATE INDEX idx_expert_market_releases_status
  ON expert_market_releases(status, created_at DESC);
CREATE INDEX idx_expert_market_releases_publisher
  ON expert_market_releases(publisher_organization_id, created_at DESC);

CREATE FUNCTION prevent_expert_market_release_immutable_update()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.id IS DISTINCT FROM OLD.id
    OR NEW.application_id IS DISTINCT FROM OLD.application_id
    OR NEW.source_expert_id IS DISTINCT FROM OLD.source_expert_id
    OR NEW.publisher_organization_id IS DISTINCT FROM OLD.publisher_organization_id
    OR NEW.publisher_user_id IS DISTINCT FROM OLD.publisher_user_id
    OR NEW.version IS DISTINCT FROM OLD.version
    OR NEW.name IS DISTINCT FROM OLD.name
    OR NEW.summary IS DISTINCT FROM OLD.summary
    OR NEW.description IS DISTINCT FROM OLD.description
    OR NEW.category IS DISTINCT FROM OLD.category
    OR NEW.icon IS DISTINCT FROM OLD.icon
    OR NEW.tags IS DISTINCT FROM OLD.tags
    OR NEW.outcomes IS DISTINCT FROM OLD.outcomes
    OR NEW.featured IS DISTINCT FROM OLD.featured
    OR NEW.expert_snapshot IS DISTINCT FROM OLD.expert_snapshot
    OR NEW.worker_spec_snapshot IS DISTINCT FROM OLD.worker_spec_snapshot
    OR NEW.skill_dependencies IS DISTINCT FROM OLD.skill_dependencies
    OR NEW.created_at IS DISTINCT FROM OLD.created_at
  THEN
    RAISE EXCEPTION 'expert market release immutable fields cannot be updated';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER expert_market_releases_immutable
BEFORE UPDATE ON expert_market_releases
FOR EACH ROW EXECUTE FUNCTION prevent_expert_market_release_immutable_update();

ALTER TABLE expert_market_applications
  ADD CONSTRAINT expert_market_applications_latest_release_fkey
  FOREIGN KEY (id, latest_published_release_id)
  REFERENCES expert_market_releases(application_id, id);

ALTER TABLE experts
  ADD COLUMN source_market_application_id BIGINT,
  ADD COLUMN source_market_release_id BIGINT,
  ADD CONSTRAINT experts_market_source_pair_check CHECK (
    (source_market_application_id IS NULL) =
    (source_market_release_id IS NULL)
  ),
  ADD CONSTRAINT experts_market_release_fkey
    FOREIGN KEY (source_market_application_id, source_market_release_id)
    REFERENCES expert_market_releases(application_id, id)
    ON DELETE SET NULL;

CREATE UNIQUE INDEX idx_experts_org_market_application
  ON experts(organization_id, source_market_application_id)
  WHERE source_market_application_id IS NOT NULL;

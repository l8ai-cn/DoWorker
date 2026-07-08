CREATE TABLE authored_skills (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    slug VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    license VARCHAR(100) NOT NULL DEFAULT '',

    git_repo_path VARCHAR(255) NOT NULL,
    default_branch VARCHAR(255) NOT NULL DEFAULT 'main',
    http_clone_url VARCHAR(1000),

    install_source VARCHAR(20) NOT NULL DEFAULT 'gitops',
    content_sha VARCHAR(64) NOT NULL DEFAULT '',
    storage_key VARCHAR(500) NOT NULL DEFAULT '',
    package_size BIGINT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,

    created_by_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100)
);

CREATE UNIQUE INDEX idx_authored_skills_org_slug ON authored_skills(organization_id, slug);
CREATE INDEX idx_authored_skills_org_updated ON authored_skills(organization_id, updated_at DESC);

COMMENT ON TABLE authored_skills IS 'DB cache/index for platform-authored (git-backed, namespace am-skills) skills. Git is the source of truth; this table backs List/Get and holds the packaged-artifact pointers (content_sha/storage_key). Coexists additively with the external-import/marketplace skill flow.';
COMMENT ON COLUMN authored_skills.git_repo_path IS 'am-skills/org<ID>-<slug>.';
COMMENT ON COLUMN authored_skills.install_source IS 'Provenance marker for the authoring source (always ''gitops'' for this table).';

CREATE TABLE marketplace.marketplace_quota_plans (
    id BIGSERIAL PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    slug VARCHAR(100) NOT NULL
        CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    period VARCHAR(16) NOT NULL CHECK (period IN ('monthly', 'total')),
    grant_credits NUMERIC(20,6) NOT NULL CHECK (grant_credits >= 0),
    charge_scope VARCHAR(16) NOT NULL
        CHECK (charge_scope IN ('marketplace', 'organization', 'group', 'user')),
    renewal_day SMALLINT,
    status VARCHAR(16) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'active', 'retired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (marketplace_id, slug),
    UNIQUE (marketplace_id, id),
    CHECK (
        (period = 'monthly' AND renewal_day BETWEEN 1 AND 28)
        OR (period = 'total' AND renewal_day IS NULL)
    )
);

CREATE TABLE marketplace.marketplace_quota_accounts (
    id UUID PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    subject_type VARCHAR(16) NOT NULL
        CHECK (subject_type IN ('marketplace', 'organization', 'group', 'user')),
    subject_ref BIGINT NOT NULL CHECK (subject_ref > 0),
    quota_plan_id BIGINT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'closed')),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (marketplace_id, subject_type, subject_ref, quota_plan_id),
    UNIQUE (marketplace_id, id),
    CHECK (period_end > period_start),
    FOREIGN KEY (marketplace_id, quota_plan_id)
        REFERENCES marketplace.marketplace_quota_plans(marketplace_id, id)
);

-- Quota management + billing module.
--
-- virtual_api_keys: a platform-issued handle (dwk_...) that wraps a real
-- ai_models credential and carries an independent per-Worker token budget.
-- Only the sha256 hash is stored; the plaintext token is shown once at
-- creation. A Worker (pod) binds a virtual key for usage attribution.
--
-- token_quotas: org-level (user_id NULL) and per-user token ceilings,
-- optionally scoped to a single model. Report-only: consumption is
-- aggregated on read from pod_session_usage and compared to the limit.

CREATE TABLE virtual_api_keys (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ai_model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    key_hash VARCHAR(64) NOT NULL,

    token_budget BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT virtual_api_keys_status_ck CHECK (status IN ('active', 'revoked', 'exhausted'))
);

CREATE UNIQUE INDEX uq_virtual_api_keys_hash ON virtual_api_keys(key_hash);
CREATE INDEX idx_virtual_api_keys_org ON virtual_api_keys(organization_id);
CREATE INDEX idx_virtual_api_keys_user ON virtual_api_keys(user_id);
CREATE INDEX idx_virtual_api_keys_model ON virtual_api_keys(ai_model_id);

CREATE TABLE token_quotas (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    model VARCHAR(200),

    limit_tokens BIGINT NOT NULL,
    period VARCHAR(20) NOT NULL DEFAULT 'total',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT token_quotas_period_ck CHECK (period IN ('total', 'monthly')),
    CONSTRAINT token_quotas_limit_ck CHECK (limit_tokens >= 0)
);

-- One quota row per (org, user-or-org-wide, model-or-all).
CREATE UNIQUE INDEX uq_token_quotas_scope
    ON token_quotas(organization_id, COALESCE(user_id, 0), COALESCE(model, ''));

ALTER TABLE pods ADD COLUMN virtual_api_key_id BIGINT
    REFERENCES virtual_api_keys(id) ON DELETE SET NULL;
CREATE INDEX idx_pods_virtual_api_key ON pods(virtual_api_key_id)
    WHERE virtual_api_key_id IS NOT NULL;

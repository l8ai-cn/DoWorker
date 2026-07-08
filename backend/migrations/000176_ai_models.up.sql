-- Unified model pool: one row per configured model a Worker can be launched with.
-- Supersedes the user-only `user_ai_providers` semantics with an org/user
-- two-tier pool (organization_id NULL => user-private, user_id NULL => org-shared).
-- Credentials are encrypted at the service layer (same Encryptor as user_ai_providers).

CREATE TABLE ai_models (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT REFERENCES organizations(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    provider_type VARCHAR(50) NOT NULL,
    model VARCHAR(200) NOT NULL,
    base_url VARCHAR(500) NOT NULL DEFAULT '',

    encrypted_credentials TEXT NOT NULL DEFAULT '',

    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    token_budget BIGINT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ai_models_scope_ck CHECK (organization_id IS NOT NULL OR user_id IS NOT NULL)
);

CREATE INDEX idx_ai_models_org ON ai_models(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX idx_ai_models_user ON ai_models(user_id) WHERE user_id IS NOT NULL;

CREATE TABLE provider_connections (
    id BIGSERIAL PRIMARY KEY,
    owner_scope VARCHAR(16) NOT NULL CHECK (owner_scope IN ('user', 'org')),
    owner_id BIGINT NOT NULL CHECK (owner_id > 0),
    identifier VARCHAR(100) NOT NULL CHECK (
        identifier ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
        AND char_length(identifier) BETWEEN 2 AND 100
    ),
    provider_key VARCHAR(100) NOT NULL CHECK (
        provider_key ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
        AND char_length(provider_key) BETWEEN 2 AND 100
    ),
    name VARCHAR(200) NOT NULL,
    base_url VARCHAR(1000) NOT NULL DEFAULT '',
    credentials_encrypted TEXT NOT NULL DEFAULT '',
    configured_fields JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(configured_fields) = 'array'),
    status VARCHAR(16) NOT NULL DEFAULT 'unchecked' CHECK (status IN ('unchecked', 'valid', 'invalid')),
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_validated_at TIMESTAMPTZ,
    validation_error TEXT NOT NULL DEFAULT '',
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_scope, owner_id, identifier)
);

CREATE INDEX idx_provider_connections_owner ON provider_connections(owner_scope, owner_id);
CREATE INDEX idx_provider_connections_provider ON provider_connections(provider_key);
CREATE INDEX idx_provider_connections_enabled ON provider_connections(is_enabled) WHERE is_enabled;

CREATE FUNCTION enforce_provider_connection_owner() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.owner_scope = 'user' THEN
        PERFORM 1 FROM users WHERE id = NEW.owner_id FOR KEY SHARE;
    ELSE
        PERFORM 1 FROM organizations WHERE id = NEW.owner_id FOR KEY SHARE;
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'provider connection % owner % does not exist', NEW.owner_scope, NEW.owner_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_provider_connection_owner
BEFORE INSERT ON provider_connections
FOR EACH ROW EXECUTE FUNCTION enforce_provider_connection_owner();

CREATE FUNCTION prevent_ai_resource_owner_delete() RETURNS TRIGGER AS $$
DECLARE
    deleting_scope VARCHAR(16);
BEGIN
    IF TG_TABLE_NAME = 'users' THEN
        deleting_scope := 'user';
    ELSE
        deleting_scope := 'org';
    END IF;
    IF EXISTS (
        SELECT 1 FROM provider_connections
         WHERE owner_scope = deleting_scope AND owner_id = OLD.id
    ) THEN
        RAISE EXCEPTION 'cannot delete % % with provider connections', deleting_scope, OLD.id;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_user_with_provider_connections_delete
BEFORE DELETE ON users
FOR EACH ROW EXECUTE FUNCTION prevent_ai_resource_owner_delete();

CREATE TRIGGER prevent_org_with_provider_connections_delete
BEFORE DELETE ON organizations
FOR EACH ROW EXECUTE FUNCTION prevent_ai_resource_owner_delete();

CREATE TABLE model_resources (
    id BIGSERIAL PRIMARY KEY,
    provider_connection_id BIGINT NOT NULL REFERENCES provider_connections(id) ON DELETE CASCADE,
    identifier VARCHAR(100) NOT NULL CHECK (
        identifier ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
        AND char_length(identifier) BETWEEN 2 AND 100
    ),
    model_id VARCHAR(500) NOT NULL CHECK (char_length(btrim(model_id)) > 0),
    display_name VARCHAR(200) NOT NULL,
    modalities JSONB NOT NULL CHECK (jsonb_typeof(modalities) = 'array' AND jsonb_array_length(modalities) > 0),
    capabilities JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(capabilities) = 'array'),
    status VARCHAR(16) NOT NULL DEFAULT 'unchecked' CHECK (status IN ('unchecked', 'valid', 'invalid')),
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_validated_at TIMESTAMPTZ,
    validation_error TEXT NOT NULL DEFAULT '',
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider_connection_id, identifier)
);

CREATE INDEX idx_model_resources_connection ON model_resources(provider_connection_id);
CREATE INDEX idx_model_resources_enabled ON model_resources(is_enabled) WHERE is_enabled;
CREATE INDEX idx_model_resources_modalities ON model_resources USING GIN(modalities);

CREATE TABLE model_resource_defaults (
    owner_scope VARCHAR(16) NOT NULL CHECK (owner_scope IN ('user', 'org')),
    owner_id BIGINT NOT NULL CHECK (owner_id > 0),
    modality VARCHAR(16) NOT NULL CHECK (modality IN ('chat', 'image', 'audio', 'video', 'embedding', 'multimodal')),
    model_resource_id BIGINT NOT NULL REFERENCES model_resources(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (owner_scope, owner_id, modality)
);

CREATE INDEX idx_model_resource_defaults_resource ON model_resource_defaults(model_resource_id);

CREATE FUNCTION enforce_model_resource_default() RETURNS TRIGGER AS $$
DECLARE
    resource_owner_scope VARCHAR(16);
    resource_owner_id BIGINT;
    resource_modalities JSONB;
BEGIN
    SELECT connection.owner_scope, connection.owner_id, resource.modalities
      INTO resource_owner_scope, resource_owner_id, resource_modalities
      FROM model_resources resource
      JOIN provider_connections connection ON connection.id = resource.provider_connection_id
     WHERE resource.id = NEW.model_resource_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'model resource % does not exist', NEW.model_resource_id;
    END IF;
    IF resource_owner_scope <> NEW.owner_scope OR resource_owner_id <> NEW.owner_id THEN
        RAISE EXCEPTION 'model resource default owner does not match connection owner';
    END IF;
    IF NOT (resource_modalities ? NEW.modality) THEN
        RAISE EXCEPTION 'model resource % does not support modality %', NEW.model_resource_id, NEW.modality;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_model_resource_default
BEFORE INSERT OR UPDATE ON model_resource_defaults
FOR EACH ROW EXECUTE FUNCTION enforce_model_resource_default();

CREATE FUNCTION keep_ai_resource_parent_invariants() RETURNS TRIGGER AS $$
BEGIN
    IF TG_TABLE_NAME = 'provider_connections' THEN
        IF NEW.owner_scope <> OLD.owner_scope OR NEW.owner_id <> OLD.owner_id THEN
            RAISE EXCEPTION 'provider connection owner is immutable';
        END IF;
    ELSE
        IF NEW.provider_connection_id <> OLD.provider_connection_id THEN
            RAISE EXCEPTION 'model resource connection is immutable';
        END IF;
        IF EXISTS (
            SELECT 1 FROM model_resource_defaults defaults
             WHERE defaults.model_resource_id = OLD.id
               AND NOT (NEW.modalities ? defaults.modality)
        ) THEN
            RAISE EXCEPTION 'cannot remove a modality with an active default';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER keep_provider_connection_owner
BEFORE UPDATE ON provider_connections
FOR EACH ROW EXECUTE FUNCTION keep_ai_resource_parent_invariants();

CREATE TRIGGER keep_model_resource_parent
BEFORE UPDATE ON model_resources
FOR EACH ROW EXECUTE FUNCTION keep_ai_resource_parent_invariants();

CREATE TABLE ai_resource_migration_map (
    id BIGSERIAL PRIMARY KEY,
    source_kind VARCHAR(32) NOT NULL CHECK (source_kind IN ('ai_model', 'env_bundle')),
    source_id BIGINT NOT NULL CHECK (source_id > 0),
    provider_connection_id BIGINT REFERENCES provider_connections(id) ON DELETE SET NULL,
    model_resource_id BIGINT REFERENCES model_resources(id) ON DELETE SET NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'migrated', 'error')),
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_kind, source_id)
);

CREATE INDEX idx_ai_resource_migration_connection ON ai_resource_migration_map(provider_connection_id);
CREATE INDEX idx_ai_resource_migration_resource ON ai_resource_migration_map(model_resource_id);

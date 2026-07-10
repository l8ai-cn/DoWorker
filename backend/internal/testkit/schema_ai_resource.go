package testkit

func aiResourceTableDDLs() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS provider_connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_scope TEXT NOT NULL CHECK (owner_scope IN ('user', 'org')),
			owner_id INTEGER NOT NULL CHECK (owner_id > 0),
			identifier TEXT NOT NULL,
			provider_key TEXT NOT NULL,
			name TEXT NOT NULL,
			base_url TEXT NOT NULL DEFAULT '',
			credentials_encrypted TEXT NOT NULL DEFAULT '',
			configured_fields TEXT NOT NULL DEFAULT '[]',
			status TEXT NOT NULL DEFAULT 'unchecked',
			is_enabled INTEGER NOT NULL DEFAULT 1,
			last_validated_at DATETIME,
			validation_error TEXT NOT NULL DEFAULT '',
			revision INTEGER NOT NULL DEFAULT 1 CHECK (revision > 0),
			created_by INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(owner_scope, owner_id, identifier)
		)`,
		`CREATE TABLE IF NOT EXISTS model_resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			provider_connection_id INTEGER NOT NULL,
			identifier TEXT NOT NULL,
			model_id TEXT NOT NULL,
			display_name TEXT NOT NULL,
			modalities TEXT NOT NULL DEFAULT '[]',
			capabilities TEXT NOT NULL DEFAULT '[]',
			status TEXT NOT NULL DEFAULT 'unchecked',
			is_enabled INTEGER NOT NULL DEFAULT 1,
			last_validated_at DATETIME,
			validation_error TEXT NOT NULL DEFAULT '',
			revision INTEGER NOT NULL DEFAULT 1 CHECK (revision > 0),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(provider_connection_id, identifier)
		)`,
		`CREATE TABLE IF NOT EXISTS model_resource_defaults (
			owner_scope TEXT NOT NULL,
			owner_id INTEGER NOT NULL,
			modality TEXT NOT NULL,
			model_resource_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(owner_scope, owner_id, modality)
		)`,
		`CREATE TRIGGER IF NOT EXISTS validate_model_resource_default
			BEFORE INSERT ON model_resource_defaults
			BEGIN
				SELECT CASE WHEN NOT EXISTS (
					SELECT 1 FROM model_resources mr
					JOIN provider_connections pc ON pc.id = mr.provider_connection_id
					JOIN json_each(mr.modalities) modalities ON modalities.value = NEW.modality
					WHERE mr.id = NEW.model_resource_id
						AND pc.owner_scope = NEW.owner_scope
						AND pc.owner_id = NEW.owner_id
				) THEN RAISE(ABORT, 'invalid model resource default') END;
			END`,
		`CREATE TRIGGER IF NOT EXISTS validate_model_resource_default_update
			BEFORE UPDATE ON model_resource_defaults
			BEGIN
				SELECT CASE WHEN NOT EXISTS (
					SELECT 1 FROM model_resources mr
					JOIN provider_connections pc ON pc.id = mr.provider_connection_id
					JOIN json_each(mr.modalities) modalities ON modalities.value = NEW.modality
					WHERE mr.id = NEW.model_resource_id
						AND pc.owner_scope = NEW.owner_scope
						AND pc.owner_id = NEW.owner_id
				) THEN RAISE(ABORT, 'invalid model resource default') END;
			END`,
		`CREATE TABLE IF NOT EXISTS ai_resource_migration_map (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_kind TEXT NOT NULL,
			source_id INTEGER NOT NULL,
			provider_connection_id INTEGER,
			model_resource_id INTEGER,
			status TEXT NOT NULL DEFAULT 'pending',
			error_message TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source_kind, source_id)
		)`,
		`CREATE TRIGGER IF NOT EXISTS validate_provider_connection_owner
			BEFORE INSERT ON provider_connections
			BEGIN
				SELECT CASE
					WHEN NEW.owner_scope = 'user' AND NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.owner_id)
						THEN RAISE(ABORT, 'provider connection user owner does not exist')
					WHEN NEW.owner_scope = 'org' AND NOT EXISTS (SELECT 1 FROM organizations WHERE id = NEW.owner_id)
						THEN RAISE(ABORT, 'provider connection org owner does not exist')
				END;
			END`,
		`CREATE TRIGGER IF NOT EXISTS prevent_user_with_provider_connections_delete
			BEFORE DELETE ON users
			WHEN EXISTS (
				SELECT 1 FROM provider_connections
				WHERE owner_scope = 'user' AND owner_id = OLD.id
			)
			BEGIN
				SELECT RAISE(ABORT, 'cannot delete an AI resource owner');
			END`,
		`CREATE TRIGGER IF NOT EXISTS prevent_org_with_provider_connections_delete
			BEFORE DELETE ON organizations
			WHEN EXISTS (
				SELECT 1 FROM provider_connections
				WHERE owner_scope = 'org' AND owner_id = OLD.id
			)
			BEGIN
				SELECT RAISE(ABORT, 'cannot delete an AI resource owner');
			END`,
		`CREATE TRIGGER IF NOT EXISTS keep_provider_connection_owner
			BEFORE UPDATE OF owner_scope, owner_id ON provider_connections
			WHEN OLD.owner_scope <> NEW.owner_scope OR OLD.owner_id <> NEW.owner_id
			BEGIN
				SELECT RAISE(ABORT, 'provider connection owner is immutable');
			END`,
		`CREATE TRIGGER IF NOT EXISTS keep_model_resource_connection
			BEFORE UPDATE OF provider_connection_id ON model_resources
			WHEN OLD.provider_connection_id <> NEW.provider_connection_id
			BEGIN
				SELECT RAISE(ABORT, 'model resource connection is immutable');
			END`,
		`CREATE TRIGGER IF NOT EXISTS keep_defaulted_modalities
			BEFORE UPDATE OF modalities ON model_resources
			WHEN EXISTS (
				SELECT 1 FROM model_resource_defaults defaults
				WHERE defaults.model_resource_id = OLD.id
					AND NOT EXISTS (
						SELECT 1 FROM json_each(NEW.modalities) modalities
						WHERE modalities.value = defaults.modality
					)
			)
			BEGIN
				SELECT RAISE(ABORT, 'cannot remove a defaulted modality');
			END`,
	}
}

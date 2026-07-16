DO $$
DECLARE
  adapter_column_exists BOOLEAN;
  actual_fingerprint TEXT;
  migration_dirty BOOLEAN;
  migration_version INTEGER;
  unknown_agent_count INTEGER;
BEGIN
  IF (SELECT count(*) FROM schema_migrations) <> 1 THEN
    RAISE EXCEPTION 'expected exactly one schema_migrations row';
  END IF;
  SELECT version, dirty INTO migration_version, migration_dirty
  FROM schema_migrations;
  SELECT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'agents'
      AND column_name = 'adapter_id'
  ) INTO adapter_column_exists;
  IF migration_dirty AND migration_version NOT IN (208, 222) THEN
    RAISE EXCEPTION 'refusing unrelated dirty migration %', migration_version;
  END IF;
  IF NOT migration_dirty
    AND (migration_version < 208 OR migration_version > 222) THEN
    RAISE EXCEPTION 'unexpected clean migration version %', migration_version;
  END IF;
  IF NOT migration_dirty AND NOT adapter_column_exists THEN
    RAISE EXCEPTION 'clean migration state is missing agents.adapter_id';
  END IF;
  IF migration_dirty AND migration_version = 222
    AND (
      NOT adapter_column_exists
      OR EXISTS (SELECT 1 FROM agents WHERE slug = 'video-studio')
    ) THEN
    RAISE EXCEPTION 'dirty migration 222 is not in the expected rolled-back state';
  END IF;

  SELECT md5(pg_get_constraintdef(constraint_record.oid))
  INTO actual_fingerprint
  FROM pg_constraint constraint_record
  WHERE constraint_record.conrelid = 'pod_config_revisions'::regclass
    AND constraint_record.conname = 'pod_config_revisions_status_check';
  IF actual_fingerprint IS DISTINCT FROM
    '44e380da5f99bf816e907093d03bc24a' THEN
    RAISE EXCEPTION 'migration 205 constraint contract is incomplete';
  END IF;

  SELECT md5(string_agg(format(
    '%s|%s|%s|%s|%s|%s|%s',
    table_name, ordinal_position, column_name, data_type, udt_name,
    is_nullable, coalesce(column_default, '')
  ), E'\n' ORDER BY table_name, ordinal_position))
  INTO actual_fingerprint
  FROM information_schema.columns
  WHERE table_schema = 'public'
    AND (
      table_name = 'execution_clusters'
      OR (table_name = 'runners' AND column_name IN (
        'cluster_id', 'tunnel_state', 'tunnel_last_seen_at', 'tunnel_last_error'
      ))
      OR (table_name = 'pods' AND column_name = 'cluster_id')
      OR (table_name = 'runner_grpc_registration_tokens'
        AND column_name = 'cluster_id')
      OR (table_name = 'runner_pending_auths'
        AND column_name IN ('cluster_id', 'authorized'))
    );
  IF actual_fingerprint IS DISTINCT FROM
    '67b424efb3a5b844df0184388b3cf822' THEN
    RAISE EXCEPTION 'migration 206 column contract is incomplete';
  END IF;

  SELECT md5(string_agg(format(
    '%s|%s|%s',
    constraint_record.conrelid::regclass::text,
    constraint_record.conname,
    pg_get_constraintdef(constraint_record.oid)
  ), E'\n' ORDER BY
    constraint_record.conrelid::regclass::text,
    constraint_record.conname))
  INTO actual_fingerprint
  FROM pg_constraint constraint_record
  WHERE constraint_record.conrelid IN (
    'execution_clusters'::regclass,
    'runners'::regclass,
    'pods'::regclass,
    'runner_grpc_registration_tokens'::regclass,
    'runner_pending_auths'::regclass
  )
    AND (
      constraint_record.conrelid = 'execution_clusters'::regclass
      OR constraint_record.conname IN (
        'runners_execution_cluster_organization_fkey',
        'pods_execution_cluster_organization_fkey',
        'runner_grpc_registration_tokens_cluster_organization_fkey',
        'runner_pending_auths_cluster_organization_fkey',
        'runner_pending_auths_cluster_ownership_check'
      )
    );
  IF actual_fingerprint IS DISTINCT FROM
    '8248044a68445136126905472f6fbc02' THEN
    RAISE EXCEPTION 'migration 206 constraint contract is incomplete';
  END IF;

  SELECT md5(string_agg(format(
    '%s|%s|%s', tablename, indexname, indexdef
  ), E'\n' ORDER BY tablename, indexname))
  INTO actual_fingerprint
  FROM pg_indexes
  WHERE schemaname = 'public'
    AND (
      tablename = 'execution_clusters'
      OR indexname IN (
        'idx_runners_cluster_id',
        'idx_pods_cluster_id',
        'idx_runner_grpc_registration_tokens_cluster_id',
        'idx_runner_pending_auths_cluster_id'
      )
    );
  IF actual_fingerprint IS DISTINCT FROM
    'faba46db825a34e91fe33398ba447ccd' THEN
    RAISE EXCEPTION 'migration 206 index contract is incomplete';
  END IF;
  IF EXISTS (SELECT 1 FROM runners WHERE cluster_id IS NULL)
    OR EXISTS (SELECT 1 FROM pods WHERE cluster_id IS NULL)
    OR EXISTS (
      SELECT 1 FROM runner_grpc_registration_tokens WHERE cluster_id IS NULL
    ) THEN
    RAISE EXCEPTION 'migration 206 cluster backfill is incomplete';
  END IF;
  IF EXISTS (
    SELECT 1 FROM organizations organization_record
    WHERE NOT EXISTS (
      SELECT 1 FROM execution_clusters cluster_record
      WHERE cluster_record.organization_id = organization_record.id
        AND cluster_record.slug = 'online'
    ) OR NOT EXISTS (
      SELECT 1 FROM execution_clusters cluster_record
      WHERE cluster_record.organization_id = organization_record.id
        AND cluster_record.slug = 'local'
    )
  ) THEN
    RAISE EXCEPTION 'migration 206 organization cluster backfill is incomplete';
  END IF;
  IF EXISTS (
    SELECT 1
    FROM runners runner_record
    JOIN execution_clusters cluster_record
      ON cluster_record.id = runner_record.cluster_id
    WHERE cluster_record.slug <> 'online'
  ) THEN
    RAISE EXCEPTION 'migration 206 runner cluster mapping is incorrect';
  END IF;
  IF EXISTS (
    SELECT 1
    FROM pods pod_record
    JOIN runners runner_record ON runner_record.id = pod_record.runner_id
    WHERE pod_record.cluster_id IS DISTINCT FROM runner_record.cluster_id
  ) THEN
    RAISE EXCEPTION 'migration 206 pod cluster mapping is incorrect';
  END IF;
  IF EXISTS (
    SELECT 1
    FROM runner_grpc_registration_tokens token_record
    JOIN execution_clusters cluster_record
      ON cluster_record.id = token_record.cluster_id
    WHERE cluster_record.slug <> 'local'
  ) THEN
    RAISE EXCEPTION 'migration 206 registration token mapping is incorrect';
  END IF;
  IF EXISTS (
    SELECT 1
    FROM runner_pending_auths pending_record
    JOIN runners runner_record ON runner_record.id = pending_record.runner_id
    WHERE pending_record.cluster_id IS DISTINCT FROM runner_record.cluster_id
      OR pending_record.organization_id IS DISTINCT FROM runner_record.organization_id
  ) OR EXISTS (
    SELECT 1
    FROM runner_pending_auths pending_record
    JOIN execution_clusters cluster_record
      ON cluster_record.id = pending_record.cluster_id
    WHERE pending_record.runner_id IS NULL
      AND cluster_record.slug <> 'local'
  ) THEN
    RAISE EXCEPTION 'migration 206 pending auth mapping is incorrect';
  END IF;

  IF NOT adapter_column_exists THEN
    SELECT count(*) INTO unknown_agent_count
    FROM agents
    WHERE slug NOT IN (
      'aider', 'claude-code', 'codex-cli', 'cursor-cli', 'do-agent',
      'gemini-cli', 'grok-build', 'loopal', 'minimax-cli', 'opencode'
    );
    IF unknown_agent_count <> 0 THEN
      RAISE EXCEPTION 'migration 207 has no adapter mapping for % agents',
        unknown_agent_count;
    END IF;
  END IF;
END $$;

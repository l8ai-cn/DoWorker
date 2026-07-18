#!/usr/bin/env bash

backup_database() {
  echo "==> backup database before migrations"
  dexec "set -eu; backup_dir=/root/backups/agentsmesh; timestamp=\$(date -u +%Y%m%dT%H%M%SZ); backup=\${backup_dir}/pre-migrate-${RELEASE_DEPLOY_COMMIT:0:12}-\${timestamp}.dump; umask 077; mkdir -p \"\${backup_dir}\"; kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" exec pg_dump --format=custom --no-owner --no-privileges --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\"' > \"\${backup}.tmp\"; test -s \"\${backup}.tmp\"; mv \"\${backup}.tmp\" \"\${backup}\"; sha256sum \"\${backup}\" > \"\${backup}.sha256\"; echo \"database backup: \${backup}\""
}

require_clean_migration_state() {
  local state
  state="$(dexec "kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" psql --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\" --tuples-only --no-align --command=\"SELECT version, dirty FROM schema_migrations\"'" |
    tail -n 1 | tr -d '\r')"
  [[ "${state}" =~ ^[0-9]+\|f$ ]] || {
    echo "database migration state must be clean before deploy, got ${state}" >&2
    return 1
  }
  echo "==> current migration state ${state}"
}

require_empty_pending_commands() {
  [[ "${APP_WRITES_STOPPED}" == true ]] || return
  local count
  count="$(dexec "kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" psql --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\" --tuples-only --no-align --command=\"SELECT COUNT(*) FROM pending_runner_commands\"'" |
    tail -n 1 | tr -d '\r')"
  [[ "${count}" == "0" ]] || {
    echo "pending_runner_commands must be empty before encrypted-payload cutover, got ${count}" >&2
    return 1
  }
}

migrate_database() {
  echo "==> apply migration prerequisites"
  dexec "kubectl apply -f 02-configmap.yaml -f 30-backend-rbac.yaml"
  require_clean_migration_state
  backup_database
  apply_pinned_manifest "10-postgres.yaml" pgvector
  apply_pinned_manifest "11-redis.yaml" redis
  apply_pinned_manifest "12-minio.yaml" minio
  dexec "kubectl -n ${NS} rollout status deploy/postgres --timeout=300s"
  dexec "kubectl -n ${NS} delete job migrate --ignore-not-found"
  apply_backend_job "20-migrate-job.yaml"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/migrate --timeout=300s"
}

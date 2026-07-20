#!/usr/bin/env bash

latest_backend_migration_version() {
  find "${DIR}/../../../backend/migrations" -name '*.up.sql' -exec basename {} \; |
    awk -F_ '{ print $1 }' |
    sort -n |
    tail -1
}

require_dosql_database_evidence() {
  local expected_version
  expected_version="$(latest_backend_migration_version)"
  [[ -n "${DOSQL_RELEASE_DB_TARGET:-}" ]] || {
    echo "DOSQL_RELEASE_DB_TARGET is required for audited database changes" >&2
    return 1
  }
  [[ -n "${DOSQL_RELEASE_DB_SESSION:-}" ]] || {
    echo "DOSQL_RELEASE_DB_SESSION is required for audited database changes" >&2
    return 1
  }
  [[ "${DOSQL_RELEASE_MIGRATION_VERSION:-}" == "${expected_version}" ]] || {
    echo "DOSQL_RELEASE_MIGRATION_VERSION must equal latest backend migration ${expected_version}" >&2
    return 1
  }
  [[ -n "${DOSQL_RELEASE_CHANGE_ID:-}" ]] || {
    echo "DOSQL_RELEASE_CHANGE_ID is required for audited database changes" >&2
    return 1
  }
  echo "==> database schema/seed pre-applied by DoSql session ${DOSQL_RELEASE_DB_SESSION} on ${DOSQL_RELEASE_DB_TARGET}; change ${DOSQL_RELEASE_CHANGE_ID}; version ${expected_version}"
}

apply_stateful_prerequisites() {
  echo "==> apply stateful prerequisites"
  dexec "kubectl apply -f 02-configmap.yaml -f 30-backend-rbac.yaml"
  apply_pinned_manifest "10-postgres.yaml" pgvector
  apply_pinned_manifest "11-redis.yaml" redis
  apply_pinned_manifest "12-minio.yaml" minio
  dexec "kubectl -n ${NS} rollout status deploy/postgres --timeout=300s"
}

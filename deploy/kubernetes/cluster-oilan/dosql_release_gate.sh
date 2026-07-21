#!/usr/bin/env bash

latest_backend_migration_version() {
  find "${DIR}/../../../backend/migrations" -name '*.up.sql' -exec basename {} \; |
    awk -F_ '{ print $1 }' |
    sort -n |
    tail -1
}

require_text_env() {
  local name="$1"
  [[ -n "${!name:-}" ]] || {
    echo "${name} is required for audited database changes" >&2
    return 1
  }
}

require_dosql_database_evidence() {
  local expected_version repo_root audit_root
  expected_version="$(latest_backend_migration_version)"
  repo_root="$(cd "${DIR}/../../.." && pwd)"
  audit_root="${repo_root}/.dosql"
  require_text_env DOSQL_RELEASE_DB_TARGET
  require_text_env DOSQL_RELEASE_DB_MODE
  require_text_env DOSQL_RELEASE_DB_SESSION
  require_text_env DOSQL_RELEASE_CHANGE_ID
  require_text_env DOSQL_RELEASE_OPERATION_ID
  if [[ "${DOSQL_RELEASE_DB_TARGET}" != "db_agentcloud_prod_postgres" ]]; then
    echo "DOSQL_RELEASE_DB_TARGET must equal db_agentcloud_prod_postgres" >&2
    return 1
  fi
  if [[ -n "${DOSQL_RELEASE_MIGRATION_VERSION:-}" &&
      "${DOSQL_RELEASE_MIGRATION_VERSION}" != "${expected_version}" ]]; then
    echo "DOSQL_RELEASE_MIGRATION_VERSION must equal latest backend migration ${expected_version}" >&2
    return 1
  fi
  node "${DIR}/dosql_release_evidence_validate.mjs" \
    --target "${DOSQL_RELEASE_DB_TARGET}" \
    --mode "${DOSQL_RELEASE_DB_MODE}" \
    --session "${DOSQL_RELEASE_DB_SESSION}" \
    --change-id "${DOSQL_RELEASE_CHANGE_ID}" \
    --operation-id "${DOSQL_RELEASE_OPERATION_ID}" \
    --expected-version "${expected_version}" \
    --canonical-root "${audit_root}" \
    --journal "${audit_root}/changes/${DOSQL_RELEASE_CHANGE_ID}/journal.jsonl" \
    --evidence "${audit_root}/evidence/${DOSQL_RELEASE_OPERATION_ID}.json"
}

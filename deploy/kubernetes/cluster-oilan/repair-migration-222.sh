#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${DIR}/../../.." && pwd)"
TARGET="${DOOPS_TARGET:-gw-oilan-node}"
SESSION="${DOOPS_SESSION:-$(doops session | tr -d '[:space:]')}"
WS="/root/ws/${SESSION}"
NS=agentsmesh
ACK="${MIGRATION_REPAIR_ACK:-}"
REG="repo.aiedulab.cn:8443"
REMOTE_WORKSPACE_MAY_EXIST=false
REPAIR_BUNDLE=""
REPAIR_SUCCEEDED=false
BACKEND_REPLICAS=""
MARKETPLACE_REPLICAS=""

# shellcheck source=release_source_guard.sh
source "${DIR}/release_source_guard.sh"

dexec() {
  doops -session "${SESSION}" exec --target "${TARGET}" --cmd "cd ${WS} && $1"
}

cleanup() {
  local result=$?
  trap - EXIT
  [[ -z "${REPAIR_BUNDLE}" || ! -d "${REPAIR_BUNDLE}" ]] || rm -rf "${REPAIR_BUNDLE}"
  if [[ "${REMOTE_WORKSPACE_MAY_EXIST}" == true ]]; then
    doops -session "${SESSION}" clean --target "${TARGET}" --workspace "${SESSION}" || result=1
  fi
  if [[ "${REPAIR_SUCCEEDED}" != true && -n "${BACKEND_REPLICAS}" ]]; then
    echo "repair failed; application writes remain stopped for operator recovery" >&2
  fi
  exit "${result}"
}
trap cleanup EXIT

migration_state() {
  dexec "kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" psql --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\" --tuples-only --no-align --command=\"SELECT version, dirty FROM schema_migrations\"'" |
    tail -n 1 | tr -d '\r'
}

backup_database() {
  dexec "set -eu; dir=/root/backups/agentsmesh; stamp=\$(date -u +%Y%m%dT%H%M%SZ); file=\${dir}/pre-repair-222-$(git -C "${REPO_ROOT}" rev-parse --short=12 HEAD)-\${stamp}.dump; umask 077; mkdir -p \"\${dir}\"; kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" exec pg_dump --format=custom --no-owner --no-privileges --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\"' > \"\${file}.tmp\"; test -s \"\${file}.tmp\"; mv \"\${file}.tmp\" \"\${file}\"; sha256sum \"\${file}\" > \"\${file}.sha256\"; sha256sum -c \"\${file}.sha256\"; echo \"database backup: \${file}\""
}

push_repair_manifest() {
  REPAIR_BUNDLE="$(mktemp -d)"
  cp "${DIR}/24-migration-222-repair-job.yaml" "${REPAIR_BUNDLE}/"
  REMOTE_WORKSPACE_MAY_EXIST=true
  doops -session "${SESSION}" push --target "${TARGET}" --src "${REPAIR_BUNDLE}"
  rm -rf "${REPAIR_BUNDLE}"
  REPAIR_BUNDLE=""
}

main() {
  local backend_digest backend_image expected_version state video_count null_adapters
  [[ "${ACK}" == "repair-dirty-222-video-studio" ]] || {
    echo "set MIGRATION_REPAIR_ACK=repair-dirty-222-video-studio" >&2
    return 1
  }
  release_require_pushed_clean_tree "${REPO_ROOT}"
  release_verify_source_metadata "${REPO_ROOT}"
  release_verify_image_provenance "${REPO_ROOT}" backend
  state="$(migration_state)"
  [[ "${state}" == "222|t" ]] || {
    echo "expected migration state 222|t, got ${state}" >&2
    return 1
  }
  video_count="$(dexec "kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" psql --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\" --tuples-only --no-align --command=\"SELECT slug FROM agents ORDER BY slug\"' | grep -c '^video-studio\$' || true" | tail -n 1 | tr -d '\r')"
  [[ "${video_count}" == "0" ]] || {
    echo "video-studio already exists; refusing to rerun migration 222" >&2
    return 1
  }
  null_adapters="$(dexec "kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" psql --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\" --tuples-only --no-align --command=\"SELECT count(*) FROM agents WHERE adapter_id IS NULL\"'" | tail -n 1 | tr -d '\r')"
  [[ "${null_adapters}" == "0" ]] || {
    echo "agents.adapter_id contains null values" >&2
    return 1
  }

  BACKEND_REPLICAS="$(dexec "kubectl -n ${NS} get deploy backend -o jsonpath='{.spec.replicas}'" | tail -n 1 | tr -d '\r')"
  MARKETPLACE_REPLICAS="$(dexec "kubectl -n ${NS} get deploy marketplace -o jsonpath='{.spec.replicas}'" | tail -n 1 | tr -d '\r')"
  dexec "kubectl -n ${NS} scale deploy/backend deploy/marketplace --replicas=0"
  dexec "kubectl -n ${NS} wait --for=delete pod -l app=backend --timeout=180s"
  dexec "kubectl -n ${NS} wait --for=delete pod -l app=marketplace --timeout=180s"
  backup_database

  backend_digest="$(awk '$1 == "-" && $2 == "name:" && $3 ~ /agentsmesh\/backend$/ { found=1; next } found && $1 == "digest:" { print $2; exit }' "${DIR}/release/kustomization.yaml")"
  [[ "${backend_digest}" =~ ^sha256:[a-f0-9]{64}$ ]]
  backend_image="${REG}/agentsmesh/backend@${backend_digest}"
  expected_version="$(find "${REPO_ROOT}/backend/migrations" -name '*.up.sql' -exec basename {} \; |
    sed -E 's/^0*([0-9]+)_.*/\1/' | sort -n | tail -n 1)"
  push_repair_manifest
  dexec "kubectl -n ${NS} delete job migration-222-repair --ignore-not-found"
  dexec "sed -e 's|__BACKEND_IMAGE__|${backend_image}|g' -e 's|__BACKEND_DIGEST__|${backend_digest}|g' -e 's|__EXPECTED_VERSION__|${expected_version}|g' -e 's|__RELEASE_COMMIT__|$(git -C "${REPO_ROOT}" rev-parse HEAD)|g' 24-migration-222-repair-job.yaml | kubectl apply -f -"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/migration-222-repair --timeout=300s"
  dexec "kubectl -n ${NS} logs job/migration-222-repair"
  [[ "$(migration_state)" == "${expected_version}|f" ]]
  dexec "set -eu; kubectl -n ${NS} scale deploy/backend --replicas=${BACKEND_REPLICAS}; kubectl -n ${NS} scale deploy/marketplace --replicas=${MARKETPLACE_REPLICAS}; kubectl -n ${NS} rollout status deploy/backend --timeout=300s; kubectl -n ${NS} rollout status deploy/marketplace --timeout=300s"
  REPAIR_SUCCEEDED=true
}

main "$@"

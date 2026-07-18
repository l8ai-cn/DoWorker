#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"
WORKER_CATALOG="${REPO_ROOT}/config/worker-types/catalog.json"

cd "$REPO_ROOT"
go test ./backend/internal/service/workerdefinition ./backend/internal/service/workercreation \
  ./backend/internal/domain/workerruntime ./backend/cmd/server -count=1
bash "${LOOP_ROOT}/scripts/verify-dev-database-projection.sh"

jq -e '
  .schema_version == 1 and
  .status == "verified" and
  .definition_bundle == "definition.json + NUL + AgentFile" and
  (.blocked_registration_slugs == []) and
  (.runtime_catalog_lock == "backend/internal/domain/workerruntime/runtime_catalog.lock.json")
' "${LOOP_ROOT}/evidence/definition-chain.json" >/dev/null

expected_worker_slugs="$(jq -c '[.worker_types[].slug] | sort' "$WORKER_CATALOG")"
jq -e --argjson expected_worker_slugs "$expected_worker_slugs" '
  (.projection_verified_slugs | sort) == $expected_worker_slugs
' "${LOOP_ROOT}/evidence/definition-chain.json" >/dev/null

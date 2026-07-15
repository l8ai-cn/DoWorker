#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"

cd "$REPO_ROOT"
go test ./backend/internal/service/workerdefinition ./backend/internal/service/workercreation \
  ./backend/internal/domain/workerruntime ./backend/cmd/server -count=1

jq -e '
  .schema_version == 1 and
  .status == "verified" and
  .definition_bundle == "definition.json + NUL + AgentFile" and
  (.projection_verified_slugs | length == 12) and
  (.blocked_registration_slugs == []) and
  (.runtime_catalog_lock == "backend/internal/domain/workerruntime/runtime_catalog.lock.json")
' "${LOOP_ROOT}/evidence/definition-chain.json" >/dev/null

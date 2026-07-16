#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/worker-context.sh"
slug="$(worker_slug)"
definition="${LOOP_ROOT}/artifacts/definition.json"
contract="${LOOP_ROOT}/evidence/contracts/frontend-backend.json"
test_log="${LOOP_ROOT}/evidence/tests/frontend-backend.txt"
definition_hash="$(worker_definition_hash)"

[[ -s "$contract" ]] || {
  echo "missing frontend/backend contract evidence: $contract" >&2
  exit 1
}
[[ -s "$test_log" ]] || {
  echo "missing frontend/backend test log: $test_log" >&2
  exit 1
}
jq -e --arg slug "$slug" --arg definition_hash "$definition_hash" '
  .schema_version == 1 and
  .worker_slug == $slug and
  .definition_hash == $definition_hash and
  ([.api_surface[].operation] | sort) ==
    ["CreateWorker", "GetWorkerCreationOptions", "GetWorkerTypeDefinition", "ListWorkerTypes", "PreflightWorkerDraft"] and
  (.source_paths | type == "array" and length > 0) and
  (.form_states | sort) == ["disabled", "error", "incompatible", "loading", "success"]
' "$contract" >/dev/null

jq -r '.source_paths[]' "$contract" | while IFS= read -r path; do
  [[ "$path" != /* && "$path" != *".."* ]] || exit 1
  test -e "${REPO_ROOT}/${path}"
done

rg -qi 'definition_hash|definitionhash' "${REPO_ROOT}/backend" "${REPO_ROOT}/clients/core" "${REPO_ROOT}/clients/web"
rg -qi 'adapter_id|adapterid' "${REPO_ROOT}/backend" "${REPO_ROOT}/proto" "${REPO_ROOT}/clients/core" "${REPO_ROOT}/clients/web"
test ! -e "${REPO_ROOT}/clients/web/src/components/settings/envBundleCredentialForms/credentialBuiltinFallbacks.ts"
(cd "$REPO_ROOT" && go test ./backend/internal/service/workercreation -count=1)
(cd "$REPO_ROOT" && pnpm run web:typecheck)

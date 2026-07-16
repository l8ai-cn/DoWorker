#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <formal-worker-slug>" >&2
  exit 64
fi

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"
slug="$1"
run="${LOOP_ROOT}/runs/${slug}"
catalog="${REPO_ROOT}/config/worker-types/catalog.json"

grep -Fxq "$slug" "${LOOP_ROOT}/catalog/formal-worker-slugs.txt" || {
  echo "not a formal Worker slug: ${slug}" >&2
  exit 1
}
test -d "$run" || {
  echo "missing Worker Loop run: ${run}" >&2
  exit 1
}

expected_hash="$(jq -r --arg slug "$slug" '
  .worker_types[] | select(.slug == $slug) | .definition_hash
' "$catalog")"
[[ "$expected_hash" =~ ^sha256:[a-f0-9]{64}$ ]] || {
  echo "missing definition hash for ${slug}" >&2
  exit 1
}

jq -e --arg slug "$slug" --arg definition_hash "$expected_hash" '
  .schema_version == 1 and
  .slug == $slug and
  .definition_hash == $definition_hash and
  (.status | IN("verified_local_dev", "blocked"))
' "${run}/worker.json" >/dev/null

status="$(jq -r '.status' "${run}/worker.json")"
if [[ "$status" == "verified_local_dev" ]]; then
  jq -e '.status == "complete"' "${run}/state.json" >/dev/null
  ! grep -Eq -- '- \[ \]' "${run}/ACCEPTANCE.md"
  bash "${run}/scripts/verify.sh"
else
  jq -e '.status == "blocked"' "${run}/state.json" >/dev/null
  jq -e --arg slug "$slug" --arg definition_hash "$expected_hash" '
    .schema_version == 1 and
    .worker_slug == $slug and
    .definition_hash == $definition_hash and
    .status == "blocked" and
    (.blocker_code | type == "string" and test("^[a-z0-9]+(-[a-z0-9]+)*$")) and
    (.verifier_id | type == "string" and length > 0) and
    (.observed_at | type == "string" and length > 0) and
    (.evidence_refs | type == "array" and length > 0)
  ' "${run}/evidence/blocked.json" >/dev/null

  jq -r '.evidence_refs[]' "${run}/evidence/blocked.json" | while IFS= read -r ref; do
    [[ "$ref" != /* && "$ref" != *".."* && -s "${run}/${ref}" ]] || {
      echo "${slug}: invalid blocked evidence reference: ${ref}" >&2
      exit 1
    }
  done
fi

printf '%s\n' "$status"

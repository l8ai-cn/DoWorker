#!/usr/bin/env bash
set -euo pipefail

SOURCE_LOOP="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

REPO_ROOT="${TMP_ROOT}/repo"
LOOP_ROOT="${REPO_ROOT}/tools/loops/worker-onboarding/catalog-loop"
mkdir -p "${REPO_ROOT}/config/worker-types" "${LOOP_ROOT}/catalog" \
  "${LOOP_ROOT}/evidence" "${LOOP_ROOT}/runs" "${LOOP_ROOT}/scripts"

cp "${SOURCE_LOOP}/catalog/formal-worker-slugs.txt" "${LOOP_ROOT}/catalog/"
cp "${SOURCE_LOOP}/scripts/verify-worker-run.sh" "${LOOP_ROOT}/scripts/"
cp "${SOURCE_LOOP}/scripts/verify-worker-runs.sh" "${LOOP_ROOT}/scripts/"
cp "${SOURCE_LOOP}/../../../../config/worker-types/catalog.json" \
  "${REPO_ROOT}/config/worker-types/"

workers="$(
  while IFS= read -r slug; do
    definition_hash="$(jq -r --arg slug "$slug" '
      .worker_types[] | select(.slug == $slug) | .definition_hash
    ' "${REPO_ROOT}/config/worker-types/catalog.json")"
    run="${LOOP_ROOT}/runs/${slug}"
    mkdir -p "${run}/scripts" "${run}/evidence"
    jq -n --arg slug "$slug" --arg definition_hash "$definition_hash" '
      {
        schema_version: 1,
        slug: $slug,
        status: "verified_local_dev",
        definition_hash: $definition_hash
      }
    ' >"${run}/worker.json"
    jq -n '{status: "complete"}' >"${run}/state.json"
    : >"${run}/ACCEPTANCE.md"
    printf '%s\n' '#!/usr/bin/env bash' 'exit 0' >"${run}/scripts/verify.sh"
    chmod +x "${run}/scripts/verify.sh"
    jq -n --arg slug "$slug" '
      {slug: $slug, status: "verified_local_dev"}
    '
  done <"${LOOP_ROOT}/catalog/formal-worker-slugs.txt" | jq -s 'sort_by(.slug)'
)"

jq -n --argjson workers "$workers" '
  {schema_version: 1, status: "verified", workers: $workers}
' >"${LOOP_ROOT}/evidence/queue-summary.json"

bash "${LOOP_ROOT}/scripts/verify-worker-runs.sh" --all-verified

blocked_slug="gemini-cli"
blocked_run="${LOOP_ROOT}/runs/${blocked_slug}"
definition_hash="$(jq -r --arg slug "$blocked_slug" '
  .worker_types[] | select(.slug == $slug) | .definition_hash
' "${REPO_ROOT}/config/worker-types/catalog.json")"
printf '%s\n' 'missing authorized model resource' >"${blocked_run}/evidence/model-resource.txt"
jq -n --arg slug "$blocked_slug" --arg definition_hash "$definition_hash" '
  {
    schema_version: 1,
    slug: $slug,
    status: "blocked",
    definition_hash: $definition_hash
  }
' >"${blocked_run}/worker.json"
jq -n '{status: "blocked"}' >"${blocked_run}/state.json"
jq -n --arg slug "$blocked_slug" --arg definition_hash "$definition_hash" '
  {
    schema_version: 1,
    worker_slug: $slug,
    definition_hash: $definition_hash,
    status: "blocked",
    blocker_code: "missing-model-resource",
    verifier_id: "model-resource-guard",
    observed_at: "2026-07-12T00:00:00Z",
    evidence_refs: ["evidence/model-resource.txt"]
  }
' >"${blocked_run}/evidence/blocked.json"

processed_workers="$(
  jq --arg slug "$blocked_slug" '
    map(if .slug == $slug then .status = "blocked" else . end)
  ' <<<"$workers"
)"
jq -n --argjson workers "$processed_workers" '
  {schema_version: 1, status: "processed", workers: $workers}
' >"${LOOP_ROOT}/evidence/queue-summary.json"

bash "${LOOP_ROOT}/scripts/verify-worker-runs.sh" --processed
if bash "${LOOP_ROOT}/scripts/verify-worker-runs.sh" --all-verified; then
  echo "all-verified mode accepted a blocked Worker" >&2
  exit 1
fi

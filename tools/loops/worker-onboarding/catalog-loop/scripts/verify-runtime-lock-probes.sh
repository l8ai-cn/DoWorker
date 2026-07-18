#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(git -C "${LOOP_ROOT}" rev-parse --show-toplevel)"
LOCK="${REPO_ROOT}/backend/internal/domain/workerruntime/runtime_catalog.lock.json"
PROBES="${LOOP_ROOT}/evidence/runtime-lock-probes"

jq -r '.revision' "$LOCK" | grep -Eq '.+'
jq -c '.images[]' "$LOCK" | while IFS= read -r image; do
  reference="$(jq -r '.reference' <<<"$image")"
  digest="$(jq -r '.digest' <<<"$image")"
  jq -r '.worker_type_slugs[]' <<<"$image" | while IFS= read -r slug; do
    probe="${PROBES}/${slug}.json"
    jq -e --arg slug "$slug" --arg reference "$reference" --arg digest "$digest" '
      .schema_version == 1 and
      .worker_slug == $slug and
      .image_reference == $reference and
      .image_digest == $digest and
      (.status | IN("available", "not_found", "unavailable")) and
      (.exit_code | type == "number") and
      (.output | type == "string") and
      (.observed_at | type == "string" and length > 0)
    ' "$probe" >/dev/null
    printf '%s: runtime lock probe verified\n' "$slug"
  done
done

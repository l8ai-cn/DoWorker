#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"
SLUGS="${LOOP_ROOT}/catalog/formal-worker-slugs.txt"
RUNS="${LOOP_ROOT}/runs"
CATALOG="${REPO_ROOT}/config/worker-types/catalog.json"

[[ -d "$RUNS" ]] || {
  echo "missing Worker Loop runs directory: $RUNS" >&2
  exit 1
}

expected="$(jq -Rsc 'split("\n") | map(select(length > 0)) | sort' "$SLUGS")"
actual="$(find "$RUNS" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; |
  jq -Rsc 'split("\n") | map(select(length > 0)) | sort')"
[[ "$actual" == "$expected" ]] || {
  echo "Worker run directories do not match the formal catalog" >&2
  exit 1
}

while IFS= read -r slug; do
  expected_hash="$(jq -r --arg slug "$slug" '
    .worker_types[] | select(.slug == $slug) | .definition_hash
  ' "$CATALOG")"
  run="${RUNS}/${slug}"
  jq -e --arg slug "$slug" --arg definition_hash "$expected_hash" '
    .schema_version == 1 and
    .slug == $slug and
    .status == "planned" and
    .definition_hash == $definition_hash and
    .source_refs == ["config/worker-types/catalog.json"]
  ' "${run}/worker.json" >/dev/null
  jq -e '.status == "planned"' "${run}/state.json" >/dev/null
done <"$SLUGS"

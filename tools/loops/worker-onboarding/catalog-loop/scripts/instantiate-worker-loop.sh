#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <formal-worker-slug>" >&2
  exit 64
fi

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ONBOARDING_ROOT="$(cd "${LOOP_ROOT}/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"
slug="$1"

grep -Fxq "$slug" "${LOOP_ROOT}/catalog/formal-worker-slugs.txt" || {
  echo "not a formal Worker slug: $slug" >&2
  exit 1
}

template="${ONBOARDING_ROOT}/worker-loop-template"
run="${LOOP_ROOT}/runs/${slug}"
definition_hash="$(jq -r --arg slug "$slug" '
  .worker_types[] | select(.slug == $slug) | .definition_hash
' "${REPO_ROOT}/config/worker-types/catalog.json")"
test -d "$template"
[[ "$definition_hash" =~ ^sha256:[a-f0-9]{64}$ ]] || {
  echo "missing definition hash for ${slug}" >&2
  exit 1
}
[[ ! -e "$run" ]] || {
  echo "Worker Loop run already exists: $run" >&2
  exit 1
}

mkdir -p "${LOOP_ROOT}/runs"
cp -R "$template" "$run"
jq -n --arg slug "$slug" --arg definition_hash "$definition_hash" \
  '{
    schema_version: 1,
    slug: $slug,
    status: "planned",
    definition_hash: $definition_hash,
    source_refs: ["config/worker-types/catalog.json"]
  }' \
  >"${run}/worker.json"
echo "created Worker Loop run: ${run}"

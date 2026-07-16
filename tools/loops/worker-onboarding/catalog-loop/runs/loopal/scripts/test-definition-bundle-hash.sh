#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/worker-context.sh"

for definition in "${REPO_ROOT}"/config/worker-types/*/definition.json; do
  slug="$(jq -r '.slug' "$definition")"
  expected="$(jq -r --arg slug "$slug" '
    .worker_types[] | select(.slug == $slug) | .definition_hash
  ' "${REPO_ROOT}/config/worker-types/catalog.json")"
  actual="$(definition_bundle_sha256 "$definition" "$(dirname "$definition")/AgentFile")"
  [[ "$actual" == "$expected" ]] || {
    echo "${slug}: bundle hash does not match the catalog" >&2
    exit 1
  }
done

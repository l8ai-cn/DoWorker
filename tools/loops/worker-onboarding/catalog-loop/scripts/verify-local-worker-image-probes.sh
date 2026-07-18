#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(git -C "${LOOP_ROOT}" rev-parse --show-toplevel)"
PROBES="${LOOP_ROOT}/evidence/local-image-probes"
CATALOG="${REPO_ROOT}/config/worker-types/catalog.json"
DEV_ENV="${REPO_ROOT}/deploy/dev/.env"
COMPOSE_PROJECT_NAME="$(sed -nE 's/^COMPOSE_PROJECT_NAME=([a-z0-9][a-z0-9_-]*)$/\1/p' "$DEV_ENV")"
[[ -n "$COMPOSE_PROJECT_NAME" ]] || {
  echo "deploy/dev/.env must define COMPOSE_PROJECT_NAME" >&2
  exit 1
}

while IFS= read -r slug; do
  probe="${PROBES}/${slug}.json"
  definition="${REPO_ROOT}/config/worker-types/${slug}/definition.json"
  test -s "$probe"
  definition_hash="$(jq -r --arg slug "$slug" '
    .worker_types[] | select(.slug == $slug) | .definition_hash
  ' "$CATALOG")"
  image_runtime="$(jq -r '.image.runtime' "$definition")"
  expected_image="${COMPOSE_PROJECT_NAME}-runner-${image_runtime}:latest"

  jq -e --arg slug "$slug" --arg hash "$definition_hash" --arg image "$expected_image" '
    .schema_version == 1 and
    .worker_slug == $slug and
    .definition_hash == $hash and
    .platform == "linux/amd64" and
    .image_reference == $image and
    (.probe_command | type == "array" and length >= 2) and
    (.status | IN("passed", "image_missing", "probe_failed")) and
    (.exit_code | type == "number") and
    (.output | type == "string") and
    (.observed_at | type == "string" and length > 0) and
    (if .status == "passed" then
      .exit_code == 0 and (.image_id | test("^sha256:[a-f0-9]{64}$"))
    else
      .exit_code != 0
    end)
  ' "$probe" >/dev/null
  printf '%s: local image probe evidence verified\n' "$slug"
done <"${LOOP_ROOT}/catalog/formal-worker-slugs.txt"

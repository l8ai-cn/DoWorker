#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"
SLUGS="${LOOP_ROOT}/catalog/formal-worker-slugs.txt"
CATALOG="${REPO_ROOT}/config/worker-types/catalog.json"
SCHEMA="${REPO_ROOT}/config/worker-types/schema/definition.schema.json"

require_nonempty() {
  [[ -s "$1" ]] || {
    echo "missing required artifact: $1" >&2
    exit 1
  }
}

require_nonempty "$CATALOG"
require_nonempty "$SCHEMA"
expected="$(jq -Rsc 'split("\n") | map(select(length > 0)) | sort' "$SLUGS")"

definition_bundle_sha256() {
  local definition="$1"
  local agentfile="$2"
  printf 'sha256:%s\n' "$(
    { cat "$definition"; printf '\0'; cat "$agentfile"; } |
      shasum -a 256 | awk '{print $1}'
  )"
}

jq -e --argjson expected "$expected" '
  .schema_version == 1 and
  (.worker_types | type == "array") and
  ([.worker_types[].slug] | sort) == $expected and
  all(.worker_types[];
    (.slug | type == "string") and
    (.definition_path | type == "string") and
    (.definition_hash | test("^sha256:[a-f0-9]{64}$"))
  )
' "$CATALOG" >/dev/null

jq -e '
  .type == "object" and
  (.properties | type == "object") and
  (.properties.adapter_id.type == "string") and
  (.properties.interaction_modes.type == "array") and
  (.properties.credential_bindings.type == "array") and
  (.properties.config_documents.type == "array")
' "$SCHEMA" >/dev/null

while IFS= read -r slug; do
  path="$(jq -r --arg slug "$slug" '.worker_types[] | select(.slug == $slug) | .definition_path' "$CATALOG")"
  [[ "$path" == "config/worker-types/"*/definition.json ]] || {
    echo "$slug has an invalid definition path: $path" >&2
    exit 1
  }
  definition="${REPO_ROOT}/${path}"
  agentfile="$(dirname "$definition")/AgentFile"
  require_nonempty "$definition"
  require_nonempty "$agentfile"
  expected_hash="$(definition_bundle_sha256 "$definition" "$agentfile")"
  actual_hash="$(jq -r --arg slug "$slug" '.worker_types[] | select(.slug == $slug) | .definition_hash' "$CATALOG")"
  [[ "$actual_hash" == "$expected_hash" ]] || {
    echo "$slug definition hash does not match ${path}" >&2
    exit 1
  }
  jq -e --arg slug "$slug" '
    .schema_version == 1 and
    .slug == $slug and
    (.definition_version | type == "string" and length > 0) and
    (.adapter_id | type == "string" and length > 0) and
    (.interaction_modes | type == "array" and length > 0) and
    (.credential_bindings | type == "array") and
    (.config_documents | type == "array")
  ' "$definition" >/dev/null
done <"$SLUGS"

#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/worker-context.sh"
slug="$(worker_slug)"
definition="${LOOP_ROOT}/artifacts/definition.json"
agentfile="${LOOP_ROOT}/artifacts/AgentFile"
schema="${LOOP_ROOT}/artifacts/schemas/definition.schema.json"
bindings="${LOOP_ROOT}/artifacts/credential-bindings.json"
documents="${LOOP_ROOT}/artifacts/config-documents.json"

require_nonempty() {
  [[ -s "$1" ]] || {
    echo "missing required artifact: $1" >&2
    exit 1
  }
}

require_nonempty "$definition"
require_nonempty "$agentfile"
require_nonempty "$schema"
require_nonempty "$bindings"
require_nonempty "$documents"
definition_hash="$(definition_bundle_sha256 "$definition" "$agentfile")"
[[ "$(worker_definition_hash)" == "$definition_hash" ]]

jq -e --arg slug "$slug" '
  .schema_version == 1 and
  .slug == $slug and
  (.definition_version | type == "string" and length > 0) and
  (.adapter_id | type == "string" and length > 0) and
  (.interaction_modes | type == "array" and length > 0) and
  all(.interaction_modes[]; IN("pty", "acp")) and
  (.credential_bindings | type == "array") and
  (.config_documents | type == "array")
' "$definition" >/dev/null

jq -e --arg slug "$slug" '
  .schema_version == 1 and
  .worker_slug == $slug and
  (.bindings | type == "array") and
  all(.bindings[];
    (.id | type == "string" and length > 0) and
    (.source.kind | IN("model_resource", "credential_bundle")) and
    (.source.ref | type == "string" and length > 0) and
    (.target.kind | IN("env", "config_document")) and
    (.target.name | type == "string" and length > 0)
  )
' "$bindings" >/dev/null

if jq -e '
  .. | objects | to_entries[] |
  select((.key | ascii_downcase) as $key |
    ["value", "password", "passwd", "token", "api_key", "apikey", "private_key", "secret"] |
    index($key)
  )
' "$bindings" >/dev/null; then
  echo "credential bindings must contain references, not plaintext values" >&2
  exit 1
fi

jq -e --arg slug "$slug" '
  .schema_version == 1 and
  .worker_slug == $slug and
  (.documents | type == "array") and
  all(.documents[];
    (.id | type == "string" and length > 0) and
    (.format | IN("json", "yaml", "toml", "text")) and
    (.target_path | type == "string" and length > 0)
  )
' "$documents" >/dev/null

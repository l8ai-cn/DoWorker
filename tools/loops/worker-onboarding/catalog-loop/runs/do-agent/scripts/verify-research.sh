#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/worker-context.sh"
slug="$(worker_slug)"
research="${LOOP_ROOT}/evidence/research/research.json"
sources="${LOOP_ROOT}/evidence/research/source-index.json"
required='["installation","version","license","protocol","authentication","configuration","persistence","platforms"]'

test -s "$research"
test -s "$sources"
jq -e --arg slug "$slug" --argjson required "$required" '
  .schema_version == 1 and
  .worker_slug == $slug and
  (.categories | type == "object") and
  (. as $document |
    all($required[];
      . as $category |
      ($document.categories[$category] |
        type == "object" and
        (.conclusion | type == "string" and length > 0 and . != "unknown") and
        (.evidence_refs | type == "array" and length > 0)
      )
    )
  )
' "$research" >/dev/null

jq -e '
  .schema_version == 1 and
  (.sources | type == "array" and length > 0) and
  all(.sources[];
    (.id | type == "string" and length > 0) and
    (.kind | IN("official", "captured_command")) and
    (.checked_at | type == "string" and length > 0) and
    (.evidence_hash | test("^sha256:[a-f0-9]{64}$")) and
    (if .kind == "official" then (.url | type == "string" and length > 0)
     else (.command | type == "string" and length > 0)
     end)
  )
' "$sources" >/dev/null

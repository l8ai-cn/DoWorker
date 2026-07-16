#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"
SLUGS="${LOOP_ROOT}/catalog/formal-worker-slugs.txt"
INVENTORY="${LOOP_ROOT}/catalog/inventory.json"
DRIFT="${LOOP_ROOT}/catalog/drift.json"

require_nonempty() {
  [[ -s "$1" ]] || {
    echo "missing required artifact: $1" >&2
    exit 1
  }
}

require_nonempty "$SLUGS"
if LC_ALL=C sort "$SLUGS" | uniq -d | grep -q .; then
  echo "formal Worker slugs must be unique" >&2
  exit 1
fi
while IFS= read -r slug; do
  [[ "$slug" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]] || {
    echo "invalid formal Worker slug: $slug" >&2
    exit 1
  }
done <"$SLUGS"

require_nonempty "$INVENTORY"
require_nonempty "$DRIFT"
expected="$(jq -Rsc 'split("\n") | map(select(length > 0)) | sort' "$SLUGS")"

jq -e --argjson expected "$expected" '
  .schema_version == 1 and
  (.workers | type == "array") and
  ([.workers[].slug] | sort) == $expected and
  all(.workers[];
    (.slug | type == "string") and
    (.source_refs | type == "array" and length > 0) and
    (.layers | type == "array" and length > 0) and
    (.definition_hash | test("^sha256:[a-f0-9]{64}$")) and
    (.executable | type == "string" and length > 0) and
    (.adapter_id | type == "string" and length > 0) and
    (.runtime_catalog.status | IN("locked_available", "blocked_no_published_digest", "invalid_published_digest")) and
    (.runtime_evidence | type == "string" and length > 0) and
    (.support_status | IN("verified_local_dev", "not_supported"))
  )
' "$INVENTORY" >/dev/null

jq -r '.workers[].source_refs[]' "$INVENTORY" | while IFS= read -r ref; do
  [[ "$ref" != /* && "$ref" != *".."* && -e "${REPO_ROOT}/${ref}" ]] || {
    echo "inventory has an invalid source reference: ${ref}" >&2
    exit 1
  }
done

blocked="$(jq -c '[.workers[] | select(.support_status != "verified_local_dev") | .slug] | sort' "$INVENTORY")"
jq -e --argjson expected "$blocked" '
  .schema_version == 1 and
  (.mismatches | type == "array") and
  ([.mismatches[].slug] | sort) == $expected and
  all(.mismatches[];
    (.slug | type == "string") and
    (.layer | type == "string") and
    (.observed | type == "string") and
    (.target | type == "string") and
    (.status == "blocked") and
    (.blocker_code | type == "string" and test("^[a-z0-9]+(-[a-z0-9]+)*$")) and
    (.evidence_refs | type == "array" and length > 0)
  )
' "$DRIFT" >/dev/null

jq -r '.mismatches[].evidence_refs[]' "$DRIFT" | while IFS= read -r ref; do
  [[ "$ref" != /* && "$ref" != *".."* && -s "${LOOP_ROOT}/${ref}" ]] || {
    echo "drift has an invalid evidence reference: ${ref}" >&2
    exit 1
  }
done

#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SLUGS="${LOOP_ROOT}/catalog/formal-worker-slugs.txt"
RUNS="${LOOP_ROOT}/runs"
SUMMARY="${LOOP_ROOT}/evidence/queue-summary.json"

if [[ $# -ne 1 || ( "$1" != "--processed" && "$1" != "--all-verified" ) ]]; then
  echo "usage: $0 --processed|--all-verified" >&2
  exit 64
fi
mode="$1"

[[ -d "$RUNS" ]] || {
  echo "missing Worker Loop runs directory: $RUNS" >&2
  exit 1
}
expected="$(jq -Rsc 'split("\n") | map(select(length > 0)) | sort' "$SLUGS")"
actual="$(find "$RUNS" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | jq -Rsc 'split("\n") | map(select(length > 0)) | sort')"
[[ "$actual" == "$expected" ]] || {
  echo "Worker run directories do not match the formal catalog" >&2
  exit 1
}

while IFS= read -r slug; do
  bash "${LOOP_ROOT}/scripts/verify-worker-run.sh" "$slug" >/dev/null
done <"$SLUGS"

[[ -s "$SUMMARY" ]] || {
  echo "missing queue summary: $SUMMARY" >&2
  exit 1
}

actual="$(
  while IFS= read -r slug; do
    status="$(bash "${LOOP_ROOT}/scripts/verify-worker-run.sh" "$slug")"
    jq -n --arg slug "$slug" --arg status "$status" '{slug: $slug, status: $status}'
  done <"$SLUGS" | jq -s 'sort_by(.slug)'
)"

if [[ "$mode" == "--all-verified" ]]; then
  jq -e --argjson expected "$expected" --argjson actual "$actual" '
    .schema_version == 1 and
    .status == "verified" and
    ([.workers[].slug] | sort) == $expected and
    (.workers | sort_by(.slug)) == $actual and
    all(.workers[]; .status == "verified_local_dev")
  ' "$SUMMARY" >/dev/null
  exit 0
fi

jq -e --argjson expected "$expected" --argjson actual "$actual" '
  .schema_version == 1 and
  .status == "processed" and
  ([.workers[].slug] | sort) == $expected and
  (.workers | sort_by(.slug)) == $actual and
  all(.workers[]; .status | IN("verified_local_dev", "blocked"))
' "$SUMMARY" >/dev/null

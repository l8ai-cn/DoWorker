#!/usr/bin/env bash

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(git -C "${LOOP_ROOT}" rev-parse --show-toplevel)"

worker_slug() {
  test -s "${LOOP_ROOT}/worker.json"
  local slug
  slug="$(jq -r '.slug' "${LOOP_ROOT}/worker.json")"
  [[ "$slug" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]] || {
    echo "worker.json must declare a concrete Worker slug" >&2
    return 1
  }
  printf '%s\n' "$slug"
}

worker_definition_hash() {
  jq -r '.definition_hash' "${LOOP_ROOT}/worker.json"
}

definition_bundle_sha256() {
  local definition="$1"
  local agentfile="$2"
  printf 'sha256:%s\n' "$(
    { cat "$definition"; printf '\0'; cat "$agentfile"; } |
      shasum -a 256 | awk '{print $1}'
  )"
}

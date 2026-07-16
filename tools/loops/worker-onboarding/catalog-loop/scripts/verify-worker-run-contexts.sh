#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(git -C "${LOOP_ROOT}" rev-parse --show-toplevel)"

while IFS= read -r slug; do
  run="${LOOP_ROOT}/runs/${slug}"
  actual="$(
    source "${run}/scripts/worker-context.sh"
    printf '%s' "$REPO_ROOT"
  )"
  [[ "$actual" == "$REPO_ROOT" ]] || {
    echo "${slug}: incorrect repository root: ${actual}" >&2
    exit 1
  }
  bash "${run}/scripts/test-definition-bundle-hash.sh"
  printf '%s: repository context verified\n' "$slug"
done <"${LOOP_ROOT}/catalog/formal-worker-slugs.txt"

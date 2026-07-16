#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

while IFS= read -r slug; do
  bash "${LOOP_ROOT}/runs/${slug}/scripts/verify-contract.sh"
  printf '%s: contract verified\n' "$slug"
done <"${LOOP_ROOT}/catalog/formal-worker-slugs.txt"

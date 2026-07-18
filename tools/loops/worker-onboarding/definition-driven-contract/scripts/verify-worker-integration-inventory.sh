#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
loop_root="$(cd "$script_dir/.." && pwd)"
inventory="$loop_root/evidence/worker-integration-inventory-2026-07-16.md"
temporary="$(mktemp)"

trap 'rm -f "$temporary"' EXIT

node "$script_dir/render-worker-integration-inventory.mjs" --output "$temporary" >/dev/null
cmp -s "$temporary" "$inventory"

printf 'Worker integration inventory is current\n'

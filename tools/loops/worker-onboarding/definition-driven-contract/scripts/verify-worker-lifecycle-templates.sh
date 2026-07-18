#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
loop_root="$(cd "$script_dir/.." && pwd)"
templates="$loop_root/evidence/worker-lifecycle-templates-2026-07-16.md"
temporary="$(mktemp)"

trap 'rm -f "$temporary"' EXIT

node "$script_dir/render-worker-lifecycle-templates.mjs" --output "$temporary" >/dev/null
cmp -s "$temporary" "$templates"

printf 'Worker lifecycle templates are current\n'

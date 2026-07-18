#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="$ROOT/deploy/dev"

source "$SCRIPT_DIR/lib/log.sh"
source "$SCRIPT_DIR/lib/worker_runtime_catalog.sh"

catalog="$(mktemp)"
trap 'rm -f "$catalog"' EXIT

cat >"$catalog" <<'JSON'
{
  "images": [
    {"worker_type_slugs": ["codex-cli"]},
    {"worker_type_slugs": ["gemini-cli"]},
    {"worker_type_slugs": ["minimax-cli"]},
    {"worker_type_slugs": ["openclaw"]},
    {"worker_type_slugs": ["do-agent", "seedance-expert"]},
    {"worker_type_slugs": ["loopal"]},
    {"worker_type_slugs": ["aider"]},
    {"worker_type_slugs": ["claude-code"]},
    {"worker_type_slugs": ["cursor-cli"]},
    {"worker_type_slugs": ["grok-build"]},
    {"worker_type_slugs": ["hermes"]},
    {"worker_type_slugs": ["opencode"]}
  ]
}
JSON

services="$(local_worker_runner_services "$catalog")"
[[ "$services" == "runner-codex-cli runner-gemini-cli runner-minimax-cli runner-openclaw runner-do-agent runner-loopal runner-aider runner-claude-code runner-cursor-cli runner-grok-build runner-hermes runner-opencode" ]] || {
    echo "unexpected local Worker runner services: $services" >&2
    exit 1
}

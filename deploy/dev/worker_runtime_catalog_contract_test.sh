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
    {"worker_type_slugs": ["e2e-echo"]},
    {"worker_type_slugs": ["loopal"]}
  ]
}
JSON

services="$(local_worker_runner_services "$catalog")"
[[ "$services" == "runner-codex-cli runner-gemini-cli runner-minimax-cli runner-openclaw runner-do-agent runner-e2e-echo" ]] || {
    echo "unexpected local Worker runner services: $services" >&2
    exit 1
}

cat >"$catalog" <<'JSON'
{"images":[{"worker_type_slugs":["codex-cli"]}]}
JSON

services="$(local_worker_bootstrap_services "$catalog")"
[[ "$services" == "runner-codex-cli runner-e2e-echo" ]] || {
    echo "bootstrap omitted required e2e runner: $services" >&2
    exit 1
}

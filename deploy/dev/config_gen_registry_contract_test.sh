#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="$ROOT/deploy/dev"
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

source "$SCRIPT_DIR/lib/log.sh"
source "$SCRIPT_DIR/lib/config_gen.sh"

grep -qx '    export MCP_REGISTRY_ENABLED="${MCP_REGISTRY_ENABLED:-false}"' \
  "$SCRIPT_DIR/lib/host_services_lite.sh"

get_worktree_name() {
  printf 'registry-contract'
}

calculate_port_offset() {
  printf '1'
}

ENV_FILE="$TEMP_DIR/.env"
generate_env
grep -qx 'MCP_REGISTRY_ENABLED=false' "$ENV_FILE"

sed -i.bak '/^MCP_REGISTRY_ENABLED=/d' "$ENV_FILE"
rm -f "$ENV_FILE.bak"
generate_env
grep -qx 'MCP_REGISTRY_ENABLED=false' "$ENV_FILE"

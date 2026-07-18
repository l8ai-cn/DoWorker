#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="$ROOT/deploy/dev"
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

source "$SCRIPT_DIR/lib/log.sh"
source "$SCRIPT_DIR/lib/host_services_lite.sh"

_runtime_dir() {
    printf '%s/runtime' "$TEMP_DIR"
}

go() {
    [[ "$1" == "build" ]]
}

_reap_port() {
    :
}

_wait_http() {
    :
}

captured_secret=""
_launch_air() {
    captured_secret="${INTERNAL_API_SECRET:-}"
}

write_env() {
    cat > "$ENV_FILE" <<EOF
POSTGRES_PORT=10002
MARKETPLACE_HTTP_PORT=10022
PRIMARY_DOMAIN=localhost:10000
BACKEND_HTTP_PORT=10015
$1
EOF
}

ENV_FILE="$TEMP_DIR/.env"
unset INTERNAL_API_SECRET
write_env ""
start_marketplace_host_lite >/dev/null
[[ "$captured_secret" == "dev-internal-secret" ]]

unset INTERNAL_API_SECRET
write_env "INTERNAL_API_SECRET=custom-internal-secret"
start_marketplace_host_lite >/dev/null
[[ "$captured_secret" == "custom-internal-secret" ]]

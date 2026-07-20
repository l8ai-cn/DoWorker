#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DEV_DIR="$ROOT/deploy/dev"
RUNTIME_ENV="$DEV_DIR/runtime/backend/pattern-locked-backend.env"
DEV_ENV="$DEV_DIR/.env"

load_env_file() {
    local file="$1"
    [[ -f "$file" ]] || return 1
    set -a
    # shellcheck disable=SC1090
    source "$file"
    set +a
}

if ! load_env_file "$RUNTIME_ENV"; then
    load_env_file "$DEV_ENV" || {
        echo "backend runtime env not found: $RUNTIME_ENV or $DEV_ENV" >&2
        exit 1
    }
    export DB_HOST="${DB_HOST:-localhost}"
    export DB_PORT="${DB_PORT:-${POSTGRES_PORT:-}}"
    export DB_USER="${DB_USER:-agentsmesh}"
    export DB_PASSWORD="${DB_PASSWORD:-${POSTGRES_PASSWORD:-agentsmesh_dev}}"
    export DB_NAME="${DB_NAME:-agentsmesh}"
    export DB_SSLMODE="${DB_SSLMODE:-disable}"
    export SERVER_ADDRESS="${SERVER_ADDRESS:-:${BACKEND_HTTP_PORT:-}}"
    export GRPC_ADDRESS="${GRPC_ADDRESS:-:${BACKEND_GRPC_PORT:-}}"
fi

if [[ -z "${DB_PORT:-}" || -z "${SERVER_ADDRESS:-}" || -z "${GRPC_ADDRESS:-}" ]]; then
    echo "backend runtime env is incomplete: DB_PORT/SERVER_ADDRESS/GRPC_ADDRESS required" >&2
    exit 1
fi

exec "$ROOT/deploy/dev/runtime/backend/air/main"

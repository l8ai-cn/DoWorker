#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/log.sh"
source "$SCRIPT_DIR/lib/bootstrap.sh"

POSTGRES_PORT=15432
POSTGRES_PASSWORD=test-password
PRIMARY_DOMAIN=localhost:10000
PUBLIC_WEB_URL=http://localhost:10007
MOBILE_PUBLIC_BASE_URL=http://localhost:10020
PREVIEW_PUBLIC_ORIGIN=http://preview.localhost:10000
USE_HTTPS=false

go() {
    [[ "$1" == "run" ]]
    [[ "$DB_PORT" == "$POSTGRES_PORT" ]]
    [[ "$PRIMARY_DOMAIN" == "localhost:10000" ]]
    [[ "$PUBLIC_WEB_URL" == "http://localhost:10007" ]]
    [[ "$MOBILE_PUBLIC_BASE_URL" == "http://localhost:10020" ]]
    [[ "$PREVIEW_PUBLIC_ORIGIN" == "http://preview.localhost:10000" ]]
    [[ "$USE_HTTPS" == "false" ]]
}

sync_worker_definition_projections

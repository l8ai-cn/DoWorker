#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="$ROOT/deploy/dev/.env"

if [[ ! -f "$ENV_FILE" ]]; then
    echo "missing deploy/dev/.env; start the dev stack first" >&2
    exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

if SESSION_COMPAT_API_URL="http://localhost:${BACKEND_HTTP_PORT}" \
    node "$ROOT/tests/session-compat-smoke/wait-e2e-echo-runners-online.mjs"; then
    exit 0
fi

echo "no selectable e2e-echo runner online after wait" >&2
docker compose -f "$ROOT/deploy/dev/docker-compose.yml" \
    -f "$ROOT/deploy/dev/docker-compose.runners.yml" \
    --project-name "${COMPOSE_PROJECT_NAME}" \
    logs --tail=80 runner-e2e-echo runner-e2e-echo-2 >&2 || true
exit 1

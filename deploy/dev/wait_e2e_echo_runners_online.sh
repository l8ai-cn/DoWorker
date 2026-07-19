#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="$ROOT/deploy/dev/.env"

if [[ ! -f "$ENV_FILE" ]]; then
    echo "missing deploy/dev/.env; start the dev stack first" >&2
    exit 1
fi

source "$ENV_FILE"

for _ in {1..60}; do
    online=$(docker exec "${COMPOSE_PROJECT_NAME}-postgres-1" \
        psql -U agentsmesh -d agentsmesh -Atc \
        "SELECT count(*) FROM runners
         WHERE node_id IN ('dev-runner','dev-runner-2')
           AND status='online'
           AND available_agents @> '[\"e2e-echo\"]'
           AND last_heartbeat > now()-interval '60 seconds'" 2>/dev/null || echo 0)
    if [[ "${online}" -ge 1 ]]; then
        echo "e2e-echo runner(s) online: ${online}"
        exit 0
    fi
    sleep 3
done

echo "no selectable e2e-echo runner online after wait" >&2
docker compose -f "$ROOT/deploy/dev/docker-compose.yml" \
    -f "$ROOT/deploy/dev/docker-compose.runners.yml" \
    --project-name "${COMPOSE_PROJECT_NAME}" \
    logs --tail=80 runner-e2e-echo runner-e2e-echo-2 >&2 || true
exit 1

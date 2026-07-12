#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$ROOT/deploy/dev/.env"

if [[ -f "$ENV_FILE" ]]; then
  source "$ENV_FILE"
fi

PORT="${MOBILE_LOVABLE_PORT:-10021}"
HOST="${MOBILE_DEV_HOST:-127.0.0.1}"
BACKEND_PORT="${BACKEND_HTTP_PORT:-10015}"

export DO_WORKER_API_URL="${DO_WORKER_API_URL:-http://127.0.0.1:${BACKEND_PORT}}"

exec pnpm --dir "$ROOT" --filter @do-worker/mobile exec vite --host "$HOST" --port "$PORT"

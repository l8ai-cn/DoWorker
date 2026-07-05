#!/usr/bin/env bash
# ensure_coordinator_runner.sh — auto-provision dev runner for coordinator dispatch.
#
# Invoked by backend when COORDINATOR_RUNNER_LAUNCHER points here and no online
# runner supports the requested agent. Args: <org_id> <agent_slug> (informational).
#
# Brings up the agent-specific docker runner service from deploy/dev.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
agent_slug="${2:-}"

if [[ -f "$SCRIPT_DIR/.env" ]]; then
  # shellcheck disable=SC1091
  source "$SCRIPT_DIR/.env"
fi

export COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml:docker-compose.runners.yml}"

case "$agent_slug" in
  claude-code) service=runner-claude-code ;;
  codex-cli) service=runner-codex-cli ;;
  gemini-cli) service=runner-gemini-cli ;;
  e2e-echo) service=runner-e2e-echo ;;
  loopal) service=runner-loopal ;;
  *)
    echo "No dev runner compose service for agent ${agent_slug}" >&2
    exit 1
    ;;
esac

if docker compose ps "$service" --status running 2>/dev/null | grep -q "$service"; then
  exit 0
fi

docker compose up -d "$service"

#!/usr/bin/env bash
# ensure_coordinator_runner.sh — auto-provision dev runner for coordinator dispatch.
#
# Invoked by backend when COORDINATOR_RUNNER_LAUNCHER points here and no online
# runner supports the requested agent. Args: <org_id> <agent_slug> (informational).
#
# Brings up the docker runner service from deploy/dev if it is not running.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [[ -f "$SCRIPT_DIR/.env" ]]; then
  # shellcheck disable=SC1091
  source "$SCRIPT_DIR/.env"
fi

if docker compose ps runner --status running 2>/dev/null | grep -q runner; then
  exit 0
fi

docker compose up -d runner

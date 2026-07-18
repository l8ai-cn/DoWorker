#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/coordinator_runners.sh"

requested="$(
    resolve_requested_runners_launcher "" --backend-only --coordinator-runners
)"
[[ "$requested" == "coordinator" ]]

effective="$(resolve_effective_runners_launcher "$requested" docker false)"
[[ "$effective" == "coordinator" ]]

requested="$(resolve_requested_runners_launcher docker --runners-k8s)"
effective="$(resolve_effective_runners_launcher "$requested" coordinator true)"
[[ "$effective" == "k8s" ]]

grep -q 'export COMPOSE_FILE="docker-compose.yml:docker-compose.runners.yml"' \
    "$SCRIPT_DIR/ensure_coordinator_runner.sh"

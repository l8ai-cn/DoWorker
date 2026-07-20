#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source "$ROOT/deploy/dev/lib/coordinator_runners.sh"

[[ "$(resolve_requested_runners_launcher "" --coordinator-runners)" == "coordinator" ]]
[[ "$(resolve_requested_runners_launcher docker --runners-k8s)" == "k8s" ]]
[[ "$(resolve_requested_runners_launcher coordinator --backend-only)" == "coordinator" ]]
[[ "$(resolve_requested_runners_launcher "" --coordinator-runners --runners-k8s)" == "k8s" ]]
[[ "$(resolve_effective_runners_launcher k8s coordinator true)" == "k8s" ]]
[[ "$(resolve_effective_runners_launcher "" coordinator false)" == "coordinator" ]]
[[ "$(resolve_effective_runners_launcher "" docker true)" == "coordinator" ]]

grep -q 'persist_runners_launcher_mode "$effective_runner_launcher"' "$ROOT/deploy/dev/dev.sh"
grep -q 'effective_runner_launcher="$(resolve_effective_runners_launcher' "$ROOT/deploy/dev/dev.sh"
grep -q 'dev_lite_enabled && echo true || echo false' "$ROOT/deploy/dev/dev.sh"

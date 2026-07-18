#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

grep -Fxq 'include_dir = ["backend"]' air/backend.toml
grep -Fq '"backend/logs",' air/backend.toml
grep -Fxq '    export LOG_FILE="$(_runtime_dir)/backend/agentsmesh.log"' \
  lib/host_services_lite.sh

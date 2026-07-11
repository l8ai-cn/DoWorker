#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

grep -Fq 'exec /usr/local/bin/do-worker-runner run --config "$CONFIG_FILE"' \
  runner-entrypoint.sh

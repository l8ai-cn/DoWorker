#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

if ! grep -Eq '^include_dir = \["relay"\]$' air/relay.toml; then
    echo "relay air watcher must only include relay source files" >&2
    exit 1
fi

if ! grep -Fq 'export BACKEND_URL="http://127.0.0.1:${BACKEND_HTTP_PORT}"' lib/host_services_lite.sh; then
    echo "host relay must use the direct backend listener" >&2
    exit 1
fi

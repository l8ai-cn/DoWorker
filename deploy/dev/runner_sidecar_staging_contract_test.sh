#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="$ROOT/deploy/dev"

source "$SCRIPT_DIR/lib/log.sh"
source "$SCRIPT_DIR/lib/host_services_lite.sh"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

source_binary="$tmp_dir/source"
destination="$tmp_dir/binaries/sidecar"
printf '#!/bin/sh\nexit 0\n' >"$source_binary"
chmod +x "$source_binary"

stage_runner_sidecar_binary "$source_binary" "$destination" "test-sidecar"
cmp "$source_binary" "$destination"
[[ -x "$destination" ]]

if stage_runner_sidecar_binary "$tmp_dir/missing" "$destination" "missing-sidecar"; then
    echo "expected missing sidecar staging to fail" >&2
    exit 1
fi

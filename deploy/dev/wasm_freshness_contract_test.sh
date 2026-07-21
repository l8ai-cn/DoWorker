#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

source "$ROOT/deploy/dev/lib/lifecycle_wasm.sh"

OUT="$TEMP_DIR/packages/agent-cloud-wasm/wasm_pkg.js"
mkdir -p \
    "$TEMP_DIR/proto/pod/v1" \
    "$TEMP_DIR/clients/core/crates/wasm/src" \
    "$TEMP_DIR/clients/core/crates/proto/pod/src" \
    "$TEMP_DIR/clients/core/target/debug" \
    "$TEMP_DIR/scripts" \
    "$(dirname "$OUT")"

touch "$OUT"
if _agent_cloud_wasm_needs_build "$TEMP_DIR"; then
    echo "fresh wasm unexpectedly requires a rebuild" >&2
    exit 1
fi

touch "$TEMP_DIR/proto/pod/v1/worker_creation.proto"
if ! _agent_cloud_wasm_needs_build "$TEMP_DIR"; then
    echo "newer proto must require a wasm rebuild" >&2
    exit 1
fi

touch "$OUT"
touch "$TEMP_DIR/clients/core/crates/proto/pod/src/lib.rs"
touch "$TEMP_DIR/clients/core/target/debug/ignored"
if _agent_cloud_wasm_needs_build "$TEMP_DIR"; then
    echo "generated Rust and target artifacts must not require a wasm rebuild" >&2
    exit 1
fi

touch "$TEMP_DIR/clients/core/crates/wasm/src/lib.rs"
if ! _agent_cloud_wasm_needs_build "$TEMP_DIR"; then
    echo "newer Rust source must require a wasm rebuild" >&2
    exit 1
fi

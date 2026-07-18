#!/usr/bin/env bash
# Generate Go protobuf/grpc stubs into proto/gen/go/ (committed mirror).
#
# Requires: protoc + protoc-gen-go + protoc-gen-go-grpc
#
# Usage:
#   ./scripts/proto-gen-go.sh          # regenerate if needed
#   ./scripts/proto-gen-go.sh --force  # always regenerate

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Homebrew protoc on Apple Silicon often isn't on non-interactive PATH.
export PATH="/opt/homebrew/bin:/usr/local/bin:${PATH}"

FORCE=false
[[ "${1:-}" == "--force" ]] && FORCE=true

PROTO_FILES=()
while IFS= read -r source; do
    PROTO_FILES+=("$source")
done < <(find "$ROOT/proto" -name '*.proto' ! -path '*/gen/*' | sort)
((${#PROTO_FILES[@]} > 0)) || {
    echo "error: no .proto files found under proto/" >&2
    exit 1
}

_generated_stubs_current() {
    local source relative generated grpc_generated newest_source
    newest_source="${PROTO_FILES[0]}"
    for source in "${PROTO_FILES[@]}"; do
        [[ "$source" -nt "$newest_source" ]] && newest_source="$source"
    done
    for source in "${PROTO_FILES[@]}"; do
        relative="${source#"$ROOT/proto/"}"
        generated="$ROOT/proto/gen/go/${relative%.proto}.pb.go"
        [[ -f "$generated" && ! "$newest_source" -nt "$generated" ]] || return 1
        if grep -Eq '^[[:space:]]*service[[:space:]]+[A-Za-z_]' "$source"; then
            grpc_generated="$ROOT/proto/gen/go/${relative%.proto}_grpc.pb.go"
            [[ -f "$grpc_generated" && ! "$newest_source" -nt "$grpc_generated" ]] || return 1
        fi
    done
}

if [[ "$FORCE" == "false" ]] && _generated_stubs_current; then
    echo "proto/gen/go stubs are current; use --force to regenerate"
    exit 0
fi

GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
export PATH="$GOBIN:$PATH"

LOCK_DIR="$ROOT/proto/gen/go/.generation.lock"
_release_lock() {
    rm -f "$LOCK_DIR/pid"
    rmdir "$LOCK_DIR" 2>/dev/null || true
}

_acquire_lock() {
    local attempt owner
    mkdir -p "$(dirname "$LOCK_DIR")"
    for attempt in {1..120}; do
        if mkdir "$LOCK_DIR" 2>/dev/null; then
            printf '%s\n' "$$" > "$LOCK_DIR/pid"
            trap _release_lock EXIT
            return
        fi
        owner="$(cat "$LOCK_DIR/pid" 2>/dev/null || true)"
        if [[ "$owner" =~ ^[0-9]+$ ]] && ! kill -0 "$owner" 2>/dev/null; then
            rm -f "$LOCK_DIR/pid"
            rmdir "$LOCK_DIR" 2>/dev/null || true
        fi
        sleep 0.25
    done
    echo "error: timed out waiting for Go proto generation lock" >&2
    exit 1
}

_install_go_plugins() {
    if ! command -v protoc-gen-go >/dev/null 2>&1; then
        echo "Installing protoc-gen-go..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
    fi
    if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
        echo "Installing protoc-gen-go-grpc..."
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
    fi
}

_gen_with_protoc() {
    _install_go_plugins
    mkdir -p "$ROOT/proto/gen/go"
    protoc \
        --proto_path="$ROOT/proto" \
        --go_out="$ROOT/proto/gen/go" --go_opt=paths=source_relative \
        --go-grpc_out="$ROOT/proto/gen/go" --go-grpc_opt=paths=source_relative \
        "${PROTO_FILES[@]}"
    echo "Generated ${#PROTO_FILES[@]} proto inputs → proto/gen/go/ (protoc)"
}

if ! command -v protoc >/dev/null 2>&1; then
    echo "error: protoc not found" >&2
    echo "  Install: brew install protobuf" >&2
    exit 1
fi

_acquire_lock
if [[ "$FORCE" == "false" ]] && _generated_stubs_current; then
    echo "proto/gen/go stubs became current while waiting for generation"
    exit 0
fi
_gen_with_protoc

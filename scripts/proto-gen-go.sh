#!/usr/bin/env bash
# Generate Go protobuf/grpc stubs into proto/gen/go/ (committed mirror).
#
# Primary path: protoc + protoc-gen-go + protoc-gen-go-grpc (no Bazel daemon).
# Fallback: one-shot `bazel build` of all go_proto_library targets, then sync
# from bazel-bin — used when protoc is not installed yet.
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

MARKER="$ROOT/proto/gen/go/loop/v1/loop.pb.go"
if [[ "$FORCE" == "false" && -f "$MARKER" ]]; then
    echo "proto/gen/go already present ($MARKER); use --force to regenerate"
    exit 0
fi

GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
export PATH="$GOBIN:$PATH"

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

_sync_from_bazel_bin() {
    local copied=0
    while IFS= read -r -d '' src; do
        local rel="${src#*github.com/anthropics/agentsmesh/proto/gen/go/}"
        local dest="$ROOT/proto/gen/go/$rel"
        mkdir -p "$(dirname "$dest")"
        cp "$src" "$dest"
        copied=$((copied + 1))
    done < <(find "$ROOT/bazel-bin/proto" -name '*.pb.go' -print0 2>/dev/null || true)
    if (( copied == 0 )); then
        echo "error: bazel build produced no .pb.go files under bazel-bin/proto" >&2
        return 1
    fi
    echo "Synced $copied Go proto files from bazel-bin → proto/gen/go/"
}

_gen_with_protoc() {
    _install_go_plugins
    mkdir -p "$ROOT/proto/gen/go"
    mapfile -t proto_files < <(find "$ROOT/proto" -name '*.proto' ! -path '*/gen/*' | sort)
    if ((${#proto_files[@]} == 0)); then
        echo "error: no .proto files found under proto/" >&2
        return 1
    fi
    protoc \
        --proto_path="$ROOT/proto" \
        --go_out="$ROOT/proto/gen/go" --go_opt=paths=source_relative \
        --go-grpc_out="$ROOT/proto/gen/go" --go-grpc_opt=paths=source_relative \
        "${proto_files[@]}"
    echo "Generated ${#proto_files[@]} proto inputs → proto/gen/go/ (protoc)"
}

_gen_with_bazel() {
    if ! command -v bazel >/dev/null 2>&1; then
        echo "error: protoc not found and bazel unavailable" >&2
        echo "  Install protoc:  brew install protobuf" >&2
        echo "  Or install bazel for one-time bootstrap sync" >&2
        return 1
    fi
    echo "protoc not found — bootstrapping via one-shot bazel build //proto/..."
    bazel build //proto/...
    _sync_from_bazel_bin
}

if command -v protoc >/dev/null 2>&1; then
    _gen_with_protoc
else
    _gen_with_bazel
fi

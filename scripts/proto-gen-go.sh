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

_gen_with_protoc() {
    _install_go_plugins
    mkdir -p "$ROOT/proto/gen/go"
    local proto_files=()
    while IFS= read -r f; do
        proto_files+=("$f")
    done < <(find "$ROOT/proto" -name '*.proto' ! -path '*/gen/*' | sort)
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

if ! command -v protoc >/dev/null 2>&1; then
    echo "error: protoc not found" >&2
    echo "  Install: brew install protobuf" >&2
    exit 1
fi

_gen_with_protoc

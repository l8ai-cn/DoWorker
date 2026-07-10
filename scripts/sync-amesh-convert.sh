#!/usr/bin/env bash
# Generate *_convert.amesh.go files for plain `go build` / air dev.
#
# Output: backend/internal/api/connect/**/<name>_convert.amesh.go (gitignored).
# Requires proto/gen/go (see scripts/proto-gen-go.sh).
#
# Usage:
#   ./scripts/sync-amesh-convert.sh
#   ./scripts/sync-amesh-convert.sh --force

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export PATH="/opt/homebrew/bin:/usr/local/bin:${PATH}"

FORCE=false
[[ "${1:-}" == "--force" ]] && FORCE=true

MARKER="$ROOT/backend/internal/api/connect/pod/pod_convert.amesh.go"
if [[ "$FORCE" == "false" && -f "$MARKER" ]]; then
    echo "amesh convert files present ($MARKER); use --force to regenerate"
    exit 0
fi

if [[ -f "$ROOT/proto/gen/go/loop/v1/loop.pb.go" ]]; then
    echo "proto/gen/go present — skipping proto regen"
else
    bash "$ROOT/scripts/proto-gen-go.sh"
fi

GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
export PATH="$GOBIN:$PATH"

PLUGIN="/tmp/protoc-gen-amesh-convert"
echo "Building protoc-gen-amesh-convert..."
go build -o "$PLUGIN" ./tools/protoc-gen-amesh-convert

if command -v buf >/dev/null 2>&1; then
    echo "Generating amesh convert via buf..."
    buf generate --config buf.amesh.yaml --template buf.amesh.gen.yaml
    count=$(find backend/internal/api/connect -name '*_convert.amesh.go' 2>/dev/null | wc -l | tr -d ' ')
    echo "Generated $count amesh convert files (buf)"
    exit 0
fi

echo "error: buf not found — install: brew install bufbuild/buf/buf" >&2
exit 1

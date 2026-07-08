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

if ! command -v bazel >/dev/null 2>&1; then
    echo "error: need buf (brew install bufbuild/buf/buf) or bazel for amesh codegen" >&2
    exit 1
fi

echo "buf not found — syncing amesh convert from bazel build..."
mapfile -t targets < <(bazel query 'filter(_convert_amesh, //backend/...)' 2>/dev/null)
bazel build "${targets[@]}"
copied=0
while IFS= read -r -d '' src; do
    rel="${src#"$ROOT/bazel-bin/backend/"}"
    dest="$ROOT/backend/$rel"
    mkdir -p "$(dirname "$dest")"
    rm -f "$dest"
    cat "$src" > "$dest"
    copied=$((copied + 1))
done < <(find "$ROOT/bazel-bin/backend" -name '*_convert.amesh.go' -print0 2>/dev/null || true)
echo "Synced $copied amesh convert files from bazel-bin"
if (( copied == 0 )); then
    echo "error: no amesh files produced" >&2
    exit 1
fi

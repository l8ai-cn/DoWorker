#!/usr/bin/env bash
# Ensure proto/gen/go + amesh convert exist for plain go build/test.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
export PATH="$(go env GOPATH)/bin:/usr/local/bin:/opt/homebrew/bin:/usr/local/bin:${PATH}"

if ! command -v protoc >/dev/null 2>&1; then
  echo "error: protoc required (run scripts/ci-install-codegen-tools.sh)" >&2
  exit 1
fi
if ! command -v buf >/dev/null 2>&1; then
  echo "error: buf required for amesh convert (run scripts/ci-install-codegen-tools.sh)" >&2
  exit 1
fi
if ! command -v protoc-gen-go >/dev/null 2>&1; then
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
fi
if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
fi

bash "$ROOT/scripts/proto-gen-go.sh" --force
bash "$ROOT/scripts/sync-amesh-convert.sh" --force

marker="$ROOT/backend/internal/api/connect/pod/pod_convert.amesh.go"
if [[ ! -f "$marker" ]]; then
  echo "error: amesh convert missing after codegen ($marker)" >&2
  exit 1
fi
echo "codegen ok: proto/gen/go + amesh convert"

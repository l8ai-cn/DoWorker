#!/usr/bin/env bash
# Install protoc + buf for CI / local codegen. Idempotent.
set -euo pipefail

export PATH="$(go env GOPATH)/bin:/usr/local/bin:/opt/homebrew/bin:${PATH}"

if ! command -v protoc >/dev/null 2>&1; then
  if command -v apt-get >/dev/null 2>&1; then
    sudo apt-get update
    sudo apt-get install -y protobuf-compiler
  elif command -v choco >/dev/null 2>&1; then
    choco install protoc -y
  elif command -v brew >/dev/null 2>&1; then
    brew install protobuf
  else
    echo "error: install protoc (protobuf-compiler)" >&2
    exit 1
  fi
fi

if ! command -v buf >/dev/null 2>&1; then
  # go install is portable across linux/mac/windows CI runners.
  go install github.com/bufbuild/buf/cmd/buf@v1.50.0
fi

echo "protoc=$(protoc --version)"
echo "buf=$(buf --version)"
echo "PATH=$PATH"

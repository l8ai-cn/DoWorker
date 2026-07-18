#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

bash scripts/proto-gen-all.sh

generated_paths=(
  proto/gen/ts
  proto/gen/go
  backend/internal/api/connect
  clients/core/crates/proto
)

if ! git diff --quiet -- "${generated_paths[@]}"; then
  echo "proto codegen is out of date; run pnpm run proto:gen-all and commit generated files" >&2
  git diff --name-only -- "${generated_paths[@]}" >&2
  exit 1
fi

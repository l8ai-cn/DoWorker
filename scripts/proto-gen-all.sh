#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

export PATH="$(go env GOPATH)/bin:/opt/homebrew/bin:/usr/local/bin:${PATH}"

pnpm exec buf generate
bash scripts/proto-gen-go.sh --force
bash scripts/sync-amesh-convert.sh --force
bash scripts/seed-rust-proto-stubs.sh
(cd clients/core && cargo run -p agent_cloud_proto_gen --bin gen-proto)

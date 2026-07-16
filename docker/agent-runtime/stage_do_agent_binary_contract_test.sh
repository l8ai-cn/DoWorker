#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRIPT="${ROOT}/docker/agent-runtime/stage_do_agent_binary.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

cat >"${TMP}/main.go" <<'EOF'
package main

func main() {}
EOF
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "${TMP}/do-agent" "${TMP}/main.go"

if REQUIRE_DO_AGENT_BINARY=1 "$SCRIPT" "${TMP}/missing" >/dev/null 2>&1; then
  echo "release staging accepted an unspecified do-agent artifact" >&2
  exit 1
fi

if REQUIRE_DO_AGENT_BINARY=1 \
  DO_AGENT_BINARY="${TMP}/do-agent" \
  DO_AGENT_BINARY_SHA256="sha256:$(printf '0%.0s' {1..64})" \
  "$SCRIPT" "${TMP}/bad-hash" >/dev/null 2>&1; then
  echo "release staging accepted a mismatched do-agent checksum" >&2
  exit 1
fi

if REQUIRE_DO_AGENT_BINARY=1 \
  DO_AGENT_BINARY="${TMP}/do-agent" \
  DO_AGENT_BINARY_SHA256="sha256:$(shasum -a 256 "${TMP}/do-agent" | awk '{print $1}')" \
  "$SCRIPT" "${TMP}/self-approved" >/dev/null 2>&1; then
  echo "release staging accepted an artifact approved only by its caller" >&2
  exit 1
fi

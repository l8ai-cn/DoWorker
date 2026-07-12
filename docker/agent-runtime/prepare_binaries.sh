#!/usr/bin/env bash
# Stage linux/amd64 runner + agent sidecar binaries for docker build context.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
STAGING="${1:?staging directory required}"
RUNTIME="${2:-all}"
DEPLOY_DEV="${REPO_ROOT}/deploy/dev"

rm -rf "$STAGING"
mkdir -p "$STAGING"

if [[ ! -f "${REPO_ROOT}/proto/gen/go/runner/v1/runner.pb.go" ]]; then
  bash "${REPO_ROOT}/scripts/proto-gen-go.sh" --force
fi

go_cross() {
  local out="$1" pkg="$2"
  (
    cd "$REPO_ROOT"
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "${STAGING}/${out}" "${pkg}"
  )
  chmod +x "${STAGING}/${out}"
}

echo "▶ go build runner (linux/amd64)..."
go_cross "runner-binary" ./runner/cmd/runner

echo "▶ go build e2e-mock-agent (linux/amd64)..."
go_cross "e2e-mock-agent-binary" ./runner/internal/agents/mockagent/cmd/e2e-mock-agent

if [[ -f "${DEPLOY_DEV}/loopal-binary" ]]; then
  cp "${DEPLOY_DEV}/loopal-binary" "${STAGING}/loopal-binary"
else
  cp "${STAGING}/e2e-mock-agent-binary" "${STAGING}/loopal-binary"
fi
chmod +x "${STAGING}/loopal-binary"

if [[ "${RUNTIME}" == "all" || "${RUNTIME}" == "do-agent" ]]; then
  if [[ ! -f "${DEPLOY_DEV}/do-agent-binary" ]]; then
  # shellcheck source=../../deploy/dev/lib/build_do_agent_binary.sh
    source "${DEPLOY_DEV}/lib/build_do_agent_binary.sh"
    if ! build_do_agent_binary; then
      echo "do-agent binary build failed" >&2
      exit 1
    fi
  fi
  if [[ ! -x "${DEPLOY_DEV}/do-agent-binary" ]]; then
    echo "do-agent binary is required for AGENT_RUNTIME=${RUNTIME}" >&2
    exit 1
  fi
  cp "${DEPLOY_DEV}/do-agent-binary" "${STAGING}/do-agent-binary"
  chmod +x "${STAGING}/do-agent-binary"
fi

cp "${DEPLOY_DEV}/runner-entrypoint.sh" "${STAGING}/runner-entrypoint.sh"
chmod +x "${STAGING}/runner-entrypoint.sh"

echo "✓ build context ready: ${STAGING}"

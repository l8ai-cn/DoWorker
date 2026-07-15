#!/usr/bin/env bash
# Stage Linux runner + agent sidecar binaries for the requested image platform.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
STAGING="${1:?staging directory required}"
DEPLOY_DEV="${REPO_ROOT}/deploy/dev"
TARGET_ARCH="${TARGET_ARCH:-amd64}"

case "$TARGET_ARCH" in
  amd64|arm64) ;;
  *)
    echo "unsupported TARGET_ARCH=${TARGET_ARCH}" >&2
    exit 1
    ;;
esac

rm -rf "$STAGING"
mkdir -p "$STAGING"

if [[ ! -f "${REPO_ROOT}/proto/gen/go/runner/v1/runner.pb.go" ]]; then
  bash "${REPO_ROOT}/scripts/proto-gen-go.sh" --force
fi

go_cross() {
  local out="$1" pkg="$2"
  (
    cd "$REPO_ROOT"
    GOOS=linux GOARCH="$TARGET_ARCH" CGO_ENABLED=0 go build -o "${STAGING}/${out}" "${pkg}"
  )
  chmod +x "${STAGING}/${out}"
}

echo "▶ go build runner (linux/${TARGET_ARCH})..."
go_cross "runner-binary" ./runner/cmd/runner

echo "▶ go build e2e-mock-agent (linux/${TARGET_ARCH})..."
go_cross "e2e-mock-agent-binary" ./runner/internal/agents/mockagent/cmd/e2e-mock-agent

if [[ -f "${DEPLOY_DEV}/loopal-binary" ]]; then
  cp "${DEPLOY_DEV}/loopal-binary" "${STAGING}/loopal-binary"
else
  cp "${STAGING}/e2e-mock-agent-binary" "${STAGING}/loopal-binary"
fi
chmod +x "${STAGING}/loopal-binary"

if [[ -x "${DEPLOY_DEV}/do-agent-binary" ]]; then
  cp "${DEPLOY_DEV}/do-agent-binary" "${STAGING}/do-agent-binary"
elif [[ -n "${DOAGENT_DIR:-}" || -d "${HOME}/Documents/code/doagent" || -d "${HOME}/Documents/code/AgentForge/doagent" ]]; then
  # shellcheck source=../../deploy/dev/lib/build_do_agent_binary.sh
  source "${DEPLOY_DEV}/lib/build_do_agent_binary.sh"
  build_do_agent_binary
  cp "${DEPLOY_DEV}/do-agent-binary" "${STAGING}/do-agent-binary"
else
  echo "do-agent source is unavailable; only the do-agent runtime build is disabled" >&2
fi
if [[ -f "${STAGING}/do-agent-binary" ]]; then
  chmod +x "${STAGING}/do-agent-binary"
fi

cp "${DEPLOY_DEV}/runner-entrypoint.sh" "${STAGING}/runner-entrypoint.sh"
chmod +x "${STAGING}/runner-entrypoint.sh"

echo "✓ build context ready: ${STAGING}"

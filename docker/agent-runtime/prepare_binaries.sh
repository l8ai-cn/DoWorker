#!/usr/bin/env bash
# Stage one runtime's linux/amd64 runner and required sidecar binaries.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
STAGING="${1:?staging directory required}"
AGENT_RUNTIME="${2:?agent runtime required}"
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
mkdir -p "$STAGING/binaries"
touch "$STAGING/binaries/.keep"

bash "${REPO_ROOT}/scripts/proto-gen-go.sh"

go_cross() {
  local out="$1" pkg="$2"
  (
    cd "$REPO_ROOT"
    GOOS=linux GOARCH="$TARGET_ARCH" CGO_ENABLED=0 go build \
      -trimpath \
      -buildvcs=false \
      -ldflags="-s -w -buildid=" \
      -o "${STAGING}/${out}" \
      "${pkg}"
  )
  chmod +x "${STAGING}/${out}"
}

echo "▶ go build runner (linux/${TARGET_ARCH})..."
go_cross "runner-binary" ./runner/cmd/runner

stage_sidecar() {
  local name="$1" source="$2"
  [[ -x "$source" ]] || {
    echo "${AGENT_RUNTIME} requires executable ${source}" >&2
    exit 1
  }
  cp "$source" "${STAGING}/binaries/${name}"
  chmod +x "${STAGING}/binaries/${name}"
}

stage_loopal() {
  local source="${LOOPAL_BINARY:-}"
  [[ -n "$source" ]] || {
    echo "loopal requires LOOPAL_BINARY to point to a real Loopal CLI artifact" >&2
    exit 1
  }
  if grep -aFq "runner/internal/agents/mockagent" "$source"; then
    echo "loopal artifact is an E2E mock binary, not a real Loopal CLI: ${source}" >&2
    exit 1
  fi
  stage_sidecar "loopal-binary" "$source"
}

case "$AGENT_RUNTIME" in
  e2e-echo)
    echo "▶ go build e2e-mock-agent (linux/amd64)..."
    go_cross "binaries/e2e-mock-agent-binary" ./runner/internal/agents/mockagent/cmd/e2e-mock-agent
    ;;
  loopal)
    stage_loopal
    ;;
  do-agent)
    if [[ ! -x "${DEPLOY_DEV}/do-agent-binary" ]]; then
      source "${DEPLOY_DEV}/lib/build_do_agent_binary.sh"
      build_do_agent_binary
    fi
    stage_sidecar "do-agent-binary" "${DEPLOY_DEV}/do-agent-binary"
    ;;
esac

cp "${DEPLOY_DEV}/runner-entrypoint.sh" "${STAGING}/runner-entrypoint.sh"
chmod +x "${STAGING}/runner-entrypoint.sh"
cp "${DEPLOY_DEV}/runner-ssh-bootstrap.sh" "${STAGING}/runner-ssh-bootstrap.sh"
chmod +x "${STAGING}/runner-ssh-bootstrap.sh"
cp "${REPO_ROOT}/docker/agent-runtime/minimax-cli-wrapper.sh" \
  "${STAGING}/minimax-cli-wrapper.sh"
chmod +x "${STAGING}/minimax-cli-wrapper.sh"

echo "✓ build context ready for ${AGENT_RUNTIME}: ${STAGING}"

#!/usr/bin/env bash
# Stage linux/amd64 runner + agent sidecar binaries for docker build context.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
STAGING="${1:?staging directory required}"
DEPLOY_DEV="${REPO_ROOT}/deploy/dev"
LINUX_PLATFORM="@rules_go//go/toolchain:linux_amd64"

rm -rf "$STAGING"
mkdir -p "$STAGING"

bazel_build() {
  local target="$1" bazel_out="$2" staging_name="$3"
  (
    cd "$REPO_ROOT"
    # `pure` keeps cgo off so the resolution-only stub CC toolchain
    # (//tools/crosscc, registered for linux/amd64) is never invoked.
    bazel build "$target" --platforms="$LINUX_PLATFORM" --@rules_go//go/config:pure
  )
  cp -L "${REPO_ROOT}/bazel-bin/${bazel_out}" "${STAGING}/${staging_name}"
  chmod +x "${STAGING}/${staging_name}"
}

echo "▶ bazel build runner (linux/amd64)..."
bazel_build //runner/cmd/runner:runner \
  "runner/cmd/runner/runner_/runner" \
  "runner-binary"

echo "▶ bazel build e2e-mock-agent (linux/amd64)..."
bazel_build //runner/internal/agents/mockagent/cmd/e2e-mock-agent:e2e-mock-agent \
  "runner/internal/agents/mockagent/cmd/e2e-mock-agent/e2e-mock-agent_/e2e-mock-agent" \
  "e2e-mock-agent-binary"

if [[ -f "${DEPLOY_DEV}/loopal-binary" ]]; then
  cp "${DEPLOY_DEV}/loopal-binary" "${STAGING}/loopal-binary"
else
  cp "${STAGING}/e2e-mock-agent-binary" "${STAGING}/loopal-binary"
fi
chmod +x "${STAGING}/loopal-binary"

if [[ -f "${DEPLOY_DEV}/do-agent-binary" ]]; then
  cp "${DEPLOY_DEV}/do-agent-binary" "${STAGING}/do-agent-binary"
else
  # shellcheck source=../../deploy/dev/lib/build_do_agent_binary.sh
  source "${DEPLOY_DEV}/lib/build_do_agent_binary.sh"
  # build may return 0 after writing a CI stub; still fall back if no file.
  if build_do_agent_binary && [[ -f "${DEPLOY_DEV}/do-agent-binary" ]]; then
    cp "${DEPLOY_DEV}/do-agent-binary" "${STAGING}/do-agent-binary"
  else
    echo "⚠ do-agent 不可用，使用 e2e-mock-agent 占位" >&2
    cp "${STAGING}/e2e-mock-agent-binary" "${STAGING}/do-agent-binary"
  fi
fi
chmod +x "${STAGING}/do-agent-binary"

cp "${DEPLOY_DEV}/runner-entrypoint.sh" "${STAGING}/runner-entrypoint.sh"
chmod +x "${STAGING}/runner-entrypoint.sh"

echo "✓ build context ready: ${STAGING}"

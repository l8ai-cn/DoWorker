#!/usr/bin/env bash
# Build one or all Do Worker agent-runtime Docker images (runner + pre-installed CLI).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

AGENT_RUNTIMES=(
  claude-code
  codex-cli
  gemini-cli
  minimax-cli
  aider
  opencode
  loopal
  do-agent
  grok-build
  openclaw
  hermes
)

RUNTIME="${1:-all}"
IMAGE_PREFIX="${IMAGE_PREFIX:-do-worker/runner}"
BASE_IMAGE="${BASE_IMAGE:-do-worker/runner-base:latest}"
PLATFORM="${PLATFORM:-linux/amd64}"
STAGING="${STAGING_DIR:-${SCRIPT_DIR}/_context}"
BUILD_RETRIES="${BUILD_RETRIES:-3}"

usage() {
  cat <<EOF
用法: build.sh [all|<agent-runtime>]

Agent runtimes:
  claude-code   Claude Code CLI (@anthropic-ai/claude-code)
  codex-cli     OpenAI Codex CLI (@openai/codex)
  gemini-cli    Google Gemini CLI (@google/gemini-cli)
  minimax-cli   MiniMax CLI (mmx-cli)
  aider         Aider (pip)
  opencode      OpenCode CLI
  loopal        Loopal CLI
  do-agent      Do Agent binary
  grok-build    Grok Build CLI (@xai-official/grok)
  openclaw      OpenClaw CLI (openclaw)
  hermes        Hermes Agent CLI (hermes-agent)

  (e2e-echo 仅 dev compose / CI 使用，不走 build_all)

环境变量:
  IMAGE_PREFIX    镜像名前缀 (默认 do-worker/runner)
  BASE_IMAGE      共享 base 镜像 (默认 do-worker/runner-base:latest)
  PLATFORM        docker build --platform (默认 linux/amd64)
  STAGING_DIR     二进制 staging 目录
  FORCE_REBUILD   设为 1 强制重建已有镜像
  BUILD_RETRIES   docker build 失败重试次数 (默认 3)
  HERMES_AGENT_VERSION
                  hermes-agent npm/PyPI bridge version (默认 0.18.2)
EOF
}

contains_runtime() {
  local want="$1" rt
  for rt in "${AGENT_RUNTIMES[@]}"; do
    [[ "$rt" == "$want" ]] && return 0
  done
  return 1
}

docker_build_with_retry() {
  local attempt=1
  while [[ "$attempt" -le "$BUILD_RETRIES" ]]; do
    if "$@"; then
      return 0
    fi
    if [[ "$attempt" -eq "$BUILD_RETRIES" ]]; then
      return 1
    fi
    echo "⚠ docker build 失败 (attempt ${attempt}/${BUILD_RETRIES})，30s 后重试..." >&2
    sleep 30
    attempt=$((attempt + 1))
  done
}

build_base() {
  echo ""
  echo "=========================================="
  echo "  Building shared base: ${BASE_IMAGE}"
  echo "=========================================="
  docker_build_with_retry docker build --platform "$PLATFORM" \
    --target base \
    -f "${SCRIPT_DIR}/Dockerfile" \
    --build-arg "HTTP_PROXY=" \
    --build-arg "HTTPS_PROXY=" \
    --build-arg "http_proxy=" \
    --build-arg "https_proxy=" \
    -t "$BASE_IMAGE" \
    "$SCRIPT_DIR"
  echo "✓ ${BASE_IMAGE}"
}

build_one() {
  local rt="$1"
  local tag="${IMAGE_PREFIX}-${rt}:latest"
  local target="runtime"
  if [[ "$rt" == "do-agent" ]]; then
    target="do-agent-runtime"
  fi
  if [[ "${FORCE_REBUILD:-0}" != "1" ]] && docker image inspect "$tag" >/dev/null 2>&1; then
    echo "⏭ ${tag} 已存在 (FORCE_REBUILD=1 可强制重建)"
    return 0
  fi
  echo ""
  echo "=========================================="
  echo "  Building ${tag}"
  echo "  AGENT_RUNTIME=${rt}  PLATFORM=${PLATFORM}"
  echo "=========================================="
  docker_build_with_retry docker build --platform "$PLATFORM" \
    --target "$target" \
    -f "${SCRIPT_DIR}/Dockerfile" \
    --build-arg "AGENT_RUNTIME=${rt}" \
    --build-arg "HTTP_PROXY=" \
    --build-arg "HTTPS_PROXY=" \
    --build-arg "http_proxy=" \
    --build-arg "https_proxy=" \
    --build-arg "HERMES_AGENT_VERSION=${HERMES_AGENT_VERSION:-0.18.2}" \
    --cache-from "$BASE_IMAGE" \
    -t "$tag" \
    "$STAGING"
  echo "✓ ${tag}"
}

if [[ "$RUNTIME" == "-h" || "$RUNTIME" == "--help" ]]; then
  usage
  exit 0
fi

if [[ "$RUNTIME" != "all" ]] && ! contains_runtime "$RUNTIME"; then
  echo "未知 AGENT_RUNTIME: ${RUNTIME}" >&2
  usage >&2
  exit 1
fi

"${SCRIPT_DIR}/prepare_binaries.sh" "$STAGING" "$RUNTIME"
build_base

if [[ "$RUNTIME" == "all" ]]; then
  failed=()
  for rt in "${AGENT_RUNTIMES[@]}"; do
    if ! build_one "$rt"; then
      failed+=("$rt")
    fi
  done
  echo ""
  if [[ ${#failed[@]} -eq 0 ]]; then
    echo "✓ 全部 ${#AGENT_RUNTIMES[@]} 个 Do Worker agent-runtime 镜像已就绪 (${IMAGE_PREFIX}-*)"
  else
    echo "✗ 以下 runtime 构建失败: ${failed[*]}" >&2
    exit 1
  fi
else
  build_one "$RUNTIME"
fi

#!/usr/bin/env bash
# Build one or all Do Worker agent-runtime Docker images (runner + pre-installed CLI).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

AGENT_RUNTIMES=(
  claude-code
  codex-cli
  video-studio
  cursor-cli
  gemini-cli
  aider
  opencode
  loopal
  do-agent
  grok-build
  minimax-cli
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
  video-studio  Codex + FFmpeg/libass + Chromium + Remotion + Python + CJK fonts
  cursor-cli    Cursor CLI (agent)
  gemini-cli    Google Gemini CLI (@google/gemini-cli)
  aider         Aider (pip)
  opencode      OpenCode CLI
  loopal        Loopal CLI
  do-agent      Do Agent binary
  grok-build    Grok Build CLI (@xai-official/grok)
  minimax-cli   MiniMax CLI (mmx-cli)
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
  LOOPAL_BINARY
                  Loopal 的真实 CLI 二进制；构建 loopal 时必填且不得是 E2E mock
  RUNTIME_EXTENSION_BASE
                  用既有 Runner 镜像构建 OpenClaw/Hermes（本地离线初始化）
  HERMES_AGENT_VERSION
                  Hermes Agent version (默认 0.18.2)
  REMOTION_VERSION
                  Remotion runtime version (默认 4.0.489)
  MINIMAX_CLI_VERSION
                  MiniMax CLI version (默认 1.0.16)
  OPENCLAW_VERSION
                  OpenClaw CLI version (默认 2026.6.11)
  OPENCLAW_NODE_VERSION
                  OpenClaw extension Node version (默认 24.18.0)
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
  TARGET_ARCH="${PLATFORM#linux/}" "${SCRIPT_DIR}/prepare_binaries.sh" "$STAGING" "$rt"
  local -a build_cmd=(
    docker build --platform "$PLATFORM"
    -f "${SCRIPT_DIR}/Dockerfile"
  )
  if [[ -n "${RUNTIME_EXTENSION_BASE:-}" ]]; then
    build_cmd+=(--target runtime-extension --build-arg "RUNTIME_EXTENSION_BASE=${RUNTIME_EXTENSION_BASE}")
  else
    build_cmd+=(--target runtime)
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
  build_cmd+=(
    --build-arg "AGENT_RUNTIME=${rt}"
    --build-arg "HTTP_PROXY="
    --build-arg "HTTPS_PROXY="
    --build-arg "http_proxy="
    --build-arg "https_proxy="
    --build-arg "HERMES_AGENT_VERSION=${HERMES_AGENT_VERSION:-0.18.2}"
    --build-arg "MINIMAX_CLI_VERSION=${MINIMAX_CLI_VERSION:-1.0.16}"
    --build-arg "OPENCLAW_VERSION=${OPENCLAW_VERSION:-2026.6.11}"
    --build-arg "OPENCLAW_NODE_VERSION=${OPENCLAW_NODE_VERSION:-24.18.0}"
    --build-arg "REMOTION_VERSION=${REMOTION_VERSION:-4.0.489}"
    -t "$tag"
    "$STAGING"
  )
  docker_build_with_retry "${build_cmd[@]}"
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

if [[ -n "${RUNTIME_EXTENSION_BASE:-}" ]] \
  && [[ "$RUNTIME" != "openclaw" && "$RUNTIME" != "hermes" ]]; then
  echo "RUNTIME_EXTENSION_BASE is only supported for openclaw and hermes" >&2
  exit 1
fi

if [[ -z "${RUNTIME_EXTENSION_BASE:-}" ]]; then
  build_base
fi

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

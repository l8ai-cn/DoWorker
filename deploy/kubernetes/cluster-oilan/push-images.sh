#!/usr/bin/env bash
# Build + push every image Do Worker needs into the cluster Harbor so pods
# on doops-114-k8s pull from repo.aiedulab.cn:8443/agentsmesh/* (fast, node-local).
#
#   ./push-images.sh all        # platform + infra + runners
#   ./push-images.sh platform   # backend/marketplace/marketplace-web/relay/web/web-admin/mobile
#   ./push-images.sh mobile-access # backend/relay/mobile only
#   ./push-images.sh marketplace-core # backend/marketplace/marketplace-web/web
#   ./push-images.sh video-expert # backend/marketplace/marketplace-web/web/web-admin
#   ./push-images.sh web        # rebuild Web and retain other current digests
#   ./push-images.sh marketplace-web # rebuild public marketplace and retain other current digests
#   ./push-images.sh infra      # postgres/redis/minio/mc/kubectl mirrors
#   ./push-images.sh runners    # agent-runtime images (claude/codex/video/gemini/grok/openclaw/hermes/e2e-echo)
#   ./push-images.sh do-agent   # trusted do-agent artifact and immutable image only
#   ./push-images.sh video-runtime # build and push only runner-video-studio
#
# The build host must already be `docker login repo.aiedulab.cn:8443`.
set -euo pipefail

REG="repo.aiedulab.cn:8443"
PROJ="${REG}/agentsmesh"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
TARGET="${1:-all}"
PLATFORM="${PLATFORM:-linux/amd64}"
PLATFORM_DIGEST_BACKEND=""
PLATFORM_DIGEST_MARKETPLACE=""
PLATFORM_DIGEST_MARKETPLACE_WEB=""
PLATFORM_DIGEST_RELAY=""
PLATFORM_DIGEST_WEB=""
PLATFORM_DIGEST_WEB_ADMIN=""
PLATFORM_DIGEST_MOBILE=""
PLATFORM_DIGEST_PGVECTOR=""
PLATFORM_DIGEST_REDIS=""
PLATFORM_DIGEST_MINIO=""
PLATFORM_DIGEST_MC=""
PLATFORM_DIGEST_KUBECTL=""
source "$(dirname "${BASH_SOURCE[0]}")/harbor-image-publishing.sh"
source "${SCRIPT_DIR}/harbor-infra-mirror.sh"
source "${SCRIPT_DIR}/harbor_immutable_release.sh"
# shellcheck source=release_source_guard.sh
source "${SCRIPT_DIR}/release_source_guard.sh"

push_platform() {
  docker_push backend/Dockerfile backend
  docker_push marketplace/Dockerfile marketplace
  docker_push clients/marketplace-web/Dockerfile marketplace-web
  docker_push relay/Dockerfile relay
  docker_push clients/web/Dockerfile web
  docker_push clients/web-admin/Dockerfile web-admin
  docker_push clients/mobile-lovable/Dockerfile mobile
  write_platform_release
}

push_marketplace_core() {
  docker_push backend/Dockerfile backend
  docker_push marketplace/Dockerfile marketplace
  docker_push clients/marketplace-web/Dockerfile marketplace-web
  docker_push clients/web/Dockerfile web
  write_platform_release
}

push_video_expert() {
  docker_push backend/Dockerfile backend
  docker_push marketplace/Dockerfile marketplace
  docker_push clients/marketplace-web/Dockerfile marketplace-web
  docker_push clients/web/Dockerfile web
  docker_push clients/web-admin/Dockerfile web-admin
  write_platform_release
}

push_mobile_access() {
  docker_push backend/Dockerfile backend
  docker_push relay/Dockerfile relay
  docker_push clients/mobile-lovable/Dockerfile mobile
}

push_web() {
  docker_push clients/web/Dockerfile web
  write_platform_release
}

push_marketplace_web() {
  docker_push clients/marketplace-web/Dockerfile marketplace-web
  write_platform_release
}

push_infra() {
  mirror pgvector/pgvector:pg16 pgvector:pg16
  mirror redis:7-alpine         redis:7-alpine
  mirror minio/minio:latest     minio:latest
  mirror minio/mc:latest        mc:latest
  mirror alpine/k8s:1.28.4      kubectl:1.28
}

push_runners() {
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh claude-code )
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh codex-cli )
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh gemini-cli )
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh minimax-cli )
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh grok-build )
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh openclaw )
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh hermes )
  ( cd "${REPO_ROOT}" && REQUIRE_DO_AGENT_BINARY=1 FORCE_REBUILD=1 bash docker/agent-runtime/build.sh do-agent )
  docker build --platform linux/amd64 --target runtime \
    -f "${REPO_ROOT}/docker/agent-runtime/Dockerfile" \
    --build-arg AGENT_RUNTIME=e2e-echo \
    --build-arg "NODE_BASE_IMAGE=${NODE_BASE_IMAGE:-node:24-bookworm-slim}" \
    --build-arg "PYTHON_BASE_IMAGE=${PYTHON_BASE_IMAGE:-python:3.11-slim-bookworm}" \
    -t do-worker/runner-e2e-echo:latest \
    "${REPO_ROOT}/docker/agent-runtime/_context"
  if docker image inspect "do-worker/runner-minimax-cli:latest" >/dev/null 2>&1; then
    :
  elif docker image inspect "l8ai/runner-minimax-cli:latest" >/dev/null 2>&1; then
    docker tag "l8ai/runner-minimax-cli:latest" "do-worker/runner-minimax-cli:latest"
  else
    docker build --platform linux/amd64 --target runtime \
      -f "${REPO_ROOT}/docker/agent-runtime/Dockerfile" \
      --build-arg AGENT_RUNTIME=minimax-cli \
      --build-arg "NODE_BASE_IMAGE=${NODE_BASE_IMAGE:-node:24-bookworm-slim}" \
      --build-arg "PYTHON_BASE_IMAGE=${PYTHON_BASE_IMAGE:-python:3.11-slim-bookworm}" \
      -t do-worker/runner-minimax-cli:latest \
      "${REPO_ROOT}/docker/agent-runtime/_context"
  fi
  for rt in claude-code codex-cli gemini-cli do-agent grok-build openclaw hermes e2e-echo minimax-cli; do
    docker tag "do-worker/runner-${rt}:latest" "${PROJ}/runner-${rt}:latest"
    docker push "${PROJ}/runner-${rt}:latest"
  done
}
main() {
  release_require_pushed_clean_tree "${REPO_ROOT}"
  ensure_project
  harbor_require_upload_token_expiration "${REG}" 120
  case "${TARGET}" in
    platform) push_platform ;;
    marketplace-core) push_marketplace_core ;;
    video-expert) push_video_expert ;;
    mobile-access) push_mobile_access ;;
    web)      push_web ;;
    marketplace-web) push_marketplace_web ;;
    infra)    push_infra; write_platform_release ;;
    runners)  bash "${SCRIPT_DIR}/push-runner-images.sh" all ;;
    do-agent) bash "${SCRIPT_DIR}/push-runner-images.sh" do-agent ;;
    video-runtime) push_video_runtime ;;
    all)      push_infra; bash "${SCRIPT_DIR}/push-runner-images.sh" all defer-platform-source-metadata; push_platform ;;
    *) echo "usage: $0 [all|platform|marketplace-core|video-expert|mobile-access|web|marketplace-web|infra|runners|do-agent|video-runtime]" >&2; exit 1 ;;
  esac
  echo "==> done: ${TARGET}"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi

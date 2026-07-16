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
source "$(dirname "${BASH_SOURCE[0]}")/harbor-image-publishing.sh"

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
  PLATFORM_DIGEST_RELAY="$(registry_digest relay)"
  PLATFORM_DIGEST_WEB_ADMIN="$(registry_digest web-admin)"
  write_platform_release
}

push_video_expert() {
  docker_push backend/Dockerfile backend
  docker_push marketplace/Dockerfile marketplace
  docker_push clients/marketplace-web/Dockerfile marketplace-web
  docker_push clients/web/Dockerfile web
  docker_push clients/web-admin/Dockerfile web-admin
  PLATFORM_DIGEST_RELAY="$(registry_digest relay)"
  write_platform_release
}

push_mobile_access() {
  docker_push backend/Dockerfile backend
  docker_push relay/Dockerfile relay
  docker_push clients/mobile-lovable/Dockerfile mobile
}

push_web() {
  docker_push clients/web/Dockerfile web
  PLATFORM_DIGEST_BACKEND="$(registry_digest backend)"
  PLATFORM_DIGEST_MARKETPLACE="$(registry_digest marketplace)"
  PLATFORM_DIGEST_MARKETPLACE_WEB="$(registry_digest marketplace-web)"
  PLATFORM_DIGEST_RELAY="$(registry_digest relay)"
  PLATFORM_DIGEST_WEB_ADMIN="$(registry_digest web-admin)"
  write_platform_release
}

push_marketplace_web() {
  docker_push clients/marketplace-web/Dockerfile marketplace-web
  PLATFORM_DIGEST_BACKEND="$(registry_digest backend)"
  PLATFORM_DIGEST_MARKETPLACE="$(registry_digest marketplace)"
  PLATFORM_DIGEST_RELAY="$(registry_digest relay)"
  PLATFORM_DIGEST_WEB="$(registry_digest web)"
  PLATFORM_DIGEST_WEB_ADMIN="$(registry_digest web-admin)"
  write_platform_release
}

push_infra() {
  mirror pgvector/pgvector:pg16 pgvector:pg16
  mirror redis:7-alpine         redis:7-alpine
  mirror minio/minio:latest     minio:latest
  mirror minio/mc:latest        mc:latest
  mirror alpine/k8s:1.28.4      kubectl:1.28
}

push_video_runtime() {
  (
    cd "${REPO_ROOT}"
    FORCE_REBUILD=1 PLATFORM="${PLATFORM}" bash docker/agent-runtime/build.sh video-studio
  )
  docker tag "do-worker/runner-video-studio:latest" "${PROJ}/runner-video-studio:latest"
  docker push "${PROJ}/runner-video-studio:latest"
  local digest
  digest="$(manifest_digest "${PROJ}/runner-video-studio:latest")"
  node "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/update-video-runtime-digest.mjs" \
    "${digest}" "${REPO_ROOT}"
  (
    cd "${REPO_ROOT}"
    RUNTIME_PLATFORM="${PLATFORM}" node scripts/probe-worker-runtime-locks.mjs video-studio
    pnpm run worker-docs:sync
    RUNTIME_PLATFORM="${PLATFORM}" \
      bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-runtime-lock-probes.sh \
      video-studio
    pnpm run worker-docs:check
    jq -e --arg platform "${PLATFORM}" '
      .probes[] |
      select(.worker_slug == "video-studio") |
      .status == "available" and .platform == $platform
    ' tools/loops/worker-onboarding/catalog-loop/evidence/runtime-lock-probes.json \
      >/dev/null
  )
}
main() {
  ensure_project
  case "${TARGET}" in
    platform) push_platform ;;
    marketplace-core) push_marketplace_core ;;
    video-expert) push_video_expert ;;
    mobile-access) push_mobile_access ;;
    web)      push_web ;;
    marketplace-web) push_marketplace_web ;;
    infra)    push_infra ;;
    runners)  bash "${SCRIPT_DIR}/push-runner-images.sh" all ;;
    do-agent) bash "${SCRIPT_DIR}/push-runner-images.sh" do-agent ;;
    video-runtime) push_video_runtime ;;
    all)      push_platform; push_infra; bash "${SCRIPT_DIR}/push-runner-images.sh" all ;;
    *) echo "usage: $0 [all|platform|marketplace-core|video-expert|mobile-access|web|marketplace-web|infra|runners|do-agent|video-runtime]" >&2; exit 1 ;;
  esac
  echo "==> done: ${TARGET}"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi

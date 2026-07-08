#!/usr/bin/env bash
# Build + push every image AgentsMesh needs into the cluster Harbor so pods
# on doops-114-k8s pull from repo.aiedulab.cn:8443/agentsmesh/* (fast, node-local).
#
#   ./push-images.sh all        # platform + infra + runners
#   ./push-images.sh platform   # backend/relay/web/web-admin only
#   ./push-images.sh infra      # postgres/redis/minio/mc/kubectl mirrors
#   ./push-images.sh runners    # agent-runtime images (claude/codex/gemini/e2e-echo)
#
# The build host must already be `docker login repo.aiedulab.cn:8443`.
set -euo pipefail

REG="repo.aiedulab.cn:8443"
PROJ="${REG}/agentsmesh"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
TARGET="${1:-all}"

# Cluster nodes are linux/amd64; this build host may be arm64 (Apple Silicon),
# so every image must be forced to amd64 or pods hit `exec format error`.
# `pure` builds static Go binaries (no cgo) so the resolution-only stub CC
# toolchain (//tools/crosscc) is never invoked during the cross build.
AMD64_FLAGS=(--platforms=@rules_go//go/toolchain:linux_amd64 --@rules_go//go/config:pure)

# Reads Harbor user/pass from the docker credential store (credsStore), since
# Docker Desktop keeps the secret out of config.json.
harbor_creds() {
  local store
  store="$(python3 -c "import json,os;print(json.load(open(os.path.expanduser('~/.docker/config.json'))).get('credsStore',''))")"
  echo "${REG}" | "docker-credential-${store}" get
}

ensure_project() {
  local cred u p
  cred="$(harbor_creds)"
  u="$(echo "${cred}" | python3 -c "import sys,json;print(json.load(sys.stdin)['Username'])")"
  p="$(echo "${cred}" | python3 -c "import sys,json;print(json.load(sys.stdin)['Secret'])")"
  echo "==> ensuring Harbor project agentsmesh"
  curl -sk -u "${u}:${p}" -o /dev/null -w "  create project -> HTTP %{http_code}\n" \
    -X POST "https://${REG}/api/v2.0/projects" \
    -H "Content-Type: application/json" \
    -d '{"project_name":"agentsmesh","public":true}' || true
}

# bazel_push <bazel_tarball_target> <loaded_local_tag> <dest_repo>
bazel_push() {
  local target="$1" local_tag="$2" dest="$3"
  echo "==> bazel build+load ${target} (linux/amd64)"
  ( cd "${REPO_ROOT}" && bazel run "${target}" "${AMD64_FLAGS[@]}" )
  docker tag "${local_tag}" "${PROJ}/${dest}:latest"
  docker push "${PROJ}/${dest}:latest"
}

# mirror <public_image> <dest_repo:tag>
# Uses `buildx imagetools create` to copy the source's full multi-arch manifest
# list straight into Harbor. `docker pull --platform` is unreliable on Apple
# Silicon (it reuses the cached arm64 layer), which would push an arm64-only
# image and break the amd64 nodes with `exec format error`. Copying the index
# lets the amd64 nodes auto-select their variant.
mirror() {
  local src="$1" dest="$2"
  echo "==> mirror ${src} -> ${PROJ}/${dest} (multi-arch index)"
  local n=1
  until docker buildx imagetools create --tag "${PROJ}/${dest}" "${src}"; do
    [[ "${n}" -ge 4 ]] && { echo "  imagetools failed after 4 tries: ${src}" >&2; return 1; }
    echo "  retry ${n}/4 in 8s..." >&2; sleep 8; n=$((n+1))
  done
}

push_platform() {
  bazel_push //backend/cmd/server:image_tarball agentsmesh/server:latest backend
  bazel_push //relay/cmd/relay:image_tarball     agentsmesh/relay:latest  relay
  bazel_push //clients/web:image_tarball         agentsmesh/image:latest  web
  bazel_push //clients/web-admin:image_tarball   agentsmesh/image:latest  web-admin
}

push_infra() {
  mirror pgvector/pgvector:pg16 pgvector:pg16
  mirror redis:7-alpine         redis:7-alpine
  mirror minio/minio:latest     minio:latest
  mirror minio/mc:latest        mc:latest
  mirror alpine/k8s:1.28.4      kubectl:1.28
}

push_runners() {
  # build_claude_code stages binaries + builds the shared base; codex/gemini reuse cache.
  ( cd "${REPO_ROOT}" && bazel run //docker/agent-runtime:build_claude_code )
  ( cd "${REPO_ROOT}" && bazel run //docker/agent-runtime:build_codex_cli )
  ( cd "${REPO_ROOT}" && bazel run //docker/agent-runtime:build_gemini_cli )
  # e2e-echo has no bazel target (dev/CI only); build it from the staged _context.
  docker build --platform linux/amd64 --target runtime \
    -f "${REPO_ROOT}/docker/agent-runtime/Dockerfile" \
    --build-arg AGENT_RUNTIME=e2e-echo \
    -t do-worker/runner-e2e-echo:latest \
    "${REPO_ROOT}/docker/agent-runtime/_context"
  for rt in claude-code codex-cli gemini-cli e2e-echo; do
    docker tag "do-worker/runner-${rt}:latest" "${PROJ}/runner-${rt}:latest"
    docker push "${PROJ}/runner-${rt}:latest"
  done
}

ensure_project
case "${TARGET}" in
  platform) push_platform ;;
  infra)    push_infra ;;
  runners)  push_runners ;;
  all)      push_platform; push_infra; push_runners ;;
  *) echo "usage: $0 [all|platform|infra|runners]" >&2; exit 1 ;;
esac
echo "==> done: ${TARGET}"

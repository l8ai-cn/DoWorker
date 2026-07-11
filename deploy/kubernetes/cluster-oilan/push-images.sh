#!/usr/bin/env bash
# Build + push every image Do Worker needs into the cluster Harbor so pods
# on doops-114-k8s pull from repo.aiedulab.cn:8443/agentsmesh/* (fast, node-local).
#
#   ./push-images.sh all        # platform + infra + runners
#   ./push-images.sh platform   # backend/relay/web/web-admin only
#   ./push-images.sh infra      # postgres/redis/minio/mc/kubectl mirrors
#   ./push-images.sh runners    # agent-runtime images (claude/codex/gemini/grok/openclaw/hermes/e2e-echo)
#
# The build host must already be `docker login repo.aiedulab.cn:8443`.
set -euo pipefail

REG="repo.aiedulab.cn:8443"
PROJ="${REG}/agentsmesh"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
TARGET="${1:-all}"
PLATFORM="${PLATFORM:-linux/amd64}"

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

# docker_push <dockerfile> <dest_repo>
docker_push() {
  local file="$1" dest="$2"
  local tag="${PROJ}/${dest}:latest"
  echo "==> docker build ${file} -> ${tag} (${PLATFORM})"
  docker build --platform "${PLATFORM}" \
    -f "${REPO_ROOT}/${file}" \
    -t "${tag}" \
    "${REPO_ROOT}"
  docker push "${tag}"
}

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
  docker_push backend/Dockerfile backend
  docker_push relay/Dockerfile relay
  docker_push clients/web/Dockerfile web
  docker_push clients/web-admin/Dockerfile web-admin
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
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh grok-build )
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh openclaw )
  ( cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh hermes )
  docker build --platform linux/amd64 --target runtime \
    -f "${REPO_ROOT}/docker/agent-runtime/Dockerfile" \
    --build-arg AGENT_RUNTIME=e2e-echo \
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
      -t do-worker/runner-minimax-cli:latest \
      "${REPO_ROOT}/docker/agent-runtime/_context"
  fi
  for rt in claude-code codex-cli gemini-cli grok-build openclaw hermes e2e-echo minimax-cli; do
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

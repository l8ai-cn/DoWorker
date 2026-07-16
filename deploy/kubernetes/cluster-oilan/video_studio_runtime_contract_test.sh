#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
COORDINATOR_RUNNERS="${ROOT}/deploy/dev/lib/coordinator_runners.sh"
BACKEND_MANIFEST="${ROOT}/deploy/kubernetes/cluster-oilan/30-backend.yaml"
PREPULL_MANIFEST="${ROOT}/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"
PUSH_IMAGES="${ROOT}/deploy/kubernetes/cluster-oilan/push-images.sh"
PUBLISHING="${ROOT}/deploy/kubernetes/cluster-oilan/harbor-image-publishing.sh"
IMAGE="repo.aiedulab.cn:8443/agentsmesh/runner-video-studio"
DIGEST="sha256:21bc8f143d304f361b58862b9c33b4e7257307132f9507cf22d461fca1d61716"

(
  ENV_FILE="$(mktemp)"
  trap 'rm -f "$ENV_FILE"' EXIT
  printf 'BACKEND_GRPC_PORT=10016\n' > "$ENV_FILE"
  SCRIPT_DIR="${ROOT}/deploy/dev"
  source "$COORDINATOR_RUNNERS"
  export_coordinator_runner_env
  [[ ",${COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES}," == *",video-studio=runner-video-studio,"* ]]
)

grep -Fq "video-studio=${IMAGE}@${DIGEST}" "$BACKEND_MANIFEST"
grep -Fq "image: ${IMAGE}:latest" "$PREPULL_MANIFEST"

push_runners="$(
  awk '
    /^push_runners\(\)/ { capture=1 }
    capture { print }
    capture && /^}/ { exit }
  ' "$PUSH_IMAGES"
)"
grep -Fq 'bash docker/agent-runtime/build.sh video-studio' <<< "$push_runners"
grep -Eq 'for rt in .*codex-cli video-studio .*; do' <<< "$push_runners"

push_video_expert="$(
  awk '
    /^push_video_expert\(\)/ { capture=1 }
    capture { print }
    capture && /^}/ { exit }
  ' "$PUSH_IMAGES"
)"
for image in backend marketplace marketplace-web web web-admin; do
  grep -Fq "docker_push" <<< "$push_video_expert"
  grep -Fq "$image" <<< "$push_video_expert"
done
! grep -Fq "mobile" <<< "$push_video_expert"
grep -Fq 'harbor-image-publishing.sh' "$PUSH_IMAGES"
grep -Fq 'docker_build_with_retry' "$PUBLISHING"
grep -Fq 'docker build failed; retry' "$PUBLISHING"
grep -Fq 'docker_build_with_heartbeat' "$PUBLISHING"
grep -Fq 'docker build still running' "$PUBLISHING"
grep -Fq 'push_video_runtime' "$PUSH_IMAGES"
grep -Fq 'runner-video-studio:latest' "$PUSH_IMAGES"

for dockerfile in \
  clients/marketplace-web/Dockerfile \
  clients/web/Dockerfile \
  clients/web-admin/Dockerfile; do
  corepack_line="$(grep -n 'corepack prepare' "${ROOT}/${dockerfile}" | cut -d: -f1)"
  registry_line="$(grep -n 'COREPACK_NPM_REGISTRY' "${ROOT}/${dockerfile}" | cut -d: -f1)"
  [[ -n "$registry_line" && "$registry_line" -lt "$corepack_line" ]]
done

grep -Fq 'wasm-pack/releases/download/v0.13.1' "$ROOT/clients/web/Dockerfile"
grep -Fq 'wasm-bindgen/releases/download/0.2.105' "$ROOT/clients/web/Dockerfile"
! grep -Fq 'cargo install wasm-bindgen-cli' "$ROOT/clients/web/Dockerfile"
grep -Fq 'NEXT_BUILD_CPUS=1' "$ROOT/clients/web/Dockerfile"
grep -Fq 'NODE_OPTIONS=--max-old-space-size=2048' "$ROOT/clients/web/Dockerfile"
grep -Fq 'process.env.NEXT_BUILD_CPUS' "$ROOT/clients/web/next.config.ts"
! grep -Fq 'next/font/google' "$ROOT/clients/web/src/app/layout.tsx"
grep -Fq 'geist/font/sans' "$ROOT/clients/web/src/app/layout.tsx"
grep -Fq '@fontsource-variable/space-grotesk/wght.css' "$ROOT/clients/web/src/app/layout.tsx"

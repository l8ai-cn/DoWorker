#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
COORDINATOR_RUNNERS="${ROOT}/deploy/dev/lib/coordinator_runners.sh"
BACKEND_MANIFEST="${ROOT}/deploy/kubernetes/cluster-oilan/30-backend.yaml"
PREPULL_MANIFEST="${ROOT}/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"
PUSH_IMAGES="${ROOT}/deploy/kubernetes/cluster-oilan/push-images.sh"
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

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
COORDINATOR_RUNNERS="${ROOT}/deploy/dev/lib/coordinator_runners.sh"
BACKEND_MANIFEST="${ROOT}/deploy/kubernetes/cluster-oilan/30-backend.yaml"
PREPULL_MANIFEST="${ROOT}/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"
PUSH_IMAGES="${ROOT}/deploy/kubernetes/cluster-oilan/push-images.sh"
PUBLISHING="${ROOT}/deploy/kubernetes/cluster-oilan/harbor-image-publishing.sh"
UPDATE_DIGEST="${ROOT}/deploy/kubernetes/cluster-oilan/update-video-runtime-digest.mjs"
RUNTIME_LOCK="${ROOT}/backend/internal/domain/workerruntime/runtime_catalog.lock.json"
IMAGE="repo.aiedulab.cn:8443/agentsmesh/runner-video-studio"
DIGEST="$(
  node -e '
    const catalog = require(process.argv[1]);
    const image = catalog.images.find((entry) => entry.slug === "video-studio-stable");
    if (!image) process.exit(1);
    process.stdout.write(image.digest);
  ' "$RUNTIME_LOCK"
)"
NEXT_DIGEST="sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
if [[ "$DIGEST" == "$NEXT_DIGEST" ]]; then
  NEXT_DIGEST="sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
fi

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
grep -Fq "image: ${IMAGE}@${DIGEST}" "$PREPULL_MANIFEST"

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
grep -Fq 'FORCE_REBUILD=1 PLATFORM="${PLATFORM}"' "$PUSH_IMAGES"
grep -Fq 'digest="$(manifest_digest "${PROJ}/runner-video-studio:latest")"' "$PUSH_IMAGES"
grep -Fq 'update-video-runtime-digest.mjs' "$PUSH_IMAGES"
push_video_runtime="$(
  awk '
    /^push_video_runtime\(\)/ { capture=1 }
    capture { print }
    capture && /^}/ { exit }
  ' "$PUSH_IMAGES"
)"
grep -Fq 'node scripts/probe-worker-runtime-locks.mjs video-studio' <<< "$push_video_runtime"
grep -Fq 'pnpm run worker-docs:sync' <<< "$push_video_runtime"
grep -Fq 'verify-runtime-lock-probes.sh' <<< "$push_video_runtime"
grep -Fq '.status == "available"' <<< "$push_video_runtime"

FIXTURE_ROOT="$(mktemp -d)"
trap 'rm -rf "$FIXTURE_ROOT"' EXIT
mkdir -p \
  "$FIXTURE_ROOT/backend/internal/domain/workerruntime" \
  "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan"
cp "$RUNTIME_LOCK" "$FIXTURE_ROOT/backend/internal/domain/workerruntime/runtime_catalog.lock.json"
cp "$BACKEND_MANIFEST" "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/30-backend.yaml"
cp "$PREPULL_MANIFEST" "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"

node "$UPDATE_DIGEST" "$NEXT_DIGEST" "$FIXTURE_ROOT"
grep -Fq "\"digest\": \"$NEXT_DIGEST\"" \
  "$FIXTURE_ROOT/backend/internal/domain/workerruntime/runtime_catalog.lock.json"
grep -Fq "\"reference\": \"$IMAGE@$NEXT_DIGEST\"" \
  "$FIXTURE_ROOT/backend/internal/domain/workerruntime/runtime_catalog.lock.json"
grep -Fq "video-studio=$IMAGE@$NEXT_DIGEST" \
  "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/30-backend.yaml"
grep -Fq "image: $IMAGE@$NEXT_DIGEST" \
  "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"

FIXTURE_HASHES="$(find "$FIXTURE_ROOT" -type f -print0 | sort -z | xargs -0 shasum -a 256)"
node "$UPDATE_DIGEST" "$NEXT_DIGEST" "$FIXTURE_ROOT"
[[ "$FIXTURE_HASHES" == "$(find "$FIXTURE_ROOT" -type f -print0 | sort -z | xargs -0 shasum -a 256)" ]]

LOCK_PATH="$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/.video-runtime-digest-update.lock"
mkdir "$LOCK_PATH"
printf '%s\n' "$$" > "$LOCK_PATH/owner"
FIXTURE_HASHES="$(find "$FIXTURE_ROOT" -type f ! -path "$LOCK_PATH/*" -print0 | sort -z | xargs -0 shasum -a 256)"
if node "$UPDATE_DIGEST" "$DIGEST" "$FIXTURE_ROOT" >/dev/null 2>&1; then
  echo "concurrent digest update unexpectedly succeeded" >&2
  exit 1
fi
[[ "$FIXTURE_HASHES" == "$(find "$FIXTURE_ROOT" -type f ! -path "$LOCK_PATH/*" -print0 | sort -z | xargs -0 shasum -a 256)" ]]
rm -rf "$LOCK_PATH"

TRANSACTION_PATH="$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/.video-runtime-digest-transaction"
mkdir "$TRANSACTION_PATH"
cp "$FIXTURE_ROOT/backend/internal/domain/workerruntime/runtime_catalog.lock.json" \
  "$TRANSACTION_PATH/0.backup"
cp "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/30-backend.yaml" \
  "$TRANSACTION_PATH/1.backup"
cp "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml" \
  "$TRANSACTION_PATH/2.backup"
cat > "$TRANSACTION_PATH/manifest.json" <<'JSON'
{
  "version": 1,
  "records": [
    {
      "target": "backend/internal/domain/workerruntime/runtime_catalog.lock.json",
      "backup": "0.backup"
    },
    {
      "target": "deploy/kubernetes/cluster-oilan/30-backend.yaml",
      "backup": "1.backup"
    },
    {
      "target": "deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml",
      "backup": "2.backup"
    }
  ]
}
JSON
sed -i.bak "s/$NEXT_DIGEST/$DIGEST/" \
  "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"
rm "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml.bak"
node "$UPDATE_DIGEST" "$DIGEST" "$FIXTURE_ROOT"
[[ ! -e "$TRANSACTION_PATH" ]]
grep -Fq "\"digest\": \"$DIGEST\"" \
  "$FIXTURE_ROOT/backend/internal/domain/workerruntime/runtime_catalog.lock.json"
grep -Fq "video-studio=$IMAGE@$DIGEST" \
  "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/30-backend.yaml"
grep -Fq "image: $IMAGE@$DIGEST" \
  "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"

FIXTURE_HASHES="$(find "$FIXTURE_ROOT" -type f -print0 | sort -z | xargs -0 shasum -a 256)"
if node "$UPDATE_DIGEST" invalid-digest "$FIXTURE_ROOT" >/dev/null 2>&1; then
  echo "invalid digest update unexpectedly succeeded" >&2
  exit 1
fi
[[ "$FIXTURE_HASHES" == "$(find "$FIXTURE_ROOT" -type f -print0 | sort -z | xargs -0 shasum -a 256)" ]]

sed -i.bak "s/$DIGEST/$NEXT_DIGEST/" \
  "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"
rm "$FIXTURE_ROOT/deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml.bak"
FIXTURE_HASHES="$(find "$FIXTURE_ROOT" -type f -print0 | sort -z | xargs -0 shasum -a 256)"
if node "$UPDATE_DIGEST" "$NEXT_DIGEST" "$FIXTURE_ROOT" >/dev/null 2>&1; then
  echo "inconsistent digest update unexpectedly succeeded" >&2
  exit 1
fi
[[ "$FIXTURE_HASHES" == "$(find "$FIXTURE_ROOT" -type f -print0 | sort -z | xargs -0 shasum -a 256)" ]]

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

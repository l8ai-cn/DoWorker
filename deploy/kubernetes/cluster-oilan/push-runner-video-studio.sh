#!/usr/bin/env bash

publish_video_runtime_metadata() {
  local digest observed_at
  digest="$(platform_manifest_digest "${PROJ}/runner-video-studio:latest")"
  observed_at="$(git -C "${REPO_ROOT}" show -s --format=%cI "${RELEASE_SOURCE_COMMIT}")"
  docker run --rm --pull=never --platform "${PLATFORM:-linux/amd64}" \
    -v "${REPO_ROOT}/docker/agent-runtime/video_contract_test.sh:/tmp/video_contract_test.sh:ro" \
    --entrypoint bash agent-cloud/runner-video-studio:latest \
    /tmp/video_contract_test.sh
  RUNTIME_OBSERVED_AT="${observed_at}" \
    node "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/update-video-runtime-digest.mjs" \
      "${digest}" "${REPO_ROOT}"
  (
    cd "${REPO_ROOT}"
    RUNTIME_OBSERVED_AT="${observed_at}" \
      RUNTIME_PLATFORM="${PLATFORM:-linux/amd64}" \
      node scripts/probe-worker-runtime-locks.mjs video-studio
    pnpm run worker-docs:sync
    node scripts/generate-worker-loop-inventory.mjs
    RUNTIME_PLATFORM="${PLATFORM:-linux/amd64}" \
      bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-runtime-lock-probes.sh \
      video-studio
    pnpm run worker-docs:check
    node scripts/generate-worker-loop-inventory.mjs --check
    jq -e --arg platform "${PLATFORM:-linux/amd64}" '
      .probes[] |
      select(.worker_slug == "video-studio") |
      .status == "available" and .platform == $platform
    ' tools/loops/worker-onboarding/catalog-loop/evidence/runtime-lock-probes.json \
      >/dev/null
  )
  [[ "${RUNNER_SOURCE_METADATA_MODE}" == "defer-platform-source-metadata" ]] || release_write_source_metadata "${REPO_ROOT}"
}

push_video_studio() {
  (
    cd "${REPO_ROOT}"
    FORCE_REBUILD=1 PLATFORM="${PLATFORM:-linux/amd64}" \
      bash docker/agent-runtime/build.sh video-studio
  )
  push_runtime video-studio
  publish_video_runtime_metadata
}

#!/usr/bin/env bash
set -euo pipefail
REG="repo.aiedulab.cn:8443"
PROJ="${REG}/agentsmesh"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
TARGET="${1:-all}"; RUNNER_SOURCE_METADATA_MODE="${2:-write}"
source "${REPO_ROOT}/docker/agent-runtime/do_agent_release_manifest.sh"
source "${SCRIPT_DIR}/harbor_immutable_release.sh"
source "${SCRIPT_DIR}/runner-build-base.sh"
source "${SCRIPT_DIR}/release_source_guard.sh"
source "${SCRIPT_DIR}/push-runner-video-studio.sh"
release_require_pushed_clean_tree "${REPO_ROOT}"
[[ "${RUNNER_SOURCE_METADATA_MODE}" == "write" || ( "${TARGET}" == "all" && "${RUNNER_SOURCE_METADATA_MODE}" == "defer-platform-source-metadata" ) ]] || { echo "invalid Runner source metadata mode" >&2; exit 1; }

require_do_agent_artifact() {
  local expected actual
  [[ -n "${DO_AGENT_BINARY:-}" ]] || {
    echo "DO_AGENT_BINARY must point to the approved linux/amd64 artifact" >&2
    return 1
  }
  expected="$(do_agent_release_value artifact.binary_sha256)"
  do_agent_require_digest "artifact.binary_sha256" "${expected}"
  actual="$(do_agent_sha256 "${DO_AGENT_BINARY}")"
  [[ "${actual}" == "${expected}" ]] || {
    echo "do-agent artifact does not match the trusted release manifest" >&2
    return 1
  }
  if [[ -n "${DO_AGENT_BINARY_SHA256:-}" && "${DO_AGENT_BINARY_SHA256}" != "${expected}" ]]; then
    echo "DO_AGENT_BINARY_SHA256 does not match the trusted release manifest" >&2
    return 1
  fi
}

build_do_agent() {
  local source_commit source_date_epoch expected_hash
  require_do_agent_artifact
  source_commit="$(do_agent_release_value source.commit)"
  source_date_epoch="$(do_agent_release_value build.source_date_epoch)"
  expected_hash="$(do_agent_release_value artifact.binary_sha256)"
  (
    cd "${REPO_ROOT}"
    FORCE_REBUILD=1 \
      REQUIRE_DO_AGENT_BINARY=1 \
      DO_AGENT_SOURCE_COMMIT="${source_commit}" \
      DO_AGENT_BINARY_SHA256="${expected_hash}" \
      SOURCE_DATE_EPOCH="${source_date_epoch}" \
      bash docker/agent-runtime/build.sh do-agent
  )
}

push_runtime() {
  local runtime="$1"
  verify_runtime_source_revision "${runtime}"
  docker tag "do-worker/runner-${runtime}:latest" "${PROJ}/runner-${runtime}:latest"
  docker push "${PROJ}/runner-${runtime}:latest"
}

verify_runtime_source_revision() {
  local runtime="$1"
  local revision
  revision="$(docker image inspect "do-worker/runner-${runtime}:latest" \
    --format '{{ index .Config.Labels "org.opencontainers.image.revision" }}')"
  [[ "${revision}" == "${RELEASE_SOURCE_COMMIT}" ]] || {
    echo "runner-${runtime} source revision mismatch: ${revision}" >&2
    return 1
  }
}

verify_do_agent_labels() {
  local source_commit expected_hash actual_source actual_hash
  source_commit="$(do_agent_release_value source.commit)"
  expected_hash="$(do_agent_release_value artifact.binary_sha256)"
  actual_source="$(docker image inspect do-worker/runner-do-agent:latest \
    --format '{{ index .Config.Labels "ai.agentsmesh.do-agent.source-revision" }}')"
  actual_hash="$(docker image inspect do-worker/runner-do-agent:latest \
    --format '{{ index .Config.Labels "ai.agentsmesh.do-agent.binary-sha256" }}')"
  [[ "${actual_source}" == "${source_commit}" && "${actual_hash}" == "${expected_hash}" ]] || {
    echo "do-agent image labels do not match the trusted artifact manifest" >&2
    return 1
  }
}

manifest_digest() {
  local image="$1" digest="" attempt=1
  until digest="$(docker buildx imagetools inspect "${image}" --format '{{.Manifest.Digest}}')" &&
    [[ "${digest}" =~ ^sha256:[a-f0-9]{64}$ ]]; do
    [[ "${attempt}" -ge 4 ]] && {
      echo "invalid registry digest for ${image}: ${digest}" >&2
      return 1
    }
    sleep 3
    attempt=$((attempt + 1))
  done
  printf '%s' "${digest}"
}

publish_do_agent() {
  local repository release_tag candidate_tag candidate_digest release_digest latest_digest
  local source_commit observed_at
  repository="$(do_agent_release_value image.repository)"
  source_commit="$(do_agent_release_value source.commit)"
  release_tag="${source_commit:0:12}-runner-${RELEASE_SOURCE_COMMIT:0:12}"
  candidate_tag="candidate-${RELEASE_SOURCE_COMMIT:0:12}"
  [[ "${repository}" == "${PROJ}/runner-do-agent" ]] || {
    echo "do-agent release repository must be ${PROJ}/runner-do-agent" >&2
    return 1
  }

  verify_runtime_source_revision do-agent
  verify_do_agent_labels
  docker tag "do-worker/runner-do-agent:latest" "${repository}:${candidate_tag}"
  docker push "${repository}:${candidate_tag}"
  candidate_digest="$(manifest_digest "${repository}:${candidate_tag}")"

  harbor_ensure_immutable_tag "${REG}" agentsmesh runner-do-agent "${release_tag}"
  if release_digest="$(docker buildx imagetools inspect "${repository}:${release_tag}" \
    --format '{{.Manifest.Digest}}' 2>/dev/null)"; then
    [[ "${release_digest}" == "${candidate_digest}" ]] || {
      echo "immutable release tag digest mismatch: ${release_digest}" >&2
      return 1
    }
  else
    docker buildx imagetools create \
      --prefer-index=false \
      --tag "${repository}:${release_tag}" \
      "${repository}@${candidate_digest}"
    release_digest="$(manifest_digest "${repository}:${release_tag}")"
    [[ "${release_digest}" == "${candidate_digest}" ]] || {
      echo "release promotion digest mismatch: ${release_digest}" >&2
      return 1
    }
  fi

  observed_at="$(git -C "${REPO_ROOT}" show -s --format=%cI "${RELEASE_SOURCE_COMMIT}")"
  RUNTIME_OBSERVED_AT="${observed_at}" \
    node "${SCRIPT_DIR}/update-do-agent-runtime-digest.mjs" \
      "${candidate_digest}" "${RELEASE_SOURCE_COMMIT}" "${release_tag}" "${REPO_ROOT}"
  (
    cd "${REPO_ROOT}"
    RUNTIME_OBSERVED_AT="${observed_at}" \
      RUNTIME_PLATFORM="${PLATFORM:-linux/amd64}" \
      node scripts/probe-worker-runtime-locks.mjs do-agent
    RUNTIME_OBSERVED_AT="${observed_at}" node scripts/probe-local-worker-images.mjs do-agent
    RUNTIME_OBSERVED_AT="${observed_at}" node scripts/probe-local-worker-images.mjs seedance-expert
    pnpm run worker-docs:sync
    node scripts/generate-worker-loop-inventory.mjs
    EXPECTED_REMOTE_DIGEST="${candidate_digest}" \
      bash deploy/kubernetes/cluster-oilan/runtime_mapping_contract_test.sh
  )

  docker buildx imagetools create \
    --prefer-index=false \
    --tag "${repository}:latest" \
    "${repository}@${candidate_digest}"
  latest_digest="$(manifest_digest "${repository}:latest")"
  [[ "${latest_digest}" == "${candidate_digest}" ]] || {
    echo "latest promotion digest mismatch: ${latest_digest}" >&2
    return 1
  }
  echo "${repository}@${latest_digest}"
}

push_do_agent() {
  build_do_agent
  publish_do_agent
  [[ "${RUNNER_SOURCE_METADATA_MODE}" == "defer-platform-source-metadata" ]] || release_write_source_metadata "${REPO_ROOT}"
}

push_all() {
  local runtime
  for runtime in claude-code codex-cli video-studio gemini-cli grok-build minimax-cli openclaw hermes; do
    (cd "${REPO_ROOT}" && FORCE_REBUILD=1 bash docker/agent-runtime/build.sh "${runtime}")
  done
  build_do_agent
  docker build --platform linux/amd64 --target runtime \
    --label "org.opencontainers.image.revision=${RELEASE_SOURCE_COMMIT}" \
    -f "${REPO_ROOT}/docker/agent-runtime/Dockerfile" \
    --build-arg AGENT_RUNTIME=e2e-echo \
    --build-arg "RUNTIME_BUILD_BASE=${RUNTIME_BUILD_BASE}" \
    -t do-worker/runner-e2e-echo:latest \
    "${REPO_ROOT}/docker/agent-runtime/_context"
  for runtime in claude-code codex-cli video-studio gemini-cli grok-build minimax-cli openclaw hermes e2e-echo; do
    push_runtime "${runtime}"
  done
  publish_do_agent
  publish_video_runtime_metadata
}

verify_runner_build_base
harbor_require_upload_token_expiration "${REG}" 120
export RUNTIME_BUILD_BASE="${RUNNER_BUILD_BASE}"
case "${TARGET}" in
  all) push_all ;;
  do-agent) push_do_agent ;;
  video-studio) push_video_studio ;;
  *) echo "usage: $0 [all|do-agent|video-studio]" >&2; exit 1 ;;
esac

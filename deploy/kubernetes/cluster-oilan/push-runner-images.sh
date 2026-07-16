#!/usr/bin/env bash
set -euo pipefail

REG="repo.aiedulab.cn:8443"
PROJ="${REG}/agentsmesh"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
TARGET="${1:-all}"
source "${REPO_ROOT}/docker/agent-runtime/do_agent_release_manifest.sh"
source "${SCRIPT_DIR}/harbor_immutable_release.sh"
# shellcheck source=release_source_guard.sh
source "${SCRIPT_DIR}/release_source_guard.sh"

release_require_pushed_clean_tree "${REPO_ROOT}"

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
  docker tag "do-worker/runner-${runtime}:latest" "${PROJ}/runner-${runtime}:latest"
  docker push "${PROJ}/runner-${runtime}:latest"
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
  local repository release_tag candidate_tag expected_digest candidate_digest release_digest latest_digest
  repository="$(do_agent_release_value image.repository)"
  release_tag="$(do_agent_release_value image.tag)"
  candidate_tag="candidate-${release_tag}"
  expected_digest="$(do_agent_release_value image.digest)"
  [[ "${repository}" == "${PROJ}/runner-do-agent" ]] || {
    echo "do-agent release repository must be ${PROJ}/runner-do-agent" >&2
    return 1
  }
  do_agent_require_digest "image.digest" "${expected_digest}"

  docker tag "do-worker/runner-do-agent:latest" "${repository}:${candidate_tag}"
  docker push "${repository}:${candidate_tag}"
  candidate_digest="$(manifest_digest "${repository}:${candidate_tag}")"
  [[ "${candidate_digest}" == "${expected_digest}" ]] || {
    echo "candidate digest mismatch: expected ${expected_digest}, got ${candidate_digest}" >&2
    return 1
  }

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

  docker buildx imagetools create \
    --prefer-index=false \
    --tag "${repository}:latest" \
    "${repository}@${candidate_digest}"
  latest_digest="$(manifest_digest "${repository}:latest")"
  [[ "${latest_digest}" == "${candidate_digest}" ]] || {
    echo "latest promotion digest mismatch: ${latest_digest}" >&2
    return 1
  }
  EXPECTED_REMOTE_DIGEST="${latest_digest}" \
    bash "${SCRIPT_DIR}/runtime_mapping_contract_test.sh"
  echo "${repository}@${latest_digest}"
}

push_do_agent() {
  build_do_agent
  publish_do_agent
}

push_all() {
  local runtime
  for runtime in claude-code codex-cli gemini-cli grok-build openclaw hermes; do
    (cd "${REPO_ROOT}" && bash docker/agent-runtime/build.sh "${runtime}")
  done
  build_do_agent
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
  for runtime in claude-code codex-cli gemini-cli grok-build openclaw hermes e2e-echo minimax-cli; do
    push_runtime "${runtime}"
  done
  publish_do_agent
}

case "${TARGET}" in
  all) push_all ;;
  do-agent) push_do_agent ;;
  *) echo "usage: $0 [all|do-agent]" >&2; exit 1 ;;
esac

#!/usr/bin/env bash

RUNNER_BUILD_BASE_DIGEST="sha256:cb4e8f7c443347358b7875e717c29e27bf9befc8f5a26cf18af3c3dec80e58c5"
RUNNER_BUILD_BASE="${PROJ}/runner-node-base@${RUNNER_BUILD_BASE_DIGEST}"

verify_runner_build_base() {
  local digest raw status
  [[ "${RUNNER_BUILD_BASE_DIGEST}" =~ ^sha256:[a-f0-9]{64}$ ]] || {
    echo "malformed Runner build base digest: ${RUNNER_BUILD_BASE_DIGEST}" >&2
    return 1
  }
  digest="$(docker buildx imagetools inspect "${RUNNER_BUILD_BASE}" \
    --format '{{.Manifest.Digest}}')" || {
    echo "unable to inspect locked Runner build base: ${RUNNER_BUILD_BASE}" >&2
    return 1
  }
  [[ "${digest}" == "${RUNNER_BUILD_BASE_DIGEST}" ]] || {
    echo "Runner build base digest mismatch: ${digest}" >&2
    return 1
  }
  raw="$(docker buildx imagetools inspect "${RUNNER_BUILD_BASE}" --raw)" || {
    echo "unable to inspect Runner build base platforms: ${RUNNER_BUILD_BASE}" >&2
    return 1
  }
  if jq -e '
    def has_linux($arch):
      any(.manifests[]?; .platform.os == "linux" and .platform.architecture == $arch);
    has_linux("amd64") and has_linux("arm64")
  ' <<< "${raw}" >/dev/null; then
    return 0
  else
    status=$?
  fi
  [[ "${status}" -eq 1 ]] || {
    echo "invalid Runner build base manifest JSON" >&2
    return 1
  }
  echo "Runner build base must include linux/amd64 and linux/arm64" >&2
  return 1
}

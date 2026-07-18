#!/usr/bin/env bash
set -euo pipefail

PROJ="registry.example/agentsmesh"
MODE="valid"

docker() {
  local reference="${PROJ}/runner-node-base@${RUNNER_BUILD_BASE_DIGEST}"
  if [[ "$*" == "buildx imagetools inspect ${reference} --format {{.Manifest.Digest}}" ]]; then
    [[ "${MODE}" != "inspect-error" ]] || return 1
    if [[ "${MODE}" == "wrong-digest" ]]; then
      printf '%s\n' "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    else
      printf '%s\n' "${RUNNER_BUILD_BASE_DIGEST}"
    fi
    return
  fi
  if [[ "$*" == "buildx imagetools inspect ${reference} --raw" ]]; then
    case "${MODE}" in
      valid)
        printf '%s\n' '{"manifests":[{"platform":{"os":"linux","architecture":"amd64"}},{"platform":{"os":"linux","architecture":"arm64"}},{"platform":{"os":"linux","architecture":"ppc64le"}}]}'
        ;;
      single)
        printf '%s\n' '{"manifests":[{"platform":{"os":"linux","architecture":"amd64"}}]}'
        ;;
      invalid-json)
        printf '%s\n' 'not-json'
        ;;
      *) return 1 ;;
    esac
    return
  fi
  return 1
}

source "$(dirname "$0")/runner-build-base.sh"

DOCKERFILE="$(dirname "$0")/../../../docker/agent-runtime/Dockerfile"
DEFAULT_DIGEST="$(awk -F@ '/^ARG RUNTIME_BUILD_BASE=/ { print $2 }' "${DOCKERFILE}")"
[[ "${DEFAULT_DIGEST}" == "${RUNNER_BUILD_BASE_DIGEST}" ]]
verify_runner_build_base

for failure_mode in inspect-error wrong-digest single invalid-json; do
  MODE="${failure_mode}"
  if verify_runner_build_base >/dev/null 2>&1; then
    echo "Runner build base validation must reject ${failure_mode}" >&2
    exit 1
  fi
done

MODE="valid"
RUNNER_BUILD_BASE_DIGEST="invalid"
RUNNER_BUILD_BASE="${PROJ}/runner-node-base@${RUNNER_BUILD_BASE_DIGEST}"
if verify_runner_build_base >/dev/null 2>&1; then
  echo "Runner build base validation must reject malformed digests" >&2
  exit 1
fi

#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${DIR}/../../.." && pwd)"
REG="${REG:-repo.aiedulab.cn:8443}"
PROJ="${REG}/agentcloud"
PLATFORM="${PLATFORM:-linux/amd64}"
STAGING_SERVICES=(backend marketplace marketplace-web relay web)

# shellcheck disable=SC1091
source "${DIR}/harbor-image-publishing.sh"
# shellcheck disable=SC1091
source "${DIR}/harbor_immutable_release.sh"
# shellcheck disable=SC1091
source "${DIR}/release_source_guard.sh"

validate_manifest() {
  local manifest="$1" service
  jq -e '
    .schema_version == 1 and
    .platform == "linux/amd64" and
    (.source_commit | test("^[a-f0-9]{40}$")) and
    (.images | keys == ["backend", "marketplace", "marketplace-web", "relay", "web"]) and
    all(
      .images[];
      (.source | type == "string") and
      (.digest | test("^sha256:[a-f0-9]{64}$"))
    )
  ' "${manifest}" >/dev/null || return 1
  for service in "${STAGING_SERVICES[@]}"; do
    jq -e --arg service "${service}" --arg commit "$(jq -r '.source_commit' "${manifest}")" '
      .images[$service].source ==
        ("docker.io/l8ai/doworker-oilan-" + $service + ":sha-" + $commit)
    ' "${manifest}" >/dev/null || {
      echo "unexpected staging source for ${service}" >&2
      return 1
    }
  done
}

verify_staged_image() {
  local reference="$1" expected_commit="$2" platform revision
  docker pull --platform "${PLATFORM}" "${reference}" >/dev/null || return 1
  platform="$(
    docker image inspect "${reference}" --format '{{.Os}}/{{.Architecture}}'
  )" || return 1
  [[ "${platform}" == "${PLATFORM}" ]] || {
    echo "staging image platform mismatch for ${reference}: ${platform}" >&2
    return 1
  }
  revision="$(
    docker image inspect "${reference}" \
      --format '{{ index .Config.Labels "org.opencontainers.image.revision" }}'
  )" || return 1
  [[ "${revision}" == "${expected_commit}" ]] || {
    echo "staging image revision mismatch for ${reference}: ${revision}" >&2
    return 1
  }
}

record_platform_digest() {
  local service="$1" digest="$2"
  case "${service}" in
    backend) export PLATFORM_DIGEST_BACKEND="${digest}" ;;
    marketplace) export PLATFORM_DIGEST_MARKETPLACE="${digest}" ;;
    marketplace-web) export PLATFORM_DIGEST_MARKETPLACE_WEB="${digest}" ;;
    relay) export PLATFORM_DIGEST_RELAY="${digest}" ;;
    web) export PLATFORM_DIGEST_WEB="${digest}" ;;
  esac
}

validate_staged_service() {
  local manifest="$1" service="$2" source expected actual
  source="$(
    jq -er --arg service "${service}" '.images[$service].source' "${manifest}"
  )" || return 1
  expected="$(
    jq -er --arg service "${service}" '.images[$service].digest' "${manifest}"
  )" || return 1
  actual="$(platform_manifest_digest "${source}")" || return 1
  [[ "${actual}" == "${expected}" ]] || {
    echo "staging digest mismatch for ${service}: ${actual} != ${expected}" >&2
    return 1
  }
  verify_staged_image "${source}@${expected}" "${RELEASE_SOURCE_COMMIT}" || return 1
}

promote_service() {
  local manifest="$1" service="$2" source expected destination promoted
  source="$(
    jq -er --arg service "${service}" '.images[$service].source' "${manifest}"
  )" || return 1
  expected="$(
    jq -er --arg service "${service}" '.images[$service].digest' "${manifest}"
  )" || return 1
  destination="${PROJ}/${service}:latest"
  docker tag "${source}@${expected}" "${destination}" || return 1
  docker push "${destination}" >/dev/null || return 1
  promote_platform_manifest "${destination}" "${source}@${expected}" || return 1
  promoted="$(platform_manifest_digest "${destination}")" || return 1
  [[ "${promoted}" == "${expected}" ]] || {
    echo "Harbor promotion digest mismatch for ${service}: ${promoted} != ${expected}" >&2
    return 1
  }
  record_platform_digest "${service}" "${promoted}" || return 1
  echo "==> promoted ${service}: ${promoted}"
}

main() {
  local manifest="${1:?staging manifest path is required}" source_commit service
  [[ -f "${manifest}" ]] || {
    echo "staging manifest not found: ${manifest}" >&2
    return 1
  }
  validate_manifest "${manifest}" || return 1
  source_commit="$(jq -er '.source_commit' "${manifest}")" || return 1
  release_require_pushed_clean_tree "${REPO_ROOT}" || return 1
  [[ "${RELEASE_SOURCE_COMMIT}" == "${source_commit}" ]] || {
    echo "staging source ${source_commit} does not match main ${RELEASE_SOURCE_COMMIT}" >&2
    return 1
  }
  for service in "${STAGING_SERVICES[@]}"; do
    validate_staged_service "${manifest}" "${service}" || return 1
  done
  ensure_project || return 1
  harbor_require_upload_token_expiration "${REG}" 120 || return 1
  for service in "${STAGING_SERVICES[@]}"; do
    promote_service "${manifest}" "${service}" || return 1
  done
  write_platform_release || return 1
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi

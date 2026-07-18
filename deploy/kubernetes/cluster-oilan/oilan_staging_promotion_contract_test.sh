#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT
COMMIT="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
MANIFEST="${TMP}/oilan-staging-release.json"
LOG="${TMP}/promotion.log"
TEST_REPO="${TMP}/repo"
GUARD_COMMIT="${COMMIT}"
DIGEST_OVERRIDE=""
REVISION_OVERRIDE=""
PUSH_FAILURE=false
declare -a SEEDED_SERVICES=()
declare -a PROMOTED_SERVICES=()

jq -n --arg commit "${COMMIT}" '
  def image($service; $digit): {
    source: ("docker.io/l8ai/doworker-oilan-" + $service + ":sha-" + $commit),
    digest: ("sha256:" + ($digit * 64))
  };
  {
    schema_version: 1,
    source_commit: $commit,
    platform: "linux/amd64",
    images: {
      backend: image("backend"; "1"),
      marketplace: image("marketplace"; "2"),
      "marketplace-web": image("marketplace-web"; "3"),
      relay: image("relay"; "4"),
      web: image("web"; "5")
    }
  }
' > "${MANIFEST}"

# shellcheck disable=SC1091
source "${ROOT}/promote-staged-images.sh"
mkdir -p "${TEST_REPO}/deploy/kubernetes/cluster-oilan/release"
cp "${ROOT}/release/kustomization.yaml" \
  "${TEST_REPO}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml"
# shellcheck disable=SC2034
REPO_ROOT="${TEST_REPO}"

release_require_pushed_clean_tree() {
  RELEASE_SOURCE_COMMIT="${GUARD_COMMIT}"
  export RELEASE_SOURCE_COMMIT
}
ensure_project() { printf 'project\n' >> "${LOG}"; }
harbor_require_upload_token_expiration() { printf 'token %s %s\n' "$1" "$2" >> "${LOG}"; }
promote_platform_manifest() {
  local destination="$1" source="$2" service expected
  service="${destination#"${PROJ}"/}"
  service="${service%%:*}"
  expected="$(platform_manifest_digest "docker.io/l8ai/doworker-oilan-${service}:sha-${COMMIT}")"
  [[ "${source}" == "docker.io/l8ai/doworker-oilan-${service}:sha-${COMMIT}@${expected}" ]]
  [[ " ${SEEDED_SERVICES[*]} " == *" ${service} "* ]]
  PROMOTED_SERVICES+=("${service}")
  printf 'promote %s %s\n' "$1" "$2" >> "${LOG}"
}
platform_manifest_digest() {
  local image="$1" service digit
  if [[ -n "${DIGEST_OVERRIDE}" && "${image}" == docker.io/* ]]; then
    printf '%s' "${DIGEST_OVERRIDE}"
    return
  fi
  service="${image##*oilan-}"
  service="${service%%:*}"
  if [[ "${image}" == "${PROJ}/"* ]]; then
    service="${image#"${PROJ}"/}"
    service="${service%%:*}"
    [[ " ${PROMOTED_SERVICES[*]} " == *" ${service} "* ]] || return 1
  fi
  case "${service}" in
    backend) digit=1 ;;
    marketplace) digit=2 ;;
    marketplace-web) digit=3 ;;
    relay) digit=4 ;;
    web) digit=5 ;;
    *) return 1 ;;
  esac
  printf 'sha256:'
  printf '%*s' 64 '' | tr ' ' "${digit}"
}
docker() {
  if [[ "$1 $2" == "image inspect" && "$*" == *".Os"* ]]; then
    printf 'linux/amd64\n'
  elif [[ "$1 $2" == "image inspect" ]]; then
    printf '%s\n' "${REVISION_OVERRIDE:-${COMMIT}}"
  elif [[ "$1" == "tag" || "$1" == "push" ]]; then
    [[ "$1" != "push" || "${PUSH_FAILURE}" == false ]] || return 1
    if [[ "$1" == "push" ]]; then
      local service="${2#"${PROJ}"/}"
      SEEDED_SERVICES+=("${service%%:*}")
    fi
    printf '%s\n' "$*" >> "${LOG}"
  elif [[ "$1" != "pull" ]]; then
    return 1
  fi
}
release_write_source_metadata() { printf 'source-metadata\n' >> "${LOG}"; }
expect_failure() {
  local name="$1" manifest="$2"
  if main "${manifest}" >/dev/null 2>&1; then
    echo "${name} was accepted" >&2
    exit 1
  fi
}

unchanged_release_digests() {
  awk -v project="${PROJ}" '
    $1 == "-" && $2 == "name:" {
      name=$3
      keep=name == project "/web-admin" || name == project "/mobile" ||
        name == project "/pgvector" || name == project "/redis" ||
        name == project "/minio" || name == project "/mc" ||
        name == project "/kubectl"
      next
    }
    keep && $1 == "digest:" {
      print name, $2
      keep=0
    }
  ' "$1"
}

RELEASE_LOCK="${TEST_REPO}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml"
unchanged_before="$(unchanged_release_digests "${RELEASE_LOCK}")"
main "${MANIFEST}"

grep -Fxq 'project' "${LOG}"
grep -Fxq 'token repo.aiedulab.cn:8443 120' "${LOG}"
[[ "$(grep -c '^tag ' "${LOG}")" == "5" ]]
[[ "$(grep -c '^push ' "${LOG}")" == "5" ]]
[[ "$(grep -c '^promote ' "${LOG}")" == "5" ]]
grep -Fq "tag docker.io/l8ai/doworker-oilan-backend:sha-${COMMIT}@sha256:" "${LOG}"
grep -Fxq 'source-metadata' "${LOG}"
grep -A1 -Fq "name: ${PROJ}/backend" "${RELEASE_LOCK}"
grep -Fq 'digest: sha256:1111111111111111111111111111111111111111111111111111111111111111' \
  "${RELEASE_LOCK}"
[[ "$(unchanged_release_digests "${RELEASE_LOCK}")" == "${unchanged_before}" ]]

jq '.images.web.source = "docker.io/unapproved/web:latest"' "${MANIFEST}" > "${TMP}/invalid.json"
before_invalid="$(wc -l < "${LOG}")"
expect_failure "unexpected staging source" "${TMP}/invalid.json"
[[ "$(wc -l < "${LOG}")" == "${before_invalid}" ]]

jq '.platform = "linux/arm64"' "${MANIFEST}" > "${TMP}/platform.json"
expect_failure "unexpected staging platform" "${TMP}/platform.json"

GUARD_COMMIT="bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
expect_failure "staging commit mismatch" "${MANIFEST}"
GUARD_COMMIT="${COMMIT}"

DIGEST_OVERRIDE="sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
expect_failure "staging digest mismatch" "${MANIFEST}"
DIGEST_OVERRIDE=""

REVISION_OVERRIDE="bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
expect_failure "staging revision mismatch" "${MANIFEST}"
REVISION_OVERRIDE=""

PUSH_FAILURE=true
expect_failure "Harbor push failure" "${MANIFEST}"
PUSH_FAILURE=false

grep -Fq 'manifest:' "${ROOT}/../../../.github/workflows/oilan-image-publish.yml"
grep -Fq 'name: oilan-staging-release' "${ROOT}/../../../.github/workflows/oilan-image-publish.yml"

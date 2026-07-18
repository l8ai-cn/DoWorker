#!/usr/bin/env bash
set -euo pipefail

ROOT="$(mktemp -d)"
trap 'rm -rf "${ROOT}"' EXIT
REPO_ROOT="${ROOT}/repo"
REG="registry.example"
PROJ="${REG}/agentsmesh"
LOCKED="sha256:1111111111111111111111111111111111111111111111111111111111111111"
COPIED="sha256:2222222222222222222222222222222222222222222222222222222222222222"
REMOTE_DIGEST="${LOCKED}"
HARBOR_STATUS=200
INDEX_MODE="multi"
COPIED_INDEX_MODE="multi"
INSPECT_ERROR=0
HARBOR_QUERY_ERROR=0
MANIFEST_QUERY_ERROR=0
CREATE_MARKER="${ROOT}/create"
CREATE_LOG="${ROOT}/create.log"
mkdir -p "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/release"

write_release_lock() {
  cat > "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml" <<EOF
images:
  - name: ${PROJ}/pgvector
    digest: ${LOCKED}
  - name: ${PROJ}/redis
    digest: ${COPIED}
EOF
}

infra_manifest_digest() {
  [[ "${MANIFEST_QUERY_ERROR}" -eq 0 ]] || return 1
  printf '%s' "${REMOTE_DIGEST}"
}

docker() {
  if [[ "$*" == "buildx imagetools inspect ${PROJ}/pgvector@${REMOTE_DIGEST} --raw" ]]; then
    [[ "${INSPECT_ERROR}" -eq 0 ]] || return 1
    if [[ "${INDEX_MODE}" == "multi" ]]; then
      printf '%s\n' '{"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[{"platform":{"os":"linux","architecture":"amd64"}},{"platform":{"os":"linux","architecture":"arm64"}}]}'
    elif [[ "${INDEX_MODE}" == "single" ]]; then
      printf '%s\n' '{"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[{"platform":{"os":"linux","architecture":"amd64"}}]}'
    else
      printf '%s\n' 'not-json'
    fi
    return
  fi
  if [[ "$*" == "buildx imagetools create --tag ${PROJ}/pgvector:pg16 pgvector/pgvector:pg16" ]]; then
    printf '%s\n' "$*" >> "${CREATE_LOG}"
    touch "${CREATE_MARKER}"
    REMOTE_DIGEST="${COPIED}"
    INDEX_MODE="${COPIED_INDEX_MODE}"
    return
  fi
  return 1
}

source "$(dirname "$0")/harbor-infra-mirror.sh"

harbor_artifact_http_status() {
  [[ "${HARBOR_QUERY_ERROR}" -eq 0 ]] || return 1
  printf '%s' "${HARBOR_STATUS}"
}

reset_case() {
  rm -f "${CREATE_MARKER}" "${CREATE_LOG}"
  REMOTE_DIGEST="${LOCKED}"
  HARBOR_STATUS=200
  INDEX_MODE="multi"
  COPIED_INDEX_MODE="multi"
  INSPECT_ERROR=0
  HARBOR_QUERY_ERROR=0
  MANIFEST_QUERY_ERROR=0
  PLATFORM_DIGEST_PGVECTOR=""
  write_release_lock
}

reset_case
mirror pgvector/pgvector:pg16 pgvector:pg16
[[ "${PLATFORM_DIGEST_PGVECTOR}" == "${LOCKED}" ]]
[[ ! -e "${CREATE_MARKER}" ]]

reset_case
REMOTE_DIGEST="${COPIED}"
mirror pgvector/pgvector:pg16 pgvector:pg16
[[ "${PLATFORM_DIGEST_PGVECTOR}" == "${COPIED}" ]]
[[ -e "${CREATE_MARKER}" ]]

reset_case
INDEX_MODE="single"
mirror pgvector/pgvector:pg16 pgvector:pg16
[[ "${PLATFORM_DIGEST_PGVECTOR}" == "${COPIED}" ]]
[[ -e "${CREATE_MARKER}" ]]

reset_case
COPIED_INDEX_MODE="single"
INDEX_MODE="single"
if mirror pgvector/pgvector:pg16 pgvector:pg16; then
  echo "single-architecture source mirror must be rejected" >&2
  exit 1
fi
[[ -e "${CREATE_MARKER}" ]]

reset_case
INSPECT_ERROR=1
if mirror pgvector/pgvector:pg16 pgvector:pg16; then
  echo "Harbor inspection errors must block publication" >&2
  exit 1
fi
[[ ! -e "${CREATE_MARKER}" ]]

reset_case
INDEX_MODE="invalid"
if mirror pgvector/pgvector:pg16 pgvector:pg16; then
  echo "invalid manifest JSON must block publication" >&2
  exit 1
fi
[[ ! -e "${CREATE_MARKER}" ]]

reset_case
HARBOR_STATUS=500
if mirror pgvector/pgvector:pg16 pgvector:pg16; then
  echo "Harbor API errors must block publication" >&2
  exit 1
fi
[[ ! -e "${CREATE_MARKER}" ]]

reset_case
HARBOR_QUERY_ERROR=1
if mirror pgvector/pgvector:pg16 pgvector:pg16; then
  echo "Harbor query failures must block publication" >&2
  exit 1
fi
[[ ! -e "${CREATE_MARKER}" ]]

reset_case
MANIFEST_QUERY_ERROR=1
if mirror pgvector/pgvector:pg16 pgvector:pg16; then
  echo "manifest digest query failures must block publication" >&2
  exit 1
fi
[[ ! -e "${CREATE_MARKER}" ]]

reset_case
HARBOR_STATUS=404
mirror pgvector/pgvector:pg16 pgvector:pg16
[[ "${PLATFORM_DIGEST_PGVECTOR}" == "${COPIED}" ]]
[[ -e "${CREATE_MARKER}" ]]
grep -Fq 'pgvector/pgvector:pg16' "${CREATE_LOG}"

cat > "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml" <<EOF
images:
  - name: ${PROJ}/pgvector
  - name: ${PROJ}/redis
    digest: ${COPIED}
EOF
set +e
release_locked_infra_digest pgvector >/dev/null
status=$?
set -e
[[ "${status}" -eq "${INFRA_VERIFICATION_ERROR}" ]]

cat > "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml" <<EOF
images:
  - name: ${PROJ}/redis
    digest: ${COPIED}
EOF
rm -f "${CREATE_MARKER}"
set +e
mirror pgvector/pgvector:pg16 pgvector:pg16
status=$?
set -e
[[ "${status}" -ne 0 ]]
[[ ! -e "${CREATE_MARKER}" ]]

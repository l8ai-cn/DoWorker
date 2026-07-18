#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT
REPO_ROOT="${TMP}/repo"
DIGEST="sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
REFERENCE="repo.aiedulab.cn:8443/library/gitea@${DIGEST}"
DOCKER_LOG="${TMP}/docker.log"
PLATFORM="linux/amd64"
PULL_FAIL=false
VERSION="Gitea version 1.21.0 built with GNU Make"
USER_HELP="generate-access-token"

mkdir -p "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/release"
write_lock() {
  printf 'images:\n  - name: repo.aiedulab.cn:8443/library/gitea\n    digest: %s\n' \
    "${DIGEST}" > "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml"
}

docker() {
  printf '%s\n' "$*" >> "${DOCKER_LOG}"
  case "$1 $2" in
    "pull ${REFERENCE}")
      [[ "${PULL_FAIL}" == false ]]
      ;;
    "image inspect")
      printf '%s\n' "${PLATFORM}"
      ;;
    "run --rm")
      if [[ "$*" == *" --version" ]]; then
        printf '%s\n' "${VERSION}"
      elif [[ "$*" == *" admin user --help" ]]; then
        printf '%s\n' "${USER_HELP}"
      else
        echo "unexpected docker run command: $*" >&2
        return 1
      fi
      ;;
    *)
      echo "unexpected docker command: $*" >&2
      return 1
      ;;
  esac
}

source "${ROOT}/release_image_provenance.sh"

write_lock
release_verify_gitea_provenance "${REPO_ROOT}"
grep -Fx "pull ${REFERENCE}" "${DOCKER_LOG}" >/dev/null
! grep -F 'gitea:1.21.0' "${DOCKER_LOG}" >/dev/null

PLATFORM="linux/arm64"
if release_verify_gitea_provenance "${REPO_ROOT}" 2>/dev/null; then
  echo "wrong Gitea image platform was accepted" >&2
  exit 1
fi

PLATFORM="linux/amd64"
PULL_FAIL=true
if release_verify_gitea_provenance "${REPO_ROOT}" 2>/dev/null; then
  echo "failed Gitea image pull was accepted" >&2
  exit 1
fi

PULL_FAIL=false
VERSION="Gitea version 1.22.0 built with GNU Make"
if release_verify_gitea_provenance "${REPO_ROOT}" 2>/dev/null; then
  echo "wrong Gitea version was accepted" >&2
  exit 1
fi

VERSION="Gitea version 1.21.0 built with GNU Make"
USER_HELP="list"
if release_verify_gitea_provenance "${REPO_ROOT}" 2>/dev/null; then
  echo "Gitea image without token generation was accepted" >&2
  exit 1
fi

USER_HELP="generate-access-token"
printf 'images: []\n' > "${REPO_ROOT}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml"
if release_verify_gitea_provenance "${REPO_ROOT}" 2>/dev/null; then
  echo "missing Gitea release lock was accepted" >&2
  exit 1
fi

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT
DIGEST="sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
REGISTRY="repo.aiedulab.cn:8443/agentcloud"
GITEA_REGISTRY="repo.aiedulab.cn:8443/library"
GITEA_DIGEST="sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
LOCK="${TMP}/release-kustomization.yaml"
VALID="${TMP}/valid.yaml"
printf 'images:\n  - name: %s/gitea\n    digest: %s\n' \
  "${GITEA_REGISTRY}" "${GITEA_DIGEST}" > "${LOCK}"

{
  for image in backend marketplace marketplace-web relay web web-admin; do
    printf '        image: %s/%s@%s\n' "${REGISTRY}" "${image}" "${DIGEST}"
  done
  printf '      - image: %s/mobile@%s\n' "${REGISTRY}" "${DIGEST}"
  for image in pgvector redis minio kubectl; do
    printf '        image: %s/%s@%s\n' "${REGISTRY}" "${image}" "${DIGEST}"
  done
  printf '        image: %s/gitea@%s\n' "${GITEA_REGISTRY}" "${GITEA_DIGEST}"
  printf '    agentcloud.ai/verified-image-digest: "%s"\n' "${DIGEST}"
} > "${VALID}"

VERIFY_RELEASE_LOCK="${LOCK}" bash "${ROOT}/verify_release_images.sh" "${VALID}"

sed "s|${REGISTRY}/mobile@${DIGEST}|${REGISTRY}/mobile:latest|" \
  "${VALID}" > "${TMP}/mutable.yaml"
if VERIFY_RELEASE_LOCK="${LOCK}" \
  bash "${ROOT}/verify_release_images.sh" "${TMP}/mutable.yaml" 2>/dev/null; then
  echo "mutable mobile image was accepted" >&2
  exit 1
fi

sed "s|${GITEA_REGISTRY}/gitea@${GITEA_DIGEST}|${GITEA_REGISTRY}/gitea:1.21.0|" \
  "${VALID}" > "${TMP}/mutable-gitea.yaml"
if VERIFY_RELEASE_LOCK="${LOCK}" \
  bash "${ROOT}/verify_release_images.sh" "${TMP}/mutable-gitea.yaml" 2>/dev/null; then
  echo "mutable Gitea image was accepted" >&2
  exit 1
fi

sed "s|${GITEA_REGISTRY}/gitea@${GITEA_DIGEST}|${GITEA_REGISTRY}/gitea@${DIGEST}|" \
  "${VALID}" > "${TMP}/unapproved-gitea.yaml"
if VERIFY_RELEASE_LOCK="${LOCK}" \
  bash "${ROOT}/verify_release_images.sh" "${TMP}/unapproved-gitea.yaml" 2>/dev/null; then
  echo "unapproved Gitea digest was accepted" >&2
  exit 1
fi

sed 's/sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"/sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"/' \
  "${VALID}" > "${TMP}/drift.yaml"
if VERIFY_RELEASE_LOCK="${LOCK}" \
  bash "${ROOT}/verify_release_images.sh" "${TMP}/drift.yaml" 2>/dev/null; then
  echo "backend digest annotation drift was accepted" >&2
  exit 1
fi

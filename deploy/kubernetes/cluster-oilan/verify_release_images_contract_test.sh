#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT
DIGEST="sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
REGISTRY="repo.aiedulab.cn:8443/agentsmesh"
VALID="${TMP}/valid.yaml"

{
  for image in backend marketplace marketplace-web relay web web-admin; do
    printf '        image: %s/%s@%s\n' "${REGISTRY}" "${image}" "${DIGEST}"
  done
  printf '      - image: %s/mobile@%s\n' "${REGISTRY}" "${DIGEST}"
  for image in pgvector redis minio kubectl; do
    printf '        image: %s/%s@%s\n' "${REGISTRY}" "${image}" "${DIGEST}"
  done
  printf '    agentsmesh.ai/verified-image-digest: "%s"\n' "${DIGEST}"
} > "${VALID}"

bash "${ROOT}/verify_release_images.sh" "${VALID}"

sed "s|${REGISTRY}/mobile@${DIGEST}|${REGISTRY}/mobile:latest|" \
  "${VALID}" > "${TMP}/mutable.yaml"
if bash "${ROOT}/verify_release_images.sh" "${TMP}/mutable.yaml" 2>/dev/null; then
  echo "mutable mobile image was accepted" >&2
  exit 1
fi

sed 's/sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"/sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"/' \
  "${VALID}" > "${TMP}/drift.yaml"
if bash "${ROOT}/verify_release_images.sh" "${TMP}/drift.yaml" 2>/dev/null; then
  echo "backend digest annotation drift was accepted" >&2
  exit 1
fi

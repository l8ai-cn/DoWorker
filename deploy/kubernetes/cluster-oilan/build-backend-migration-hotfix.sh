#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${DIR}/../../.." && pwd)"
CONTEXT="${DIR}/_gen/backend-migration-hotfix"
BASE="repo.aiedulab.cn:8443/agentsmesh/backend@sha256:c160059c60e2b9d6b8d99b4919ca90ec6fae570d69787c49fce9a9aa498e7a42"
IMAGE="repo.aiedulab.cn:8443/agentsmesh/backend:migration-222-hotfix"
LATEST_IMAGE="repo.aiedulab.cn:8443/agentsmesh/backend:latest"
EXPECTED_GO_VERSION="go1.26.2"
EXPECTED_SERVER_SHA="ad4f5c8df61f08b98fee8a59732b3e10b73011f1f768acec047b46ca94f76d91"
EXPECTED_IMAGE_DIGEST="sha256:fa58ff8756f5052ee48026f6fd20500e49ac0b464e655c24d34fc23fdba972e6"

test "$(go env GOVERSION)" = "${EXPECTED_GO_VERSION}"
mkdir -p "${CONTEXT}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath -ldflags="-s -w" \
  -o "${CONTEXT}/server" "${ROOT}/backend/cmd/server"
test "$(shasum -a 256 "${CONTEXT}/server" | awk '{print $1}')" = \
  "${EXPECTED_SERVER_SHA}"
touch -t 202607161710.00 "${CONTEXT}/server"

docker pull --platform linux/amd64 "${BASE}"
docker build --provenance=false --platform linux/amd64 \
  -f "${DIR}/backend-migration-hotfix.Dockerfile" \
  -t "${IMAGE}" "${CONTEXT}"
docker push "${IMAGE}"

digest="$(
  docker buildx imagetools inspect "${IMAGE}" \
    --format '{{.Manifest.Digest}}'
)"
test "${digest}" = "${EXPECTED_IMAGE_DIGEST}"
docker tag "${IMAGE}" "${LATEST_IMAGE}"
docker push "${LATEST_IMAGE}"
test "$(
  docker buildx imagetools inspect "${LATEST_IMAGE}" \
    --format '{{.Manifest.Digest}}'
)" = "${EXPECTED_IMAGE_DIGEST}"
printf '%s\n' "${digest}"

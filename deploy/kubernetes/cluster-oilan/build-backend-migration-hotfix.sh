#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${DIR}/../../.." && pwd)"
CONTEXT="${DIR}/_gen/backend-migration-hotfix"
OCI_ARCHIVE="${CONTEXT}/backend.oci.tar"
OCI_LAYOUT="${CONTEXT}/backend-oci-layout"
BASE="repo.aiedulab.cn:8443/agentsmesh/backend@sha256:c160059c60e2b9d6b8d99b4919ca90ec6fae570d69787c49fce9a9aa498e7a42"
IMAGE="repo.aiedulab.cn:8443/agentsmesh/backend:migration-222-hotfix"
LATEST_IMAGE="repo.aiedulab.cn:8443/agentsmesh/backend:latest"
EXPECTED_GO_VERSION="go1.26.2"
EXPECTED_SERVER_SHA="3119a109efff9b7d7eab31976e7f5fa47261bd631780d417de7653b77457e472"
EXPECTED_IMAGE_DIGEST="sha256:22c384c72ee54fa6a2877b9b2f6eb464ad5ba16be7efa3d1474838a55e18bde7"
SOURCE_DATE_EPOCH="1784193000"
SERVER="${CONTEXT}/server-${EXPECTED_SERVER_SHA}"

test "$(go env GOVERSION)" = "${EXPECTED_GO_VERSION}"
mkdir -p "${CONTEXT}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath -ldflags="-s -w" \
  -o "${SERVER}" "${ROOT}/backend/cmd/server"
test "$(shasum -a 256 "${SERVER}" | awk '{print $1}')" = \
  "${EXPECTED_SERVER_SHA}"
touch -t 202607161710.00 "${SERVER}"

docker pull --platform linux/amd64 "${BASE}"
rm -f "${OCI_ARCHIVE}"
rm -rf "${OCI_LAYOUT}"
docker buildx build --no-cache --provenance=false --platform linux/amd64 \
  --build-arg SERVER_SHA="${EXPECTED_SERVER_SHA}" \
  --build-arg SOURCE_DATE_EPOCH="${SOURCE_DATE_EPOCH}" \
  -f "${DIR}/backend-migration-hotfix.Dockerfile" \
  --output "type=oci,dest=${OCI_ARCHIVE},name=${IMAGE},rewrite-timestamp=true" \
  "${CONTEXT}"
mkdir -p "${OCI_LAYOUT}"
tar -xf "${OCI_ARCHIVE}" -C "${OCI_LAYOUT}"
test "$(jq -r '.manifests[0].digest' "${OCI_LAYOUT}/index.json")" = \
  "${EXPECTED_IMAGE_DIGEST}"
docker load -i "${OCI_ARCHIVE}"
container_id="$(docker create --platform linux/amd64 "${IMAGE}")"
trap 'docker rm "${container_id}" >/dev/null 2>&1 || true' EXIT
docker cp "${container_id}:/app/server" "${CONTEXT}/image-server"
test "$(shasum -a 256 "${CONTEXT}/image-server" | awk '{print $1}')" = \
  "${EXPECTED_SERVER_SHA}"
docker rm "${container_id}" >/dev/null
trap - EXIT
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

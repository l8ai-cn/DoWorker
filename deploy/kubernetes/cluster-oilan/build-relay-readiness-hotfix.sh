#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${DIR}/../../.." && pwd)"
CONTEXT="${DIR}/_gen/relay-readiness-hotfix"
OCI_ARCHIVE="${CONTEXT}/relay.oci.tar"
OCI_LAYOUT="${CONTEXT}/relay-oci-layout"
IMAGE="repo.aiedulab.cn:8443/agentsmesh/relay:readiness-hotfix"
LATEST_IMAGE="repo.aiedulab.cn:8443/agentsmesh/relay:latest"
EXPECTED_GO_VERSION="go1.26.2"
EXPECTED_RELAY_SHA="7a95dd4a3b235a41d7c201482732c0ac3244922586febca4ce4a2732eeb32948"
EXPECTED_IMAGE_DIGEST="sha256:7ce51042743fa58f03ec045e22c211a4d7f830b215491f50e63807d332ab80e8"
SOURCE_DATE_EPOCH="1784232000"
RELAY="${CONTEXT}/relay-${EXPECTED_RELAY_SHA}"

test "$(go env GOVERSION)" = "${EXPECTED_GO_VERSION}"
mkdir -p "${CONTEXT}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath -ldflags="-s -w" \
  -o "${RELAY}" "${ROOT}/relay/cmd/relay"
test "$(shasum -a 256 "${RELAY}" | awk '{print $1}')" = \
  "${EXPECTED_RELAY_SHA}"
touch -t 202607170400.00 "${RELAY}"

rm -f "${OCI_ARCHIVE}"
rm -rf "${OCI_LAYOUT}"
docker buildx build --no-cache --provenance=false --platform linux/amd64 \
  --build-arg RELAY_SHA="${EXPECTED_RELAY_SHA}" \
  --build-arg SOURCE_DATE_EPOCH="${SOURCE_DATE_EPOCH}" \
  -f "${DIR}/relay-readiness-hotfix.Dockerfile" \
  --output "type=oci,dest=${OCI_ARCHIVE},name=${IMAGE},rewrite-timestamp=true" \
  "${CONTEXT}"
mkdir -p "${OCI_LAYOUT}"
tar -xf "${OCI_ARCHIVE}" -C "${OCI_LAYOUT}"
test "$(jq -r '.manifests[0].digest' "${OCI_LAYOUT}/index.json")" = \
  "${EXPECTED_IMAGE_DIGEST}"
docker load -i "${OCI_ARCHIVE}"

container_id="$(docker create --platform linux/amd64 "${IMAGE}")"
trap 'docker rm "${container_id}" >/dev/null 2>&1 || true' EXIT
docker cp "${container_id}:/app/relay" "${CONTEXT}/image-relay"
test "$(shasum -a 256 "${CONTEXT}/image-relay" | awk '{print $1}')" = \
  "${EXPECTED_RELAY_SHA}"
grep -aFq "publisher_ready" "${CONTEXT}/image-relay"
docker rm "${container_id}" >/dev/null
trap - EXIT

docker push "${IMAGE}"
test "$(
  docker buildx imagetools inspect "${IMAGE}" \
    --format '{{.Manifest.Digest}}'
)" = "${EXPECTED_IMAGE_DIGEST}"
docker tag "${IMAGE}" "${LATEST_IMAGE}"
docker push "${LATEST_IMAGE}"
test "$(
  docker buildx imagetools inspect "${LATEST_IMAGE}" \
    --format '{{.Manifest.Digest}}'
)" = "${EXPECTED_IMAGE_DIGEST}"
printf '%s\n' "${EXPECTED_IMAGE_DIGEST}"

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

mkdir -p \
  "${TMP}/backend/internal/domain/workerruntime" \
  "${TMP}/deploy/kubernetes/cluster-oilan" \
  "${TMP}/docker/agent-runtime" \
  "${TMP}/tools/loops/worker-onboarding/catalog-loop/evidence/runtime-builds"
cp "${ROOT}/backend/internal/domain/workerruntime/runtime_catalog.lock.json" \
  "${TMP}/backend/internal/domain/workerruntime/"
cp "${ROOT}/deploy/kubernetes/cluster-oilan/30-backend.yaml" \
  "${TMP}/deploy/kubernetes/cluster-oilan/"
cp "${ROOT}/docker/agent-runtime/do-agent-release.json" \
  "${TMP}/docker/agent-runtime/"
cp "${ROOT}/tools/loops/worker-onboarding/catalog-loop/evidence/runtime-builds/"{do-agent,seedance-expert}.json \
  "${TMP}/tools/loops/worker-onboarding/catalog-loop/evidence/runtime-builds/"

DIGEST="sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
COMMIT="bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
TAG="4e17ae3f40f1-runner-bbbbbbbbbbbb"
RUNTIME_OBSERVED_AT="2026-07-16T15:55:00+08:00" \
  node "${ROOT}/deploy/kubernetes/cluster-oilan/update-do-agent-runtime-digest.mjs" \
    "${DIGEST}" "${COMMIT}" "${TAG}" "${TMP}"

jq -e --arg digest "${DIGEST}" --arg tag "${TAG}" '
  .image.digest == $digest and .image.tag == $tag
' "${TMP}/docker/agent-runtime/do-agent-release.json" >/dev/null
jq -e --arg digest "${DIGEST}" '
  .images[] | select(.slug == "do-agent-stable") |
  .digest == $digest and (.reference | endswith("@" + $digest))
' "${TMP}/backend/internal/domain/workerruntime/runtime_catalog.lock.json" >/dev/null
[[ "$(grep -o "${DIGEST}" "${TMP}/deploy/kubernetes/cluster-oilan/30-backend.yaml" | wc -l | tr -d ' ')" == "2" ]]
for worker in do-agent seedance-expert; do
  jq -e --arg digest "${DIGEST}" '
    .image_id == $digest and (.image | endswith("@" + $digest))
  ' "${TMP}/tools/loops/worker-onboarding/catalog-loop/evidence/runtime-builds/${worker}.json" >/dev/null
done

before="$(find "${TMP}" -type f -print0 | sort -z | xargs -0 shasum -a 256)"
if RUNTIME_OBSERVED_AT="2026-07-16T15:55:00+08:00" \
  node "${ROOT}/deploy/kubernetes/cluster-oilan/update-do-agent-runtime-digest.mjs" \
    invalid "${COMMIT}" "${TAG}" "${TMP}" >/dev/null 2>&1; then
  echo "invalid do-agent digest was accepted" >&2
  exit 1
fi
[[ "${before}" == "$(find "${TMP}" -type f -print0 | sort -z | xargs -0 shasum -a 256)" ]]

#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD="${DIR}/build-relay-readiness-hotfix.sh"
DOCKERFILE="${DIR}/relay-readiness-hotfix.Dockerfile"
RELEASE="${DIR}/release/kustomization.yaml"

grep -Fq 'EXPECTED_GO_VERSION="go1.26.2"' "${BUILD}"
grep -Fq \
  'EXPECTED_RELAY_SHA="7a95dd4a3b235a41d7c201482732c0ac3244922586febca4ce4a2732eeb32948"' \
  "${BUILD}"
grep -Fq \
  'EXPECTED_IMAGE_DIGEST="sha256:7ce51042743fa58f03ec045e22c211a4d7f830b215491f50e63807d332ab80e8"' \
  "${BUILD}"
grep -Fq \
  'FROM repo.aiedulab.cn:8443/agentsmesh/relay@sha256:4e5992c1702cfc467d578ae4ff693cdece606c534c62c1c25dbd373498d4022d' \
  "${DOCKERFILE}"
grep -Fq 'COPY --chown=1000:1000 relay-${RELAY_SHA} /app/relay' \
  "${DOCKERFILE}"
grep -Fq 'grep -aFq "publisher_ready" "${CONTEXT}/image-relay"' \
  "${BUILD}"
grep -Fq \
  'digest: sha256:c9ba960b7fd6f9d6456dc52c549cbbaf7f939b9d599c8d6e71600995db177d93' \
  "${RELEASE}"

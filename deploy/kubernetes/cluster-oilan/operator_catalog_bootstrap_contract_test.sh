#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
JOB="${DIR}/26-operator-catalog-bootstrap-job.yaml"
DEPLOY="${DIR}/deploy.sh"

grep -Fq "name: operator-catalog-bootstrap" "${JOB}"
grep -Fq "image: __BACKEND_IMAGE__" "${JOB}"
grep -Fq -- "- bootstrap-marketplace" "${JOB}"
grep -Fq -- "- dev-org" "${JOB}"
grep -Fq -- "- admin@agentsmesh.local" "${JOB}"
grep -A1 -F -- "- --model-resource-id" "${JOB}" | grep -Fq -- '- "1"'
grep -A1 -F -- "- --runtime-image-id" "${JOB}" | grep -Fq -- '- "4"'
grep -Fq "mountPath: /app/access-token" "${JOB}"
grep -Fq "secretName: agentsmesh-access-token" "${JOB}"
grep -Fq "mountPath: /app/ssl" "${JOB}"
grep -Fq "secretName: agentsmesh-pki-ca" "${JOB}"
grep -Fq "mountPath: /data" "${JOB}"
grep -Fq "bootstrap_operator_catalog" "${DEPLOY}"
grep -Fq "job/operator-catalog-bootstrap" "${DEPLOY}"

if grep -Fq "26-operator-catalog-bootstrap-job.yaml" \
  "${DIR}/kustomization.yaml"; then
  echo "bootstrap job must remain out of long-running kustomize resources" >&2
  exit 1
fi

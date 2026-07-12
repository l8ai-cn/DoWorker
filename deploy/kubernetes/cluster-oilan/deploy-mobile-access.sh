#!/usr/bin/env bash
# Deploy only the services required by the mobile Worker access path. This
# intentionally avoids a full cluster reconcile and its unrelated workloads.
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="${DOOPS_TARGET:-gw-oilan-node}"
SESSION="${DOOPS_SESSION:-$(doops session | tr -d '[:space:]')}"
WORKSPACE="/root/ws/${SESSION}"
NAMESPACE=agentsmesh

require_pinned_images() {
  local manifest
  for manifest in 30-backend.yaml 31-relay.yaml 42-mobile.yaml; do
    grep -Eq 'image: .+@sha256:[a-f0-9]{64}$' "${DIR}/${manifest}" || {
      echo "immutable image digest required: ${manifest}" >&2
      exit 1
    }
  done
}

dexec() {
  doops -session "${SESSION}" exec --target "${TARGET}" --cmd "cd ${WORKSPACE} && $1"
}

backend_image() {
  awk '$1 == "image:" && $2 ~ /agentsmesh\/backend@sha256:/ { print $2; exit }' \
    "${DIR}/30-backend.yaml"
}

apply_backend_job() {
  local manifest="$1"
  local image="$2"
  local digest="${image##*@}"
  dexec "sed -E 's|repo.aiedulab.cn:8443/agentsmesh/backend@sha256:[a-f0-9]{64}|${image}|g; s|__BACKEND_IMAGE__|${image}|g; s|agentsmesh.ai/verified-image-digest: \"sha256:[a-f0-9]{64}\"|agentsmesh.ai/verified-image-digest: \"${digest}\"|g' ${manifest} | kubectl apply -f -"
}

require_pinned_images
BACKEND_IMAGE="$(backend_image)"
[[ -n "${BACKEND_IMAGE}" ]] || {
  echo "backend deployment must use an immutable image digest" >&2
  exit 1
}

echo "==> DoOps session ${SESSION} -> ${TARGET}"
doops -session "${SESSION}" push --target "${TARGET}" --src "${DIR}"
dexec "kubectl apply -f 02-configmap.yaml"
dexec "kubectl -n ${NAMESPACE} delete job migrate --ignore-not-found"
apply_backend_job "20-migrate-job.yaml" "${BACKEND_IMAGE}"
dexec "kubectl -n ${NAMESPACE} wait --for=condition=complete job/migrate --timeout=300s"
dexec "kubectl apply -f 30-backend.yaml"
dexec "kubectl -n ${NAMESPACE} rollout status deploy/backend --timeout=240s"
dexec "kubectl -n ${NAMESPACE} exec deploy/backend -- /app/worker-definition-sync"
dexec "kubectl apply -f 31-relay.yaml -f 42-mobile.yaml -f 43-mobile-ingress.yaml"

for deployment in relay mobile; do
  dexec "kubectl -n ${NAMESPACE} rollout status deploy/${deployment} --timeout=240s"
done

dexec "kubectl -n ${NAMESPACE} get ingress agentsmesh-mobile"
echo "==> deployed mobile access: https://mobile.l8ai.cn"

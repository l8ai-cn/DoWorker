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
  grep -A1 -F 'name: repo.aiedulab.cn:8443/agentsmesh/backend' \
    "${DIR}/release/kustomization.yaml" |
    grep -Eq 'digest: sha256:[a-f0-9]{64}$' || {
    echo "immutable backend image digest required" >&2
    exit 1
  }
}

dexec() {
  doops -session "${SESSION}" exec --target "${TARGET}" --cmd "cd ${WORKSPACE} && $1"
}

backend_image() {
  local digest
  digest="$(awk '
    $1 == "-" && $2 == "name:" && $3 ~ /agentsmesh\/backend$/ { found = 1; next }
    found && $1 == "digest:" { print $2; exit }
  ' "${DIR}/release/kustomization.yaml")"
  printf 'repo.aiedulab.cn:8443/agentsmesh/backend@%s' "${digest}"
}

apply_backend_job() {
  local manifest="$1"
  local image="$2"
  local digest="${image##*@}"
  dexec "sed -E 's|repo.aiedulab.cn:8443/agentsmesh/backend(:latest|@sha256:[a-f0-9]{64})|${image}|g; s|__BACKEND_IMAGE__|${image}|g; s|__BACKEND_DIGEST__|${digest}|g; s|agentsmesh.ai/verified-image-digest: \"sha256:[a-f0-9]{64}\"|agentsmesh.ai/verified-image-digest: \"${digest}\"|g' ${manifest} | kubectl apply -f -"
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
apply_backend_job "30-backend.yaml" "${BACKEND_IMAGE}"
dexec "kubectl -n ${NAMESPACE} rollout status deploy/backend --timeout=240s"
dexec "kubectl -n ${NAMESPACE} delete job worker-definition-sync --ignore-not-found"
apply_backend_job "23-worker-definition-sync-job.yaml" "${BACKEND_IMAGE}"
dexec "kubectl -n ${NAMESPACE} wait --for=condition=complete job/worker-definition-sync --timeout=300s"
dexec "kubectl apply -f 31-relay.yaml -f 42-mobile.yaml -f 43-mobile-ingress.yaml"

for deployment in relay mobile; do
  dexec "kubectl -n ${NAMESPACE} rollout status deploy/${deployment} --timeout=240s"
done

dexec "kubectl -n ${NAMESPACE} get ingress agentsmesh-mobile"
echo "==> deployed mobile access: https://mobile.l8ai.cn"

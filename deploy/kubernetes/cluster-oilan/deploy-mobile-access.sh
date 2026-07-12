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

require_pinned_images
echo "==> DoOps session ${SESSION} -> ${TARGET}"
doops -session "${SESSION}" push --target "${TARGET}" --src "${DIR}"
dexec "kubectl apply -f 02-configmap.yaml -f 30-backend.yaml -f 31-relay.yaml -f 42-mobile.yaml -f 43-mobile-ingress.yaml"

for deployment in backend relay mobile; do
  dexec "kubectl -n ${NAMESPACE} rollout status deploy/${deployment} --timeout=240s"
done

dexec "kubectl -n ${NAMESPACE} get ingress agentsmesh-mobile"
echo "==> deployed mobile access: https://mobile.l8ai.cn"

#!/usr/bin/env bash
# Deploy only the services required by the mobile Worker access path. This
# intentionally avoids a full cluster reconcile and its unrelated workloads.
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="${DOOPS_TARGET:-gw-oilan-node}"
SESSION="${DOOPS_SESSION:-$(doops session | tr -d '[:space:]')}"
WORKSPACE="/root/ws/${SESSION}"
NAMESPACE=agentsmesh

# shellcheck source=dosql_release_gate.sh
source "${DIR}/dosql_release_gate.sh"

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
  awk '$1 == "image:" && $2 ~ /agentsmesh\/backend@sha256:/ { print $2; exit }' \
    "${DIR}/30-backend.yaml"
}

main() {
  require_pinned_images
  BACKEND_IMAGE="$(backend_image)"
  [[ -n "${BACKEND_IMAGE}" ]] || {
    echo "backend deployment must use an immutable image digest" >&2
    return 1
  }

  echo "==> DoOps session ${SESSION} -> ${TARGET}"
  require_dosql_database_evidence
  doops -session "${SESSION}" push --target "${TARGET}" --src "${DIR}"
  dexec "kubectl apply -f 02-configmap.yaml"
  dexec "kubectl apply -f 30-backend.yaml"
  dexec "kubectl -n ${NAMESPACE} rollout status deploy/backend --timeout=240s"
  dexec "kubectl -n ${NAMESPACE} exec deploy/backend -- /app/worker-definition-sync"
  dexec "kubectl apply -f 31-relay.yaml -f 42-mobile.yaml -f 43-mobile-ingress.yaml"

  for deployment in relay mobile; do
    dexec "kubectl -n ${NAMESPACE} rollout status deploy/${deployment} --timeout=240s"
  done

  dexec "kubectl -n ${NAMESPACE} get ingress agentsmesh-mobile"
  echo "==> deployed mobile access: https://mobile.l8ai.cn"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi

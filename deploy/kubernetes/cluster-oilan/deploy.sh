#!/usr/bin/env bash
# Deploy AgentsMesh to the doops-oilan single-node k3s cluster via DoOps. All
# kubectl runs on the cluster node through `doops exec`; this host only generates
# secrets + pushes manifests.
#
#   ./deploy.sh            # full deploy (secrets + apply + jobs + rollout)
#
# Prereqs: images already pushed (./push-images.sh all), docker logged in to
# repo.aiedulab.cn:8443. DOOPS_SESSION may be preset; otherwise one is created.
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="${DOOPS_TARGET:-gw-oilan-node}"
SESSION="${DOOPS_SESSION:-$(doops session | tr -d '[:space:]')}"
WS="/root/ws/${SESSION}"
NS=agentsmesh
GEN="${DIR}/_gen"
SEC="${GEN}/secrets"
REG="repo.aiedulab.cn:8443"

echo "==> DoOps session ${SESSION} -> ${TARGET} (workspace ${WS})"

dexec()  { doops -session "${SESSION}" exec  --target "${TARGET}" --cmd "cd ${WS} && $1"; }
# `doops write` refuses paths outside the git snapshot, so secrets (generated,
# never committed) are streamed in base64 through exec and decoded on the node.
apply_secret() {
  local b64; b64="$(base64 < "$1" | tr -d '\n')"
  dexec "echo ${b64} | base64 -d | kubectl apply -f -"
}

# shellcheck source=cluster_secret_generation.sh
source "${DIR}/cluster_secret_generation.sh"

push_manifests() {
  echo "==> pushing manifests to ${TARGET}:${WS}"
  doops -session "${SESSION}" push --target "${TARGET}" --src "${DIR}"
}

ensure_tls_secret() {
  local tls="l8ai-wildcard-tls"
  if dexec "kubectl -n ${NS} get secret ${tls} -o name >/dev/null"; then
    echo "==> using existing ${NS}/${tls}"
  else
    dexec "kubectl get secret ${tls} -n default -o yaml | sed -e '/namespace:/d' -e '/resourceVersion:/d' -e '/uid:/d' -e '/creationTimestamp:/d' | kubectl apply -n ${NS} -f -"
  fi
  dexec "test \"\$(kubectl -n ${NS} get secret ${tls} -o jsonpath='{.type}')\" = kubernetes.io/tls"
}

apply_all() {
  echo "==> namespace + secrets"
  dexec "kubectl apply -f 00-namespace.yaml"
  for f in "${SEC}"/*.yaml; do apply_secret "${f}"; done
  echo "==> ensure wildcard TLS in ${NS}"
  ensure_tls_secret
  echo "==> apply workloads (kustomize)"
  dexec "kubectl kustomize . > /tmp/agentsmesh-release.yaml && bash verify_release_images.sh /tmp/agentsmesh-release.yaml && kubectl apply -f /tmp/agentsmesh-release.yaml"
  echo "==> wait for embedded migrations"
  dexec "kubectl -n ${NS} rollout status deploy/backend --timeout=300s"
  dexec "kubectl -n ${NS} rollout status deploy/marketplace --timeout=300s"
  dexec "kubectl -n ${NS} rollout status deploy/marketplace-web --timeout=300s"
  echo "==> seed + minio bucket"
  dexec "kubectl -n ${NS} delete job seed minio-setup --ignore-not-found"
  dexec "kubectl apply -f 21-seed-configmap.yaml -f 22-seed-job.yaml -f 13-minio-setup-job.yaml"
}

status() {
  echo "==> rollout status"
  for d in backend marketplace marketplace-web relay web web-admin mobile runner-e2e-echo; do
    dexec "kubectl -n ${NS} rollout status deploy/${d} --timeout=240s"
  done
  dexec "kubectl -n ${NS} get pods -o wide"
}

main() {
  generate_cluster_secrets
  push_manifests
  apply_all
  status
  echo "==> deployed. https://dowork.l8ai.cn · https://market.l8ai.cn · https://mobile.l8ai.cn (admin@agentsmesh.local / Ab123456)"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi

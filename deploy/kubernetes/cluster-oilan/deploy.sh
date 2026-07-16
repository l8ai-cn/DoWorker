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
RELEASE_BUNDLE=""
REMOTE_WORKSPACE_MAY_EXIST=false
SECRET_MANIFESTS=(
  agentsmesh-secrets.yaml
  agentsmesh-pki-ca.yaml
  agentsmesh-access-token.yaml
  agentsmesh-regcred.yaml
)

echo "==> DoOps session ${SESSION} -> ${TARGET} (workspace ${WS})"

dexec()  { doops -session "${SESSION}" exec  --target "${TARGET}" --cmd "cd ${WS} && $1"; }

cleanup_release_workspace() {
  local result=$?
  trap - EXIT
  [[ -z "${RELEASE_BUNDLE}" || ! -d "${RELEASE_BUNDLE}" ]] || rm -rf "${RELEASE_BUNDLE}"
  if [[ "${REMOTE_WORKSPACE_MAY_EXIST}" == true ]] &&
      ! doops -session "${SESSION}" clean --target "${TARGET}" --workspace "${SESSION}"; then
    echo "failed to clean DoOps release workspace ${SESSION}" >&2
    (( result != 0 )) || result=1
  fi
  exit "${result}"
}
trap cleanup_release_workspace EXIT

# shellcheck source=cluster_secret_generation.sh
source "${DIR}/cluster_secret_generation.sh"

push_manifests() {
  echo "==> pushing manifests to ${TARGET}:${WS}"
  RELEASE_BUNDLE="$(mktemp -d)"
  find "${DIR}" -mindepth 1 -maxdepth 1 ! -name _gen \
    -exec cp -R {} "${RELEASE_BUNDLE}/" \;
  mkdir -m 700 "${RELEASE_BUNDLE}/generated-secrets"
  for name in "${SECRET_MANIFESTS[@]}"; do
    test -f "${SEC}/${name}"
    cp "${SEC}/${name}" "${RELEASE_BUNDLE}/generated-secrets/"
  done
  chmod 600 "${RELEASE_BUNDLE}"/generated-secrets/*.yaml
  REMOTE_WORKSPACE_MAY_EXIST=true
  doops -session "${SESSION}" push --target "${TARGET}" --src "${RELEASE_BUNDLE}"
  rm -rf "${RELEASE_BUNDLE}"
  RELEASE_BUNDLE=""
}

ensure_tls_secret() {
  local tls="l8ai-wildcard-tls"
  if dexec "kubectl -n ${NS} get secret ${tls} -o name >/dev/null"; then
    echo "==> using existing ${NS}/${tls}"
  else
    dexec "kubectl get secret ${tls} -n default -o yaml | sed -e '/namespace:/d' -e '/resourceVersion:/d' -e '/uid:/d' -e '/creationTimestamp:/d' | kubectl apply -n ${NS} -f -"
  fi
  dexec "test \"\$(kubectl -n ${NS} get secret ${tls} -o jsonpath='{.type}')\" = kubernetes.io/tls"
  dexec "kubectl -n ${NS} get secret ${tls} -o jsonpath='{.data.tls\\.crt}' | base64 -d | openssl x509 -checkhost preview.l8ai.cn -noout"
  dexec "getent ahostsv4 preview.l8ai.cn >/dev/null"
}

sync_worker_definitions() {
  dexec "image=\$(awk '\$1 == \"image:\" && \$2 ~ /agentsmesh\\/backend@sha256:/ { print \$2; exit }' /tmp/agentsmesh-release.yaml); test -n \"\${image}\"; sed \"s|__BACKEND_IMAGE__|\${image}|g\" 23-worker-definition-sync-job.yaml | kubectl apply -f -"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/worker-definition-sync --timeout=300s"
}

bootstrap_operator_catalog() {
  dexec "image=\$(awk '\$1 == \"image:\" && \$2 ~ /agentsmesh\\/backend@sha256:/ { print \$2; exit }' /tmp/agentsmesh-release.yaml); test -n \"\${image}\"; sed \"s|__BACKEND_IMAGE__|\${image}|g\" 26-operator-catalog-bootstrap-job.yaml | kubectl apply -f -"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/operator-catalog-bootstrap --timeout=300s"
}

apply_all() {
  echo "==> namespace + secrets"
  dexec "kubectl apply -f 00-namespace.yaml"
  dexec "chmod 600 generated-secrets/*.yaml; status=0; cleanup_status=0; kubectl apply -f generated-secrets || status=\$?; rm -f generated-secrets/*.yaml || cleanup_status=\$?; rmdir generated-secrets || cleanup_status=\$?; test \${status} -ne 0 || status=\${cleanup_status}; exit \${status}"
  echo "==> ensure wildcard TLS in ${NS}"
  ensure_tls_secret
  echo "==> apply workloads (kustomize)"
  dexec "kubectl kustomize . > /tmp/agentsmesh-release.yaml && bash verify_release_images.sh /tmp/agentsmesh-release.yaml && kubectl apply -f /tmp/agentsmesh-release.yaml"
  echo "==> wait for embedded migrations"
  dexec "kubectl -n ${NS} rollout status deploy/backend --timeout=300s"
  dexec "kubectl -n ${NS} rollout status deploy/marketplace --timeout=300s"
  dexec "kubectl -n ${NS} rollout status deploy/marketplace-web --timeout=300s"
  echo "==> seed + minio bucket"
  dexec "kubectl -n ${NS} delete job seed minio-setup worker-definition-sync operator-catalog-bootstrap --ignore-not-found"
  dexec "kubectl apply -f 21-seed-configmap.yaml -f 22-seed-job.yaml -f 13-minio-setup-job.yaml"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/seed --timeout=300s"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/minio-setup --timeout=300s"
  sync_worker_definitions
  bootstrap_operator_catalog
}

status() {
  echo "==> rollout status"
  for d in backend marketplace marketplace-web relay web web-admin mobile runner-e2e-echo runner-video-studio; do
    dexec "kubectl -n ${NS} rollout status deploy/${d} --timeout=240s"
  done
  dexec "bash verify-video-runner-registration.sh"
  dexec "kubectl -n ${NS} get pods -o wide"
}

main() {
  generate_cluster_secrets
  push_manifests
  apply_all
  status
  echo "==> deployed. https://dowork.l8ai.cn · https://market.l8ai.cn · https://mobile.l8ai.cn · https://preview.l8ai.cn (admin@agentsmesh.local / Ab123456)"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi

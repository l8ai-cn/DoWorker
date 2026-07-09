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
TARGET="${DOOPS_TARGET:-gw-oilan}"
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

gen_secrets() {
  mkdir -p "${SEC}"
  [[ -f "${GEN}/ca.crt" ]] || {
    echo "==> generating runner mTLS CA"
    openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:prime256v1 -out "${GEN}/ca.key"
    openssl req -x509 -new -key "${GEN}/ca.key" -days 3650 -out "${GEN}/ca.crt" \
      -subj "/CN=AgentsMesh Runner CA/O=agentsmesh"
  }
  [[ -f "${GEN}/env" ]] || {
    echo "==> generating app secrets"
    {
      echo "DB_PASSWORD=$(openssl rand -hex 16)"
      echo "JWT_SECRET=$(openssl rand -hex 32)"
      echo "INTERNAL_API_SECRET=$(openssl rand -hex 24)"
      echo "MINIO_ROOT_PASSWORD=$(openssl rand -hex 16)"
    } > "${GEN}/env"
  }
  # shellcheck disable=SC1090
  source "${GEN}/env"

  kubectl create secret generic agentsmesh-secrets -n "${NS}" \
    --from-literal=DB_PASSWORD="${DB_PASSWORD}" \
    --from-literal=JWT_SECRET="${JWT_SECRET}" \
    --from-literal=INTERNAL_API_SECRET="${INTERNAL_API_SECRET}" \
    --from-literal=MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD}" \
    --from-literal=STORAGE_SECRET_KEY="${MINIO_ROOT_PASSWORD}" \
    --dry-run=client -o yaml > "${SEC}/agentsmesh-secrets.yaml"

  kubectl create secret generic agentsmesh-pki-ca -n "${NS}" \
    --from-file=ca.crt="${GEN}/ca.crt" --from-file=ca.key="${GEN}/ca.key" \
    --dry-run=client -o yaml > "${SEC}/agentsmesh-pki-ca.yaml"

  local store cred u p
  store="$(python3 -c "import json,os;print(json.load(open(os.path.expanduser('~/.docker/config.json'))).get('credsStore',''))")"
  cred="$(echo "${REG}" | "docker-credential-${store}" get)"
  u="$(echo "${cred}" | python3 -c "import sys,json;print(json.load(sys.stdin)['Username'])")"
  p="$(echo "${cred}" | python3 -c "import sys,json;print(json.load(sys.stdin)['Secret'])")"
  kubectl create secret docker-registry agentsmesh-regcred -n "${NS}" \
    --docker-server="${REG}" --docker-username="${u}" --docker-password="${p}" \
    --dry-run=client -o yaml > "${SEC}/agentsmesh-regcred.yaml"
}

push_manifests() {
  echo "==> pushing manifests to ${TARGET}:${WS}"
  doops -session "${SESSION}" push --target "${TARGET}" --src "${DIR}"
}

apply_all() {
  echo "==> namespace + secrets"
  dexec "kubectl apply -f 00-namespace.yaml"
  for f in "${SEC}"/*.yaml; do apply_secret "${f}"; done
  echo "==> copy wildcard TLS into ${NS}"
  for tls in l8ai-wildcard-tls l8an-wildcard-tls; do
    dexec "kubectl get secret ${tls} -n default -o yaml 2>/dev/null | sed -e '/namespace:/d' -e '/resourceVersion:/d' -e '/uid:/d' -e '/creationTimestamp:/d' | kubectl apply -n ${NS} -f -" || true
  done
  echo "==> apply workloads (kustomize)"
  dexec "kubectl apply -k ."
  echo "==> migrate (embedded), then seed + minio bucket"
  dexec "kubectl -n ${NS} delete job migrate seed minio-setup --ignore-not-found"
  dexec "kubectl apply -f 20-migrate-job.yaml"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/migrate --timeout=300s"
  dexec "kubectl apply -f 21-seed-configmap.yaml -f 22-seed-job.yaml -f 13-minio-setup-job.yaml"
}

status() {
  echo "==> rollout status"
  for d in backend relay web web-admin runner-e2e-echo; do
    dexec "kubectl -n ${NS} rollout status deploy/${d} --timeout=240s" || true
  done
  dexec "kubectl -n ${NS} get pods -o wide"
}

gen_secrets
push_manifests
apply_all
status
echo "==> deployed. https://dowork.l8ai.cn (admin@agentsmesh.local / Ab123456)"

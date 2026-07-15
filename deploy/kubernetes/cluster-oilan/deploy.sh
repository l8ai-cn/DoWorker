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

sync_worker_definitions() {
  local image
  image="$(backend_image)"
  [[ -n "${image}" ]] || {
    echo "backend deployment must use an immutable agentsmesh/backend digest" >&2
    return 1
  }
  apply_backend_job "23-worker-definition-sync-job.yaml" "${image}"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/worker-definition-sync --timeout=300s"
}

apply_all() {
  echo "==> namespace + secrets"
  dexec "kubectl apply -f 00-namespace.yaml"
  for f in "${SEC}"/*.yaml; do apply_secret "${f}"; done
  echo "==> copy wildcard TLS into ${NS}"
  for tls in l8ai-wildcard-tls l8an-wildcard-tls; do
    dexec "kubectl get secret ${tls} -n default -o yaml 2>/dev/null | sed -e '/namespace:/d' -e '/resourceVersion:/d' -e '/uid:/d' -e '/creationTimestamp:/d' | kubectl apply -n ${NS} -f -" || true
  done
  echo "==> apply database infrastructure"
  dexec "kubectl apply -f 02-configmap.yaml -f 10-postgres.yaml -f 11-redis.yaml -f 12-minio.yaml -f 30-backend-rbac.yaml"
  dexec "kubectl -n ${NS} rollout status statefulset/postgres --timeout=300s"
  dexec "kubectl -n ${NS} rollout status deployment/minio --timeout=300s"
  echo "==> migrate, seed, and sync Worker definitions"
  dexec "kubectl -n ${NS} delete job migrate seed minio-setup worker-definition-sync --ignore-not-found"
  local image
  image="$(backend_image)"
  [[ -n "${image}" ]] || {
    echo "backend deployment must use an immutable agentsmesh/backend digest" >&2
    return 1
  }
  apply_backend_job "20-migrate-job.yaml" "${image}"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/migrate --timeout=300s"
  dexec "kubectl apply -f 21-seed-configmap.yaml -f 22-seed-job.yaml"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/seed --timeout=300s"
  sync_worker_definitions
  dexec "kubectl apply -f 13-minio-setup-job.yaml"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/minio-setup --timeout=300s"
  echo "==> apply workloads (kustomize)"
  dexec "kubectl apply -k ."
}

status() {
  echo "==> rollout status"
  for d in backend relay web web-admin mobile runner-e2e-echo; do
    dexec "kubectl -n ${NS} rollout status deploy/${d} --timeout=240s"
  done
  dexec "kubectl -n ${NS} get pods -o wide"
}

gen_secrets
push_manifests
apply_all
status
echo "==> deployed. https://dowork.l8ai.cn · https://mobile.l8ai.cn (admin@agentsmesh.local / Ab123456)"

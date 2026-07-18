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
RELEASE_DEPLOY_COMMIT=""
REMOTE_WORKSPACE_MAY_EXIST=false
SECRET_MANIFESTS=(
  agentsmesh-secrets.yaml
  agentsmesh-pki-ca.yaml
  agentsmesh-access-token.yaml
  agentsmesh-regcred.yaml
)
# shellcheck source=release_source_guard.sh
source "${DIR}/release_source_guard.sh"
# shellcheck source=deploy-write-quiescence.sh
source "${DIR}/deploy-write-quiescence.sh"
# shellcheck source=internal_gitea_deploy.sh
source "${DIR}/internal_gitea_deploy.sh"
# shellcheck source=deploy_database.sh
source "${DIR}/deploy_database.sh"

echo "==> DoOps session ${SESSION} -> ${TARGET} (workspace ${WS})"

dexec()  { doops -session "${SESSION}" exec  --target "${TARGET}" --cmd "cd ${WS} && $1"; }

cleanup_release_workspace() {
  local result=$?
  trap - EXIT
  [[ -z "${RELEASE_BUNDLE}" || ! -d "${RELEASE_BUNDLE}" ]] || rm -rf "${RELEASE_BUNDLE}"
  rm -rf "${GEN}"
  if [[ "${APP_WRITES_STOPPED}" == true ]]; then
    echo "deployment failed; application writes remain stopped" >&2
  fi
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
  local tls="$1"
  shift
  local reference_hostname="$1"
  if dexec "kubectl -n ${NS} get secret ${tls} -o name >/dev/null"; then
    echo "==> using existing ${NS}/${tls}"
  else
    dexec "kubectl get secret ${tls} -n default -o yaml | sed -e '/namespace:/d' -e '/resourceVersion:/d' -e '/uid:/d' -e '/creationTimestamp:/d' | kubectl apply -n ${NS} -f -"
  fi
  dexec "test \"\$(kubectl -n ${NS} get secret ${tls} -o jsonpath='{.type}')\" = kubernetes.io/tls"
  for hostname in "$@"; do
    dexec "kubectl -n ${NS} get secret ${tls} -o jsonpath='{.data.tls\\.crt}' | base64 -d | openssl x509 -checkhost ${hostname} -noout"
    dexec "getent ahostsv4 ${hostname} >/dev/null"
  done
  for hostname in "$@"; do
    dexec "set -eu; reference_ip=\$(curl --insecure --silent --show-error --output /dev/null --write-out '%{remote_ip}' --connect-timeout 10 --max-time 20 https://${reference_hostname}/); hostname_ip=\$(curl --insecure --silent --show-error --output /dev/null --write-out '%{remote_ip}' --connect-timeout 10 --max-time 20 https://${hostname}/); test -n \"\${reference_ip}\"; test \"\${reference_ip}\" = \"\${hostname_ip}\""
  done
}

sync_worker_definitions() {
  dexec "image=\$(awk '\$1 == \"image:\" && \$2 ~ /agentsmesh\\/backend@sha256:/ { print \$2; exit } \$1 == \"-\" && \$2 == \"image:\" && \$3 ~ /agentsmesh\\/backend@sha256:/ { print \$3; exit }' /tmp/agentsmesh-release.yaml); test -n \"\${image}\"; sed \"s|__BACKEND_IMAGE__|\${image}|g\" 23-worker-definition-sync-job.yaml | kubectl apply -f -"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/worker-definition-sync --timeout=300s"
}

render_release() {
  dexec "kubectl kustomize . > /tmp/agentsmesh-release.yaml; image=\$(awk '\$1 == \"image:\" && \$2 ~ /agentsmesh\\/backend@sha256:/ { print \$2; exit } \$1 == \"-\" && \$2 == \"image:\" && \$3 ~ /agentsmesh\\/backend@sha256:/ { print \$3; exit }' /tmp/agentsmesh-release.yaml); test -n \"\${image}\"; digest=\${image##*@}; sed -i -e \"s|__BACKEND_DIGEST__|\${digest}|g\" -e \"s|__RELEASE_COMMIT__|${RELEASE_DEPLOY_COMMIT}|g\" /tmp/agentsmesh-release.yaml; bash verify_release_images.sh /tmp/agentsmesh-release.yaml"
}

apply_pinned_manifest() {
  local manifest="$1" image="$2" repository
  repository="${3:-${REG}/agentsmesh/${image}}"
  dexec "digest=\$(awk -v name='${repository}' '\$1 == \"-\" && \$2 == \"name:\" && \$3 == name { found=1; next } found && \$1 == \"digest:\" { print \$2; exit }' release/kustomization.yaml); test \"\${digest}\" != \"\"; sed -E \"s|image: ${repository}:[^[:space:]]+|image: ${repository}@\${digest}|g\" ${manifest} | kubectl apply -f -"
}

apply_backend_job() {
  local manifest="$1"
  dexec "image=\$(awk '\$1 == \"image:\" && \$2 ~ /agentsmesh\\/backend@sha256:/ { print \$2; exit } \$1 == \"-\" && \$2 == \"image:\" && \$3 ~ /agentsmesh\\/backend@sha256:/ { print \$3; exit }' /tmp/agentsmesh-release.yaml); test -n \"\${image}\"; digest=\${image##*@}; sed -e \"s|__BACKEND_IMAGE__|\${image}|g\" -e \"s|__BACKEND_DIGEST__|\${digest}|g\" -e \"s|__RELEASE_COMMIT__|${RELEASE_DEPLOY_COMMIT}|g\" ${manifest} | kubectl apply -f -"
}

apply_all() {
  echo "==> namespace + secrets"
  dexec "kubectl apply -f 00-namespace.yaml"
  dexec "chmod 600 generated-secrets/*.yaml; status=0; cleanup_status=0; kubectl apply -f generated-secrets || status=\$?; rm -f generated-secrets/*.yaml || cleanup_status=\$?; rmdir generated-secrets || cleanup_status=\$?; test \${status} -ne 0 || status=\${cleanup_status}; exit \${status}"
  echo "==> ensure wildcard TLS in ${NS}"
  ensure_tls_secret "l8ai-wildcard-tls" "dowork.l8ai.cn" "health-preview.l8ai.cn"
  render_release
  stop_application_writes
  require_empty_pending_commands
  ensure_internal_gitea
  migrate_database
  echo "==> apply workloads after migrations"
  dexec "kubectl apply -f /tmp/agentsmesh-release.yaml"
  dexec "kubectl -n ${NS} rollout status deploy/backend --timeout=300s"
  dexec "kubectl -n ${NS} rollout status deploy/marketplace --timeout=300s"
  dexec "kubectl -n ${NS} rollout status deploy/marketplace-web --timeout=300s"
  mark_application_writes_restored
  echo "==> seed + minio bucket"
  dexec "kubectl -n ${NS} delete job seed minio-setup worker-definition-sync --ignore-not-found"
  dexec "kubectl apply -f 21-seed-configmap.yaml"
  apply_pinned_manifest "22-seed-job.yaml" pgvector
  apply_pinned_manifest "13-minio-setup-job.yaml" mc
  dexec "kubectl -n ${NS} wait --for=condition=complete job/seed --timeout=300s"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/minio-setup --timeout=300s"
  sync_worker_definitions
}

status() {
  echo "==> rollout status"
  for d in gitea backend marketplace marketplace-web relay web web-admin mobile runner-e2e-echo; do
    dexec "kubectl -n ${NS} rollout status deploy/${d} --timeout=240s"
  done
  local preview_probe=release-preview-probe
  dexec "set -eu; body=\$(mktemp); trap 'rm -f \"\$body\"' EXIT; status=\$(curl --silent --show-error --output \"\$body\" --write-out '%{http_code}' --connect-timeout 10 --max-time 20 https://${preview_probe}.l8ai.cn/preview/${preview_probe}/); test \"\${status}\" = 401; grep -Fxq token_required \"\$body\""
  dexec "kubectl -n ${NS} get pods -o wide"
}

main() {
  local repo_root
  repo_root="$(cd "${DIR}/../../.." && pwd)"
  release_require_pushed_clean_tree "${repo_root}"
  release_verify_source_metadata "${repo_root}"
  release_verify_image_provenance "${repo_root}"
  release_verify_gitea_provenance "${repo_root}"
  RELEASE_DEPLOY_COMMIT="$(git -C "${repo_root}" rev-parse HEAD)"
  generate_cluster_secrets
  push_manifests
  apply_all
  status
  echo "==> deployed. https://dowork.l8ai.cn · https://market.l8ai.cn · https://mobile.l8ai.cn · https://<pod-key>.l8ai.cn (admin@agentsmesh.local / Ab123456)"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi

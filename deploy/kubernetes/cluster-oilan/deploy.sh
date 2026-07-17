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
  local tls="$1" hostname="$2"
  if dexec "kubectl -n ${NS} get secret ${tls} -o name >/dev/null"; then
    echo "==> using existing ${NS}/${tls}"
  else
    dexec "kubectl get secret ${tls} -n default -o yaml | sed -e '/namespace:/d' -e '/resourceVersion:/d' -e '/uid:/d' -e '/creationTimestamp:/d' | kubectl apply -n ${NS} -f -"
  fi
  dexec "test \"\$(kubectl -n ${NS} get secret ${tls} -o jsonpath='{.type}')\" = kubernetes.io/tls"
  dexec "kubectl -n ${NS} get secret ${tls} -o jsonpath='{.data.tls\\.crt}' | base64 -d | openssl x509 -checkhost ${hostname} -noout"
  dexec "getent ahostsv4 ${hostname} >/dev/null"
}

sync_worker_definitions() {
  dexec "image=\$(awk '\$1 == \"image:\" && \$2 ~ /agentsmesh\\/backend@sha256:/ { print \$2; exit } \$1 == \"-\" && \$2 == \"image:\" && \$3 ~ /agentsmesh\\/backend@sha256:/ { print \$3; exit }' /tmp/agentsmesh-release.yaml); test -n \"\${image}\"; sed \"s|__BACKEND_IMAGE__|\${image}|g\" 23-worker-definition-sync-job.yaml | kubectl apply -f -"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/worker-definition-sync --timeout=300s"
}

render_release() {
  dexec "kubectl kustomize . > /tmp/agentsmesh-release.yaml; image=\$(awk '\$1 == \"image:\" && \$2 ~ /agentsmesh\\/backend@sha256:/ { print \$2; exit } \$1 == \"-\" && \$2 == \"image:\" && \$3 ~ /agentsmesh\\/backend@sha256:/ { print \$3; exit }' /tmp/agentsmesh-release.yaml); test -n \"\${image}\"; digest=\${image##*@}; sed -i -e \"s|__BACKEND_DIGEST__|\${digest}|g\" -e \"s|__RELEASE_COMMIT__|${RELEASE_DEPLOY_COMMIT}|g\" /tmp/agentsmesh-release.yaml; bash verify_release_images.sh /tmp/agentsmesh-release.yaml"
}

apply_pinned_manifest() {
  local manifest="$1" image="$2"
  dexec "digest=\$(awk -v name='${REG}/agentsmesh/${image}' '\$1 == \"-\" && \$2 == \"name:\" && \$3 == name { found=1; next } found && \$1 == \"digest:\" { print \$2; exit }' release/kustomization.yaml); test \"\${digest}\" != \"\"; sed -E \"s|image: ${REG}/agentsmesh/${image}:[^[:space:]]+|image: ${REG}/agentsmesh/${image}@\${digest}|g\" ${manifest} | kubectl apply -f -"
}

apply_backend_job() {
  local manifest="$1"
  dexec "image=\$(awk '\$1 == \"image:\" && \$2 ~ /agentsmesh\\/backend@sha256:/ { print \$2; exit } \$1 == \"-\" && \$2 == \"image:\" && \$3 ~ /agentsmesh\\/backend@sha256:/ { print \$3; exit }' /tmp/agentsmesh-release.yaml); test -n \"\${image}\"; digest=\${image##*@}; sed -e \"s|__BACKEND_IMAGE__|\${image}|g\" -e \"s|__BACKEND_DIGEST__|\${digest}|g\" -e \"s|__RELEASE_COMMIT__|${RELEASE_DEPLOY_COMMIT}|g\" ${manifest} | kubectl apply -f -"
}

backup_database() {
  echo "==> backup database before migrations"
  dexec "set -eu; backup_dir=/root/backups/agentsmesh; timestamp=\$(date -u +%Y%m%dT%H%M%SZ); backup=\${backup_dir}/pre-migrate-${RELEASE_DEPLOY_COMMIT:0:12}-\${timestamp}.dump; umask 077; mkdir -p \"\${backup_dir}\"; kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" exec pg_dump --format=custom --no-owner --no-privileges --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\"' > \"\${backup}.tmp\"; test -s \"\${backup}.tmp\"; mv \"\${backup}.tmp\" \"\${backup}\"; sha256sum \"\${backup}\" > \"\${backup}.sha256\"; echo \"database backup: \${backup}\""
}

require_clean_migration_state() {
  local state
  state="$(dexec "kubectl -n ${NS} exec deploy/postgres -- sh -ceu 'PGPASSWORD=\"\$POSTGRES_PASSWORD\" psql --username=\"\$POSTGRES_USER\" --dbname=\"\$POSTGRES_DB\" --tuples-only --no-align --command=\"SELECT version, dirty FROM schema_migrations\"'" |
    tail -n 1 | tr -d '\r')"
  [[ "${state}" =~ ^[0-9]+\|f$ ]] || {
    echo "database migration state must be clean before deploy, got ${state}" >&2
    return 1
  }
  echo "==> current migration state ${state}"
}

migrate_database() {
  echo "==> apply migration prerequisites"
  dexec "kubectl apply -f 02-configmap.yaml -f 30-backend-rbac.yaml"
  require_clean_migration_state
  stop_application_writes
  backup_database
  apply_pinned_manifest "10-postgres.yaml" pgvector
  apply_pinned_manifest "11-redis.yaml" redis
  apply_pinned_manifest "12-minio.yaml" minio
  dexec "kubectl -n ${NS} rollout status deploy/postgres --timeout=300s"
  dexec "kubectl -n ${NS} delete job migrate --ignore-not-found"
  apply_backend_job "20-migrate-job.yaml"
  dexec "kubectl -n ${NS} wait --for=condition=complete job/migrate --timeout=300s"
}

apply_all() {
  echo "==> namespace + secrets"
  dexec "kubectl apply -f 00-namespace.yaml"
  dexec "chmod 600 generated-secrets/*.yaml; status=0; cleanup_status=0; kubectl apply -f generated-secrets || status=\$?; rm -f generated-secrets/*.yaml || cleanup_status=\$?; rmdir generated-secrets || cleanup_status=\$?; test \${status} -ne 0 || status=\${cleanup_status}; exit \${status}"
  echo "==> ensure wildcard TLS in ${NS}"
  ensure_tls_secret "l8ai-wildcard-tls" "dowork.l8ai.cn"
  ensure_tls_secret "dowork-preview-wildcard-tls" "health.preview.dowork.l8ai.cn"
  render_release
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
  for d in backend marketplace marketplace-web relay web web-admin mobile runner-e2e-echo; do
    dexec "kubectl -n ${NS} rollout status deploy/${d} --timeout=240s"
  done
  dexec "kubectl -n ${NS} get pods -o wide"
}

main() {
  local repo_root
  repo_root="$(cd "${DIR}/../../.." && pwd)"
  release_require_pushed_clean_tree "${repo_root}"
  release_verify_source_metadata "${repo_root}"
  release_verify_image_provenance "${repo_root}"
  RELEASE_DEPLOY_COMMIT="$(git -C "${repo_root}" rev-parse HEAD)"
  generate_cluster_secrets
  push_manifests
  apply_all
  status
  echo "==> deployed. https://dowork.l8ai.cn · https://market.l8ai.cn · https://mobile.l8ai.cn · https://<pod-key>.preview.dowork.l8ai.cn (admin@agentsmesh.local / Ab123456)"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi

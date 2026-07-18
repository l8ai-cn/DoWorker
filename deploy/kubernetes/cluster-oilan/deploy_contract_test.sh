#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
LOG="$TMP/doops.log"

mkdir -p "$TMP/bin"
cat > "$TMP/bin/doops" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "session" ]]; then
  printf 'ses-contract\n'
  exit 0
fi
case " $* " in
  *" push "*) printf 'push %s\n' "$*" >> "$DOOPS_LOG" ;;
  *" clean "*) printf 'clean %s\n' "$*" >> "$DOOPS_LOG" ;;
  *" exec "*)
    args=("$@")
    for ((index = 0; index < ${#args[@]}; index++)); do
      [[ "${args[index]}" == "--cmd" ]] || continue
      printf '%s\n' "${args[index + 1]}" >> "$DOOPS_LOG"
      if [[ "${args[index + 1]}" == *"SELECT version, dirty FROM schema_migrations"* ]]; then
        printf '222|f\n'
      fi
      if [[ "${args[index + 1]}" == *"SELECT COUNT(*) FROM pending_runner_commands"* ]]; then
        printf '0\n'
      fi
      if [[ "${args[index + 1]}" == *"get deploy backend -o jsonpath"* ]]; then
        printf '1\n'
      fi
      if [[ "${args[index + 1]}" == *"get deploy marketplace -o jsonpath"* ]]; then
        printf '1\n'
      fi
      if [[ "${args[index + 1]}" == *"get deploy/gitea -o jsonpath="*".spec.replicas"* ]]; then
        printf '1\n'
      fi
      exit 0
    done
    exit 1
    ;;
  *) exit 1 ;;
esac
EOF
chmod +x "$TMP/bin/doops"

PATH="$TMP/bin:$PATH" \
DOOPS_LOG="$LOG" \
DOOPS_SESSION="ses-contract" \
DOOPS_TARGET="contract-target" \
bash -c '
  set -euo pipefail
  source "$1/deploy.sh"
  release_require_pushed_clean_tree() { :; }
  release_verify_source_metadata() { :; }
  release_verify_image_provenance() { :; }
  release_verify_gitea_provenance() { :; }
  generate_cluster_secrets() {
    mkdir -p "${SEC}"
    for name in "${SECRET_MANIFESTS[@]}"; do
      printf "%s\n" "apiVersion: v1" "kind: Secret" > "${SEC}/${name}"
    done
  }
  main
' bash "$ROOT"

ROOT="$ROOT" bash -c '
  set -euo pipefail
  source "${ROOT}/deploy-write-quiescence.sh"
  NS=agentsmesh
  deployment_replicas() { printf "1\n"; }
  dexec() { [[ "$1" == *" scale "* ]]; }
  if stop_application_writes; then
    echo "pod deletion timeout was accepted" >&2
    exit 1
  fi
  [[ "${APP_WRITES_STOPPED}" == "true" ]]
'

ROOT="$ROOT" bash -c '
  set -euo pipefail
  source "${ROOT}/internal_gitea_deploy.sh"
  NS=agentsmesh
  REG=registry.example
  RELEASE_DEPLOY_COMMIT=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  RESTORE_LOG="$(mktemp)"
  trap "rm -f ${RESTORE_LOG}" EXIT
  dexec() {
    printf "%s\n" "$1" >> "${RESTORE_LOG}"
    case "$1" in
      *"get deploy/gitea -o jsonpath="*".spec.replicas"*) printf "1\n" ;;
      *"PRAGMA quick_check"*) return 1 ;;
    esac
  }
  apply_pinned_manifest() {
    printf "apply %s %s %s\n" "$1" "$2" "$3" >> "${RESTORE_LOG}"
  }
  if backup_internal_gitea; then
    echo "failed Gitea backup was accepted" >&2
    exit 1
  fi
  grep -F "scale deploy/gitea --replicas=0" "${RESTORE_LOG}" >/dev/null
  grep -F "scale deploy/gitea --replicas=1" "${RESTORE_LOG}" >/dev/null
  grep -F "rollout status deploy/gitea --timeout=300s" "${RESTORE_LOG}" >/dev/null
'

require_command() {
  grep -F -- "$1" "$LOG" >/dev/null || {
    printf 'missing remote command: %s\n' "$1" >&2
    exit 1
  }
}

line_number() {
  grep -n -F "$1" "$LOG" | head -1 | cut -d: -f1
}

require_command 'kubectl kustomize . > /tmp/agentsmesh-release.yaml'
require_command 'kubectl apply -f generated-secrets'
require_command 'rm -f generated-secrets/*.yaml'
require_command "name='repo.aiedulab.cn:8443/library/gitea'"
require_command 'scale deploy/backend deploy/marketplace --replicas=0'
require_command 'SELECT COUNT(*) FROM pending_runner_commands'
require_command 'scale deploy/gitea --replicas=0'
require_command 'wait --for=delete pod -l app=gitea'
require_command '15-gitea-backup-pod.yaml | kubectl apply -f -'
require_command 'wait --for=condition=Ready pod/gitea-backup'
require_command "sqlite3 /data/gitea/gitea.db 'PRAGMA quick_check;'"
require_command 'tar -C /data -czf -'
require_command 'sha256sum -c'
require_command 'delete pod gitea-backup --wait=true'
require_command '14-gitea.yaml | kubectl apply -f -'
require_command 'gitea-'
require_command 'rollout status deploy/gitea --timeout=300s'
require_command 'bash bootstrap_internal_gitea.sh agentsmesh'
require_command 'kubectl apply -f 02-configmap.yaml -f 30-backend-rbac.yaml'
require_command 'wait --for=delete pod -l app=backend'
require_command 'wait --for=delete pod -l app=marketplace'
require_command '10-postgres.yaml | kubectl apply -f -'
require_command '11-redis.yaml | kubectl apply -f -'
require_command '12-minio.yaml | kubectl apply -f -'
require_command 'kubectl -n agentsmesh rollout status deploy/postgres --timeout=300s'
require_command 'pg_dump --format=custom'
require_command '/root/backups/agentsmesh'
require_command 'pre-migrate-'
require_command '20-migrate-job.yaml | kubectl apply -f -'
require_command '__BACKEND_IMAGE__'
require_command '__BACKEND_DIGEST__'
require_command 'kubectl -n agentsmesh wait --for=condition=complete job/migrate --timeout=300s'
require_command 'kubectl apply -f /tmp/agentsmesh-release.yaml'
require_command 'kubectl -n agentsmesh rollout status deploy/backend --timeout=300s'
require_command "https://health-preview.l8ai.cn/"
require_command "--write-out '%{remote_ip}'"
require_command 'test -n "${reference_ip}"'
require_command 'test "${reference_ip}" = "${hostname_ip}"'
require_command 'https://release-preview-probe.l8ai.cn/preview/release-preview-probe/'
require_command 'test "${status}" = 401'
require_command 'grep -Fxq token_required "$body"'
require_command 'kubectl apply -f 21-seed-configmap.yaml'
require_command '22-seed-job.yaml | kubectl apply -f -'
require_command '13-minio-setup-job.yaml | kubectl apply -f -'
require_command '23-worker-definition-sync-job.yaml | kubectl apply -f -'

render_line="$(line_number 'kubectl kustomize . > /tmp/agentsmesh-release.yaml')"
stop_line="$(line_number 'scale deploy/backend deploy/marketplace --replicas=0')"
pending_line="$(line_number 'SELECT COUNT(*) FROM pending_runner_commands')"
gitea_stop_line="$(line_number 'scale deploy/gitea --replicas=0')"
gitea_wait_line="$(line_number 'wait --for=delete pod -l app=gitea')"
gitea_backup_pod_line="$(line_number '15-gitea-backup-pod.yaml | kubectl apply -f -')"
gitea_backup_ready_line="$(line_number 'wait --for=condition=Ready pod/gitea-backup')"
gitea_check_line="$(line_number "sqlite3 /data/gitea/gitea.db 'PRAGMA quick_check;'")"
gitea_backup_line="$(line_number 'tar -C /data -czf -')"
gitea_checksum_line="$(line_number 'sha256sum -c')"
gitea_backup_delete_line="$(line_number 'delete pod gitea-backup --wait=true')"
gitea_line="$(line_number '14-gitea.yaml | kubectl apply -f -')"
prereq_line="$(line_number 'kubectl apply -f 02-configmap.yaml -f 30-backend-rbac.yaml')"
backup_line="$(line_number 'pg_dump --format=custom')"
postgres_line="$(line_number '10-postgres.yaml | kubectl apply -f -')"
migrate_line="$(line_number '20-migrate-job.yaml | kubectl apply -f -')"
migrate_wait_line="$(line_number 'wait --for=condition=complete job/migrate')"
workload_line="$(line_number 'kubectl apply -f /tmp/agentsmesh-release.yaml')"
backend_line="$(line_number 'rollout status deploy/backend')"
sync_line="$(line_number '23-worker-definition-sync-job.yaml | kubectl apply -f -')"

(( render_line < stop_line &&
  stop_line < pending_line &&
  pending_line < gitea_stop_line &&
  gitea_stop_line < gitea_wait_line &&
  gitea_wait_line < gitea_backup_pod_line &&
  gitea_backup_pod_line < gitea_backup_ready_line &&
  gitea_backup_ready_line < gitea_check_line &&
  gitea_check_line == gitea_backup_line &&
  gitea_backup_line == gitea_checksum_line &&
  gitea_checksum_line < gitea_backup_delete_line &&
  gitea_backup_delete_line < gitea_line &&
  gitea_line < prereq_line &&
  prereq_line < backup_line &&
  backup_line < postgres_line &&
  postgres_line < migrate_line &&
  backup_line < migrate_line &&
  migrate_line < migrate_wait_line &&
  migrate_wait_line < workload_line &&
  workload_line < backend_line &&
  backend_line < sync_line )) || {
  printf 'full deployment command order is invalid\n' >&2
  exit 1
}

grep -F 'command: ["/app/server", "migrate", "up"]' "$ROOT/20-migrate-job.yaml" >/dev/null
! grep -A12 -F 'initContainers:' "$ROOT/30-backend.yaml" | grep -F 'name: migrate' >/dev/null
grep -F 'workspace-artifacts/' "$ROOT/13-minio-setup-job.yaml" >/dev/null
grep -F -- '--expire-days 1' "$ROOT/13-minio-setup-job.yaml" >/dev/null
grep -F 'PREVIEW_PUBLIC_ORIGIN: "https://l8ai.cn"' "$ROOT/02-configmap.yaml" >/dev/null
grep -F 'PREVIEW_COOKIE_MODE: "partitioned"' "$ROOT/02-configmap.yaml" >/dev/null
grep -F 'KB_GITEA_URL: "http://gitea:3000"' "$ROOT/02-configmap.yaml" >/dev/null
grep -F 'KB_GITEA_SSH_URL: "ssh://git@gitea.agentsmesh.svc.cluster.local:22"' "$ROOT/02-configmap.yaml" >/dev/null
grep -F 'name: agentsmesh-gitea' "$ROOT/30-backend.yaml" >/dev/null
grep -F 'readOnly: true' "$ROOT/15-gitea-backup-pod.yaml" >/dev/null
grep -F 'tail -f /dev/null' "$ROOT/15-gitea-backup-pod.yaml" >/dev/null
grep -F 'host: "*.l8ai.cn"' "$ROOT/44-preview-ingress.yaml" >/dev/null
grep -F 'secretName: l8ai-wildcard-tls' "$ROOT/44-preview-ingress.yaml" >/dev/null
grep -F 'ensure_tls_secret "l8ai-wildcard-tls" "dowork.l8ai.cn" "health-preview.l8ai.cn"' "$ROOT/deploy.sh" >/dev/null
! grep -F 'dowork-preview-wildcard-tls' \
  "$ROOT/02-configmap.yaml" "$ROOT/44-preview-ingress.yaml" "$ROOT/deploy.sh" "$ROOT/README.md" >/dev/null
probe_command="$(grep -F 'https://release-preview-probe.l8ai.cn/preview/release-preview-probe/' "$LOG")"
! grep -Fq -- '--insecure' <<<"$probe_command"
grep -F 'release_require_pushed_clean_tree' "$ROOT/deploy.sh" >/dev/null
grep -F 'clean -session ses-contract' "$LOG" >/dev/null

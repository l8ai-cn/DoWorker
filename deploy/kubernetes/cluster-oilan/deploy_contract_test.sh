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
  *" exec "*)
    args=("$@")
    for ((index = 0; index < ${#args[@]}; index++)); do
      [[ "${args[index]}" == "--cmd" ]] || continue
      printf '%s\n' "${args[index + 1]}" >> "$DOOPS_LOG"
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
  generate_cluster_secrets() {
    mkdir -p "${SEC}"
    printf "%s\n" "apiVersion: v1" "kind: Secret" > "${SEC}/contract.yaml"
  }
  main
' bash "$ROOT"

require_command() {
  grep -F "$1" "$LOG" >/dev/null || {
    printf 'missing remote command: %s\n' "$1" >&2
    exit 1
  }
}

line_number() {
  grep -n -F "$1" "$LOG" | head -1 | cut -d: -f1
}

require_command 'kubectl kustomize . > /tmp/agentsmesh-release.yaml'
require_command 'kubectl apply -f 02-configmap.yaml -f 10-postgres.yaml'
require_command 'kubectl -n agentsmesh rollout status deploy/postgres --timeout=300s'
require_command '20-migrate-job.yaml | kubectl apply -f -'
require_command '__BACKEND_IMAGE__'
require_command '__BACKEND_DIGEST__'
require_command 'kubectl -n agentsmesh wait --for=condition=complete job/migrate --timeout=300s'
require_command 'kubectl apply -f /tmp/agentsmesh-release.yaml'
require_command 'kubectl -n agentsmesh rollout status deploy/backend --timeout=300s'
require_command 'kubectl apply -f 21-seed-configmap.yaml -f 22-seed-job.yaml -f 13-minio-setup-job.yaml'
require_command '23-worker-definition-sync-job.yaml | kubectl apply -f -'

render_line="$(line_number 'kubectl kustomize . > /tmp/agentsmesh-release.yaml')"
prereq_line="$(line_number 'kubectl apply -f 02-configmap.yaml -f 10-postgres.yaml')"
migrate_line="$(line_number '20-migrate-job.yaml | kubectl apply -f -')"
migrate_wait_line="$(line_number 'wait --for=condition=complete job/migrate')"
workload_line="$(line_number 'kubectl apply -f /tmp/agentsmesh-release.yaml')"
backend_line="$(line_number 'rollout status deploy/backend')"
sync_line="$(line_number '23-worker-definition-sync-job.yaml | kubectl apply -f -')"

(( render_line < prereq_line &&
  prereq_line < migrate_line &&
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
grep -F 'PREVIEW_PUBLIC_ORIGIN: "https://preview.l8ai.cn"' "$ROOT/02-configmap.yaml" >/dev/null
grep -F 'host: preview.l8ai.cn' "$ROOT/44-preview-ingress.yaml" >/dev/null

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

if env \
  PATH="$TMP/bin:$PATH" \
  DOOPS_LOG="$LOG" \
  DOOPS_SESSION="ses-contract" \
  DOOPS_TARGET="contract-target" \
  DOSQL_RELEASE_DB_TARGET="db_agentsmesh_prod_postgres" \
  DOSQL_RELEASE_DB_MODE="production" \
  DOSQL_RELEASE_DB_SESSION="dosql-contract" \
  DOSQL_RELEASE_MIGRATION_VERSION="$(find "$ROOT/../../../backend/migrations" -name '*.up.sql' -exec basename {} \; | awk -F_ '{ print $1 }' | sort -n | tail -1)" \
  DOSQL_RELEASE_CHANGE_ID="change-contract" \
  DOSQL_RELEASE_OPERATION_ID="dbop-contract" \
  "$ROOT/deploy-mobile-access.sh" >/dev/null 2>&1; then
  echo "mobile deploy accepted missing canonical DoSql evidence" >&2
  exit 1
fi

if [[ -e "$LOG" ]] && grep -F ' push ' "$LOG" >/dev/null; then
  echo "mobile deploy pushed manifests before canonical DoSql evidence passed" >&2
  exit 1
fi

PATH="$TMP/bin:$PATH" \
DOOPS_LOG="$LOG" \
DOOPS_SESSION="ses-contract" \
DOOPS_TARGET="contract-target" \
bash -c '
  set -euo pipefail
  source "$1/deploy-mobile-access.sh"
  require_dosql_database_evidence() { :; }
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

require_command 'kubectl apply -f 02-configmap.yaml'
require_command 'kubectl apply -f 30-backend.yaml'
require_command 'kubectl -n agentsmesh rollout status deploy/backend --timeout=240s'
require_command 'kubectl -n agentsmesh exec deploy/backend -- /app/worker-definition-sync'
require_command 'kubectl apply -f 31-relay.yaml -f 42-mobile.yaml -f 43-mobile-ingress.yaml'

config_line="$(line_number 'kubectl apply -f 02-configmap.yaml')"
backend_line="$(line_number 'kubectl apply -f 30-backend.yaml')"
sync_line="$(line_number 'kubectl -n agentsmesh exec deploy/backend -- /app/worker-definition-sync')"
workload_line="$(line_number 'kubectl apply -f 31-relay.yaml -f 42-mobile.yaml -f 43-mobile-ingress.yaml')"

(( config_line < backend_line && backend_line < sync_line && sync_line < workload_line )) || {
  printf 'mobile deployment command order is invalid\n' >&2
  exit 1
}

! grep -F '20-migrate-job.yaml' "$LOG" >/dev/null
! grep -F 'job/migrate' "$LOG" >/dev/null

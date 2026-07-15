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
"$ROOT/deploy-mobile-access.sh"

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
require_command 'kubectl -n agentsmesh delete job migrate worker-definition-sync --ignore-not-found'
require_command '20-migrate-job.yaml | kubectl apply -f -'
require_command 'agentsmesh.ai/verified-image-digest'
require_command 'kubectl -n agentsmesh wait --for=condition=complete job/migrate --timeout=300s'
require_command '23-worker-definition-sync-job.yaml | kubectl apply -f -'
require_command 'kubectl -n agentsmesh wait --for=condition=complete job/worker-definition-sync --timeout=300s'
require_command 'kubectl apply -f 30-backend.yaml'
require_command 'kubectl -n agentsmesh rollout status deploy/backend --timeout=240s'
require_command 'kubectl apply -f 31-relay.yaml -f 42-mobile.yaml -f 43-mobile-ingress.yaml'

config_line="$(line_number 'kubectl apply -f 02-configmap.yaml')"
migrate_line="$(line_number '20-migrate-job.yaml | kubectl apply -f -')"
sync_line="$(line_number '23-worker-definition-sync-job.yaml | kubectl apply -f -')"
backend_line="$(line_number 'kubectl apply -f 30-backend.yaml')"
workload_line="$(line_number 'kubectl apply -f 31-relay.yaml -f 42-mobile.yaml -f 43-mobile-ingress.yaml')"

(( config_line < migrate_line && migrate_line < sync_line && sync_line < backend_line && backend_line < workload_line )) || {
  printf 'mobile deployment command order is invalid\n' >&2
  exit 1
}

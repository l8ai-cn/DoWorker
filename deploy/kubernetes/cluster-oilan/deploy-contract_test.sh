#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
DEPLOY_DIR="$TMP/deploy"
LOG="$TMP/doops.log"

mkdir -p "$DEPLOY_DIR" "$TMP/bin" "$TMP/home/.docker"
cp "$ROOT/deploy.sh" "$ROOT/30-backend.yaml" "$DEPLOY_DIR/"
printf '{"credsStore":"contract"}\n' > "$TMP/home/.docker/config.json"

cat > "$TMP/bin/doops" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
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

cat > "$TMP/bin/kubectl" <<'EOF'
#!/usr/bin/env bash
printf 'apiVersion: v1\nkind: Secret\nmetadata:\n  name: contract\n'
EOF

cat > "$TMP/bin/openssl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
args=("$@")
for ((index = 0; index < ${#args[@]}; index++)); do
  [[ "${args[index]}" == "-out" ]] || continue
  output="${args[index + 1]}"
  mkdir -p "$(dirname "$output")"
  printf 'contract\n' > "$output"
done
EOF

cat > "$TMP/bin/docker-credential-contract" <<'EOF'
#!/usr/bin/env bash
printf '{"Username":"contract","Secret":"contract"}\n'
EOF

cat > "$TMP/bin/curl" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF

chmod +x "$TMP/bin/"*

PATH="$TMP/bin:$PATH" \
HOME="$TMP/home" \
DOOPS_LOG="$LOG" \
DOOPS_SESSION="ses-contract" \
DOOPS_TARGET="contract-target" \
"$DEPLOY_DIR/deploy.sh"

require_command() {
  grep -F "$1" "$LOG" >/dev/null || {
    printf 'missing remote command: %s\n' "$1" >&2
    exit 1
  }
}

line_number() {
  grep -n -F "$1" "$LOG" | head -1 | cut -d: -f1
}

backend_image="$(awk '$1 == "image:" && $2 ~ /agentsmesh\/backend@sha256:/ { print $2; exit }' "$DEPLOY_DIR/30-backend.yaml")"
require_command "${backend_image}"
require_command '20-migrate-job.yaml | kubectl apply -f -'
require_command '23-worker-definition-sync-job.yaml | kubectl apply -f -'
require_command 'kubectl -n agentsmesh wait --for=condition=complete job/migrate --timeout=300s'
require_command 'kubectl -n agentsmesh wait --for=condition=complete job/seed --timeout=300s'
require_command 'kubectl -n agentsmesh wait --for=condition=complete job/worker-definition-sync --timeout=300s'
require_command 'kubectl apply -k .'

migrate_apply="$(line_number '20-migrate-job.yaml | kubectl apply -f -')"
migrate_wait="$(line_number 'job/migrate --timeout=300s')"
seed_wait="$(line_number 'job/seed --timeout=300s')"
sync_apply="$(line_number '23-worker-definition-sync-job.yaml | kubectl apply -f -')"
sync_wait="$(line_number 'job/worker-definition-sync --timeout=300s')"
workloads="$(line_number 'kubectl apply -k .')"

(( migrate_apply < migrate_wait &&
   migrate_wait < seed_wait &&
   seed_wait < sync_apply &&
   sync_apply < sync_wait &&
   sync_wait < workloads )) || {
  printf 'deployment command order is invalid\n' >&2
  exit 1
}

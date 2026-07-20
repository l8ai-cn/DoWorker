#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
DEPLOY_DIR="$TMP/deploy"
LOG="$TMP/doops.log"

mkdir -p "$DEPLOY_DIR" "$TMP/bin" "$TMP/home/.docker"
cp -R "$ROOT"/. "$DEPLOY_DIR/"
printf '{"credsStore":"contract"}\n' > "$TMP/home/.docker/config.json"

cat > "$TMP/bin/doops" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
case " $* " in
  *" push "*) printf 'push %s\n' "$*" >> "$DOOPS_LOG" ;;
  *" clean "*) printf 'clean %s\n' "$*" >> "$DOOPS_LOG" ;;
  *" exec "*)
    args=("$@")
    for ((index = 0; index < ${#args[@]}; index++)); do
      [[ "${args[index]}" == "--cmd" ]] || continue
      printf '%s\n' "${args[index + 1]}" >> "$DOOPS_LOG"
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
bash -c '
  set -euo pipefail
  source "$1/deploy.sh"
  release_require_pushed_clean_tree() { :; }
  release_verify_source_metadata() { :; }
  release_verify_image_provenance() { :; }
  release_verify_gitea_provenance() { :; }
  require_dosql_database_evidence() { :; }
  generate_cluster_secrets() {
    mkdir -p "${SEC}"
    for name in "${SECRET_MANIFESTS[@]}"; do
      printf "%s\n" "apiVersion: v1" "kind: Secret" > "${SEC}/${name}"
    done
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

backend_image="$(awk '$1 == "image:" && $2 ~ /agentsmesh\/backend@sha256:/ { print $2; exit }' "$ROOT/30-backend.yaml")"
require_command "${backend_image}"
require_command '23-worker-definition-sync-job.yaml | kubectl apply -f -'
require_command 'kubectl -n agentsmesh wait --for=condition=complete job/worker-definition-sync --timeout=300s'
require_command 'kubectl apply -f /tmp/agentsmesh-release.yaml'

sync_apply="$(line_number '23-worker-definition-sync-job.yaml | kubectl apply -f -')"
sync_wait="$(line_number 'job/worker-definition-sync --timeout=300s')"
workloads="$(line_number 'kubectl apply -f /tmp/agentsmesh-release.yaml')"

(( workloads < sync_apply &&
   sync_apply < sync_wait )) || {
  printf 'deployment command order is invalid\n' >&2
  exit 1
}

! grep -F '20-migrate-job.yaml' "$LOG" >/dev/null
! grep -F 'job/migrate' "$LOG" >/dev/null
! grep -F 'job/seed' "$LOG" >/dev/null

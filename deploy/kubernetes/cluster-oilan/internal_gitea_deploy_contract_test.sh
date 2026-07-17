#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT
TOKEN="contract-token-123456"
HOST_KEY="AAAAC3NzaC1lZDI1NTE5AAAAITestHostKey"

mkdir -p "${TMP}/bin"
cat > "${TMP}/bin/curl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
[[ " $* " == *" --config - "* ]]
config="$(cat)"
state="${CONTRACT_STATE}"

if [[ "${config}" == *'user = "agentsmesh-service:'* ]]; then
  count="$(cat "${state}.deletes" 2>/dev/null || printf 0)"
  printf '%s' "$((count + 1))" > "${state}.deletes"
  if [[ -f "${state}.active" ]]; then
    rm -f "${state}.active"
    printf '204'
  else
    printf '404'
  fi
  exit 0
fi

if [[ "${config}" != *"Authorization: token ${CONTRACT_TOKEN}"* ]]; then
  echo "missing expected token authentication" >&2
  exit 1
fi
if [[ "${config}" == *'write-out = "%{http_code}"'* ]]; then
  [[ -f "${state}.active" ]] && printf '200' || printf '401'
  exit 0
fi
if [[ "${CONTRACT_SCENARIO}" == wrong_token && "${config}" == *"/api/v1/user"* ]]; then
  printf '{"login":"wrong-admin","is_admin":true}\n'
  exit 0
fi
case "${config}" in
  *"/api/v1/user"*) printf '{"login":"agentsmesh-service","is_admin":true}\n' ;;
  *"/api/v1/admin/users?limit=1"*) printf '[]\n' ;;
  *) echo "unexpected Gitea API request" >&2; exit 1 ;;
esac
EOF

cat > "${TMP}/bin/kubectl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "${COMMAND_LOG}"
case " $* " in
  *" get secret agentsmesh-gitea "*)
    case "${CONTRACT_SCENARIO}" in
      existing|wrong_token) printf '%s' "${CONTRACT_TOKEN}" | base64 | tr -d '\n' ;;
    esac
    ;;
  *" get pods -l app=gitea "*) printf 'gitea-pod\n' ;;
  *" exec gitea-pod -- su git -c gitea admin user list"*)
    printf 'ID Username Email IsActive IsAdmin 2FA\n'
    [[ "${CONTRACT_SCENARIO}" == fresh ]] ||
      printf '1 agentsmesh-service agentsmesh-service@agentsmesh.invalid true true false\n'
    ;;
  *" admin user create --username agentsmesh-service "*) ;;
  *" exec -i gitea-pod "*)
    IFS= read -r password
    [[ "${password}" =~ ^[a-f0-9]{64}$ ]]
    ;;
  *" admin user generate-access-token "*)
    touch "${CONTRACT_STATE}.active"
    printf '%s\n' "${CONTRACT_TOKEN}"
    ;;
  *" cat /data/ssh/ssh_host_ed25519_key.pub"*)
    printf 'ssh-ed25519 %s gitea\n' "${CONTRACT_HOST_KEY}"
    ;;
  *" apply -f -"*)
    cat > "${APPLY_MANIFEST}"
    [[ "${CONTRACT_SCENARIO}" != apply_fail ]]
    ;;
  *) echo "unexpected kubectl command: $*" >&2; exit 1 ;;
esac
EOF
chmod +x "${TMP}/bin/"*

run_contract() {
  local scenario="$1"
  local log="${TMP}/${scenario}.log"
  local state="${TMP}/${scenario}.state"
  local manifest="${TMP}/${scenario}.yaml"
  CONTRACT_SCENARIO="${scenario}" CONTRACT_TOKEN="${TOKEN}" \
    CONTRACT_HOST_KEY="${HOST_KEY}" CONTRACT_STATE="${state}" \
    COMMAND_LOG="${log}" APPLY_MANIFEST="${manifest}" \
    PATH="${TMP}/bin:${PATH}" \
    bash "${ROOT}/bootstrap_internal_gitea.sh" agentsmesh
}

run_contract fresh
fresh_log="${TMP}/fresh.log"
fresh_manifest="${TMP}/fresh.yaml"
grep -F 'admin user create --username agentsmesh-service' "${fresh_log}" >/dev/null
grep -F 'admin user change-password' "${fresh_log}" >/dev/null
grep -F 'generate-access-token --username agentsmesh-service --token-name agentsmesh-backend' \
  "${fresh_log}" >/dev/null
[[ "$(cat "${TMP}/fresh.state.deletes")" == 1 ]]
grep -F 'KB_GITEA_TOKEN:' "${fresh_manifest}" >/dev/null
grep -F 'KB_GITEA_KNOWN_HOSTS:' "${fresh_manifest}" >/dev/null
! grep -F "${TOKEN}" "${fresh_log}" "${fresh_manifest}" >/dev/null

run_contract existing
existing_log="${TMP}/existing.log"
! grep -F 'change-password' "${existing_log}" >/dev/null
! grep -F 'generate-access-token' "${existing_log}" >/dev/null
grep -F 'cat /data/ssh/ssh_host_ed25519_key.pub' "${existing_log}" >/dev/null

if run_contract wrong_token 2>/dev/null; then
  echo "token belonging to the wrong Gitea user was accepted" >&2
  exit 1
fi

if run_contract apply_fail 2>/dev/null; then
  echo "failed Secret persistence was accepted" >&2
  exit 1
fi
[[ "$(cat "${TMP}/apply_fail.state.deletes")" == 2 ]]
[[ ! -f "${TMP}/apply_fail.state.active" ]]
! grep -F 'delete-access-token' "${TMP}/apply_fail.log" >/dev/null

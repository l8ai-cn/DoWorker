#!/usr/bin/env bash
set -euo pipefail

namespace="${1:?namespace is required}"
secret_name="agentsmesh-gitea"
secret_key="KB_GITEA_TOKEN"
known_hosts_key="KB_GITEA_KNOWN_HOSTS"
username="agentsmesh-service"
token_name="agentsmesh-backend"
base_url="http://gitea.${namespace}.svc.cluster.local:3000"

token_request() {
  local token="$1" method="$2" route="$3"
  printf 'silent\nshow-error\nfail\nconnect-timeout = 10\nmax-time = 20\nheader = "Authorization: token %s"\nrequest = "%s"\nurl = "%s%s"\n' \
    "${token}" "${method}" "${base_url}" "${route}" |
    curl --config -
}

validate_token() {
  local token="$1" profile
  [[ "${token}" =~ ^[A-Za-z0-9_-]{20,}$ ]] || return 1
  profile="$(token_request "${token}" GET "/api/v1/user")"
  profile="$(printf '%s' "${profile}" | tr -d '[:space:]')"
  [[ "${profile}" == *"\"login\":\"${username}\""* ]] || return 1
  [[ "${profile}" == *'"is_admin":true'* ]] || return 1
  token_request "${token}" GET "/api/v1/admin/users?limit=1" >/dev/null || return 1
}

token_status() {
  local token="$1" route="$2"
  printf 'silent\nshow-error\nconnect-timeout = 10\nmax-time = 20\noutput = "/dev/null"\nwrite-out = "%%{http_code}"\nheader = "Authorization: token %s"\nurl = "%s%s"\n' \
    "${token}" "${base_url}" "${route}" |
    curl --config -
}

delete_token_with_password() {
  local password="$1" status
  status="$(
    printf 'silent\nshow-error\nconnect-timeout = 10\nmax-time = 20\noutput = "/dev/null"\nwrite-out = "%%{http_code}"\nuser = "%s:%s"\nrequest = "DELETE"\nurl = "%s/api/v1/users/%s/tokens/%s"\n' \
      "${username}" "${password}" "${base_url}" "${username}" "${token_name}" |
      curl --config -
  )"
  [[ "${status}" == "204" || "${status}" == "404" ]]
}

encoded="$(
  kubectl -n "${namespace}" get secret "${secret_name}" \
    --ignore-not-found \
    -o "jsonpath={.data.${secret_key}}"
)"
token_generated=false
if [[ -n "${encoded}" ]]; then
  token="$(printf '%s' "${encoded}" | base64 -d)"
  [[ "${token}" =~ ^[A-Za-z0-9_-]{20,}$ ]] || {
    echo "existing ${secret_name}/${secret_key} is malformed" >&2
    exit 1
  }
  validate_token "${token}" || {
    echo "existing ${secret_name}/${secret_key} is rejected by Gitea" >&2
    exit 1
  }
  echo "==> using existing internal Gitea service token"
fi

pod="$(
  kubectl -n "${namespace}" get pods -l app=gitea \
    -o 'jsonpath={.items[0].metadata.name}'
)"
[[ -n "${pod}" ]] || {
  echo "internal Gitea pod is unavailable" >&2
  exit 1
}

cleanup_generated_token() {
  local result=$?
  [[ "${token_generated}" == true ]] || return 0
  delete_token_with_password "${bootstrap_password}" || {
    echo "failed to revoke generated Gitea token ${token_name}" >&2
    return 1
  }
  [[ "$(token_status "${token}" "/api/v1/user")" == "401" ]] || {
    echo "generated Gitea token ${token_name} remained valid after revocation" >&2
    return 1
  }
  return "${result}"
}
trap cleanup_generated_token EXIT

if [[ -z "${encoded}" ]]; then
  if ! kubectl -n "${namespace}" exec "${pod}" -- \
    su git -c 'gitea admin user list' |
    awk -v expected="${username}" 'NR > 1 && $2 == expected { found = 1 } END { exit !found }'; then
    kubectl -n "${namespace}" exec "${pod}" -- \
      su git -c "gitea admin user create --username ${username} --random-password --email ${username}@agentsmesh.invalid --admin --must-change-password=false" \
      >/dev/null
  fi
  bootstrap_password="$(openssl rand -hex 32)"
  printf '%s\n' "${bootstrap_password}" |
    kubectl -n "${namespace}" exec -i "${pod}" -- \
      env GITEA_BOOTSTRAP_USER="${username}" sh -ceu '
        IFS= read -r password
        su git -c "gitea admin user change-password --username \"${GITEA_BOOTSTRAP_USER}\" --password \"${password}\""
      ' >/dev/null
  delete_token_with_password "${bootstrap_password}"
  token="$(
    kubectl -n "${namespace}" exec "${pod}" -- \
      su git -c "gitea admin user generate-access-token --username ${username} --token-name ${token_name} --raw --scopes all" |
      tail -n 1 |
      tr -d '\r\n'
  )"
  [[ "${token}" =~ ^[A-Za-z0-9_-]{20,}$ ]] || {
    echo "Gitea returned an invalid service token" >&2
    exit 1
  }
  token_generated=true
  validate_token "${token}"
fi

host_public_key="$(
  kubectl -n "${namespace}" exec "${pod}" -- \
    cat /data/ssh/ssh_host_ed25519_key.pub |
    tr -d '\r\n'
)"
read -r host_key_type host_key_material _ <<<"${host_public_key}"
[[ "${host_key_type}" == "ssh-ed25519" && "${host_key_material}" =~ ^[A-Za-z0-9+/=]+$ ]] || {
  echo "Gitea returned an invalid SSH host key" >&2
  exit 1
}
known_hosts="gitea.${namespace}.svc.cluster.local ${host_key_type} ${host_key_material}"
encoded_token="$(printf '%s' "${token}" | base64 | tr -d '\r\n')"
encoded_known_hosts="$(printf '%s' "${known_hosts}" | base64 | tr -d '\r\n')"
printf 'apiVersion: v1\nkind: Secret\nmetadata:\n  name: %s\n  namespace: %s\ntype: Opaque\ndata:\n  %s: %s\n  %s: %s\n' \
  "${secret_name}" "${namespace}" "${secret_key}" "${encoded_token}" \
  "${known_hosts_key}" "${encoded_known_hosts}" |
  kubectl apply -f -
token_generated=false
unset token
unset encoded_token
unset encoded_known_hosts
unset known_hosts
unset bootstrap_password
echo "==> initialized internal Gitea service token"

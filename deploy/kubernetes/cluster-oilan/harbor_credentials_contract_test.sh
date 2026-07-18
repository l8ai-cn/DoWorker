#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT
export HOME="${TMP}/home"
export DOCKER_CONFIG="${HOME}/.docker"
export REG="registry.example:8443"
mkdir -p "${DOCKER_CONFIG}" "${TMP}/bin"

# shellcheck disable=SC1091
source "${ROOT}/harbor-image-publishing.sh"
# shellcheck disable=SC1091
source "${ROOT}/harbor_immutable_release.sh"

assert_creds() {
  local expected_user="$1" expected_secret="$2" credentials
  credentials="$(harbor_creds)"
  [[ "$(jq -r '.Username' <<<"${credentials}")" == "${expected_user}" ]]
  [[ "$(jq -r '.Secret' <<<"${credentials}")" == "${expected_secret}" ]]
}

auth="$(printf 'inline-user:inline-secret' | base64)"
jq -n --arg auth "${auth}" \
  '{auths: {"https://registry.example:8443/v1/": {auth: $auth}}}' \
  > "${DOCKER_CONFIG}/config.json"
assert_creds inline-user inline-secret

cat > "${TMP}/bin/docker-credential-test" <<'EOF'
#!/usr/bin/env bash
read -r server
[[ "${server}" == "registry.example:8443" ]]
printf '{"Username":"helper-user","Secret":"helper-secret"}\n'
EOF
chmod +x "${TMP}/bin/docker-credential-test"
export PATH="${TMP}/bin:${PATH}"

jq -n '{credHelpers: {"registry.example:8443": "test"}}' \
  > "${DOCKER_CONFIG}/config.json"
assert_creds helper-user helper-secret

jq -n '{credsStore: "test"}' > "${DOCKER_CONFIG}/config.json"
assert_creds helper-user helper-secret

CURL_LOG="${TMP}/curl.log"
curl() {
  printf '%q ' "$@" > "${CURL_LOG}"
  if [[ "$*" == *"/service/token"* ]]; then
    printf '{"token":"test-token"}'
  else
    printf '201'
  fi
}
jq -n --arg auth "${auth}" \
  '{auths: {"registry.example:8443": {auth: $auth}}}' \
  > "${DOCKER_CONFIG}/config.json"
ensure_project
if grep -Eq -- '(^| )(-k|--insecure)( |$)' "${CURL_LOG}"; then
  echo "Harbor API call disabled TLS verification" >&2
  exit 1
fi

touch "${TMP}/harbor-ca.pem"
HARBOR_CA_CERT="${TMP}/harbor-ca.pem" ensure_project
grep -Fq -- "--cacert ${TMP}/harbor-ca.pem" "${CURL_LOG}"

harbor_load_credentials "${REG}"
[[ "${HARBOR_USERNAME}" == "inline-user" ]]
[[ "${HARBOR_PASSWORD}" == "inline-secret" ]]
HARBOR_CA_CERT="${TMP}/harbor-ca.pem" harbor_registry_token "${REG}" >/dev/null
grep -Fq -- "--cacert ${TMP}/harbor-ca.pem" "${CURL_LOG}"
grep -Fq -- "https://${REG}/service/token" "${CURL_LOG}"

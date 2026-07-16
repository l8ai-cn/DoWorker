#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
GENERATOR="${ROOT}/deploy/kubernetes/cluster-oilan/cluster_secret_generation.sh"
DEPLOY="${ROOT}/deploy/kubernetes/cluster-oilan/deploy.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

NS=agentsmesh
GEN="${TMP}/gen"
SEC="${GEN}/secrets"
REG=registry.example.test
HOME="${TMP}/home"
mkdir -p "${GEN}" "${HOME}/.docker" "${TMP}/bin"
source "$GENERATOR"

cluster_secret_data() {
  return 0
}

cat >"${GEN}/env" <<'EOF'
DB_PASSWORD=contract-db
JWT_SECRET=contract-jwt
INTERNAL_API_SECRET=contract-internal
MINIO_ROOT_PASSWORD=contract-minio
EOF
openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:prime256v1 \
  -out "${GEN}/ca.key" 2>/dev/null
openssl req -x509 -new -key "${GEN}/ca.key" -days 1 -out "${GEN}/ca.crt" \
  -subj "/CN=Contract CA" 2>/dev/null
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 \
  -out "${GEN}/access-token-private.pem" 2>/dev/null
openssl pkey -in "${GEN}/access-token-private.pem" -pubout \
  -out "${GEN}/access-token-public.pem" 2>/dev/null

printf '{"credsStore":"contract"}\n' >"${HOME}/.docker/config.json"
cat >"${TMP}/bin/docker-credential-contract" <<'EOF'
#!/usr/bin/env bash
cat >/dev/null
printf '{"Username":"release-user","Secret":"release-pass"}\n'
EOF
chmod +x "${TMP}/bin/docker-credential-contract"

PATH="${TMP}/bin:${PATH}" generate_cluster_secrets

mode() {
  if [[ "$(uname -s)" == Darwin ]]; then
    stat -f '%Lp' "$1"
  else
    stat -c '%a' "$1"
  fi
}

secret_value() {
  awk -v key="$2:" '$1 == key { print $2 }' "$1" | base64 -d
}

test "$(mode "$SEC")" = 700
for name in agentsmesh-secrets agentsmesh-pki-ca agentsmesh-access-token agentsmesh-regcred; do
  test "$(mode "${SEC}/${name}.yaml")" = 600
done

app="${SEC}/agentsmesh-secrets.yaml"
test "$(secret_value "$app" DB_PASSWORD)" = contract-db
test "$(secret_value "$app" STORAGE_SECRET_KEY)" = contract-minio
test "$(secret_value "$app" MARKETPLACE_DATABASE_URL)" = \
  'postgres://agentsmesh:contract-db@postgres:5432/agentsmesh?sslmode=disable'

secret_value "${SEC}/agentsmesh-pki-ca.yaml" ca.crt >"${TMP}/ca.crt"
cmp "${GEN}/ca.crt" "${TMP}/ca.crt"
secret_value "${SEC}/agentsmesh-access-token.yaml" public.pem >"${TMP}/public.pem"
cmp "${GEN}/access-token-public.pem" "${TMP}/public.pem"

registry="${SEC}/agentsmesh-regcred.yaml"
grep -Fq 'type: kubernetes.io/dockerconfigjson' "$registry"
docker_config="$(secret_value "$registry" .dockerconfigjson)"
test "$(jq -r '.auths["registry.example.test"].username' <<<"$docker_config")" = release-user
test "$(jq -r '.auths["registry.example.test"].password' <<<"$docker_config")" = release-pass
test "$(jq -r '.auths["registry.example.test"].auth' <<<"$docker_config")" = \
  cmVsZWFzZS11c2VyOnJlbGVhc2UtcGFzcw==

if grep -Eq 'kubectl create secret|--from-literal|--docker-password|python3' "$GENERATOR"; then
  echo "secret generation exposes values in argv or uses unstable local tools" >&2
  exit 1
fi
if grep -Eq 'base64.*kubectl apply|echo .*base64 -d|doops .* write ' "$DEPLOY"; then
  echo "secret manifests must not be embedded in remote command arguments" >&2
  exit 1
fi
grep -Fq 'doops -session "${SESSION}" push' "$DEPLOY"
grep -Fq 'for name in "${SECRET_MANIFESTS[@]}"' "$DEPLOY"
grep -Fq 'rm -f generated-secrets/*.yaml' "$DEPLOY"
grep -Fq 'clean --target "${TARGET}" --workspace "${SESSION}"' "$DEPLOY"

echo "OILAN secret manifest generation contract passed"

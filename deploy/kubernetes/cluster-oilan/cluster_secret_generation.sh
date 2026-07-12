# shellcheck shell=bash

cluster_secret_data() {
  local secret_name="$1" secret_key="$2" escaped_key
  escaped_key="${secret_key//./\\.}"
  doops -session "${SESSION}" exec --target "${TARGET}" \
    --cmd "printf '__AGENTSMESH_SECRET__'; kubectl -n ${NS} get secret ${secret_name} --ignore-not-found -o jsonpath='{.data.${escaped_key}}'" |
    sed -n 's/^__AGENTSMESH_SECRET__//p' |
    tr -d '\r\n'
}

restore_secret_file() {
  local secret_name="$1" secret_key="$2" destination="$3" encoded temporary
  encoded="$(cluster_secret_data "${secret_name}" "${secret_key}")"
  [[ -n "${encoded}" ]] || return 0
  temporary="$(mktemp "${destination}.XXXXXX")"
  printf '%s' "${encoded}" | base64 -d > "${temporary}"
  mv "${temporary}" "${destination}"
  echo "==> restored ${secret_name}/${secret_key} from cluster"
}

decode_secret_value() {
  printf '%s' "$1" | base64 -d
}

restore_app_secret_env() {
  local db_password jwt_secret internal_secret minio_password
  db_password="$(cluster_secret_data agentsmesh-secrets DB_PASSWORD)"
  jwt_secret="$(cluster_secret_data agentsmesh-secrets JWT_SECRET)"
  internal_secret="$(cluster_secret_data agentsmesh-secrets INTERNAL_API_SECRET)"
  minio_password="$(cluster_secret_data agentsmesh-secrets MINIO_ROOT_PASSWORD)"

  if [[ -z "${db_password}${jwt_secret}${internal_secret}${minio_password}" ]]; then
    return 0
  fi
  [[ -n "${db_password}" && -n "${jwt_secret}" &&
      -n "${internal_secret}" && -n "${minio_password}" ]] || {
    echo "agentsmesh-secrets is incomplete" >&2
    return 1
  }
  {
    echo "DB_PASSWORD=$(decode_secret_value "${db_password}")"
    echo "JWT_SECRET=$(decode_secret_value "${jwt_secret}")"
    echo "INTERNAL_API_SECRET=$(decode_secret_value "${internal_secret}")"
    echo "MINIO_ROOT_PASSWORD=$(decode_secret_value "${minio_password}")"
  } > "${GEN}/env"
  chmod 600 "${GEN}/env"
  echo "==> restored application secrets from cluster"
}

generate_cluster_secrets() {
  mkdir -p "${SEC}"
  restore_secret_file agentsmesh-pki-ca ca.crt "${GEN}/ca.crt"
  restore_secret_file agentsmesh-pki-ca ca.key "${GEN}/ca.key"
  [[ -f "${GEN}/ca.crt" && -f "${GEN}/ca.key" ]] || {
    echo "==> generating runner mTLS CA"
    openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:prime256v1 -out "${GEN}/ca.key"
    openssl req -x509 -new -key "${GEN}/ca.key" -days 3650 -out "${GEN}/ca.crt" \
      -subj "/CN=AgentsMesh Runner CA/O=agentsmesh"
  }
  openssl x509 -in "${GEN}/ca.crt" -noout
  openssl pkey -in "${GEN}/ca.key" -noout
  local ca_cert_digest ca_key_digest
  ca_cert_digest="$(openssl x509 -in "${GEN}/ca.crt" -pubkey -noout |
    openssl pkey -pubin -outform DER | openssl sha256)"
  ca_key_digest="$(openssl pkey -in "${GEN}/ca.key" -pubout -outform DER | openssl sha256)"
  [[ "${ca_cert_digest}" == "${ca_key_digest}" ]] || {
    echo "runner mTLS CA certificate does not match private key" >&2
    return 1
  }
  restore_app_secret_env
  [[ -f "${GEN}/env" ]] || {
    echo "==> generating app secrets"
    {
      echo "DB_PASSWORD=$(openssl rand -hex 16)"
      echo "JWT_SECRET=$(openssl rand -hex 32)"
      echo "INTERNAL_API_SECRET=$(openssl rand -hex 24)"
      echo "MINIO_ROOT_PASSWORD=$(openssl rand -hex 16)"
    } > "${GEN}/env"
    chmod 600 "${GEN}/env"
  }
  restore_secret_file agentsmesh-access-token private.pem "${GEN}/access-token-private.pem"
  restore_secret_file agentsmesh-access-token public.pem "${GEN}/access-token-public.pem"
  [[ -f "${GEN}/access-token-private.pem" ]] || {
    echo "==> generating access token RSA key pair"
    openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 \
      -out "${GEN}/access-token-private.pem"
  }
  openssl pkey -in "${GEN}/access-token-private.pem" -noout
  [[ -f "${GEN}/access-token-public.pem" ]] || {
    openssl pkey -in "${GEN}/access-token-private.pem" -pubout \
      -out "${GEN}/access-token-public.pem"
  }
  openssl pkey -pubin -in "${GEN}/access-token-public.pem" -noout
  local private_public_digest public_digest
  private_public_digest="$(openssl pkey -in "${GEN}/access-token-private.pem" -pubout -outform DER | openssl sha256)"
  public_digest="$(openssl pkey -pubin -in "${GEN}/access-token-public.pem" -outform DER | openssl sha256)"
  [[ "${private_public_digest}" == "${public_digest}" ]] || {
    echo "access token public key does not match private key" >&2
    return 1
  }
  ACCESS_TOKEN_KEY_ID="oilan-$(printf '%s' "${public_digest}" | awk '{print substr($2,1,16)}')"
  # shellcheck disable=SC1090
  source "${GEN}/env"
  MARKETPLACE_DATABASE_URL="postgres://agentsmesh:${DB_PASSWORD}@postgres:5432/agentsmesh?sslmode=disable"
  MARKETPLACE_MIGRATION_DATABASE_URL="${MARKETPLACE_DATABASE_URL}&x-migrations-table=marketplace_schema_migrations"

  kubectl create secret generic agentsmesh-secrets -n "${NS}" \
    --from-literal=DB_PASSWORD="${DB_PASSWORD}" \
    --from-literal=JWT_SECRET="${JWT_SECRET}" \
    --from-literal=ACCESS_TOKEN_KEY_ID="${ACCESS_TOKEN_KEY_ID}" \
    --from-literal=MARKETPLACE_DATABASE_URL="${MARKETPLACE_DATABASE_URL}" \
    --from-literal=MARKETPLACE_MIGRATION_DATABASE_URL="${MARKETPLACE_MIGRATION_DATABASE_URL}" \
    --from-literal=INTERNAL_API_SECRET="${INTERNAL_API_SECRET}" \
    --from-literal=MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD}" \
    --from-literal=STORAGE_SECRET_KEY="${MINIO_ROOT_PASSWORD}" \
    --dry-run=client -o yaml > "${SEC}/agentsmesh-secrets.yaml"

  kubectl create secret generic agentsmesh-pki-ca -n "${NS}" \
    --from-file=ca.crt="${GEN}/ca.crt" --from-file=ca.key="${GEN}/ca.key" \
    --dry-run=client -o yaml > "${SEC}/agentsmesh-pki-ca.yaml"

  kubectl create secret generic agentsmesh-access-token -n "${NS}" \
    --from-file=private.pem="${GEN}/access-token-private.pem" \
    --from-file=public.pem="${GEN}/access-token-public.pem" \
    --dry-run=client -o yaml > "${SEC}/agentsmesh-access-token.yaml"

  local store cred u p
  store="$(python3 -c "import json,os;print(json.load(open(os.path.expanduser('~/.docker/config.json'))).get('credsStore',''))")"
  cred="$(echo "${REG}" | "docker-credential-${store}" get)"
  u="$(echo "${cred}" | python3 -c "import sys,json;print(json.load(sys.stdin)['Username'])")"
  p="$(echo "${cred}" | python3 -c "import sys,json;print(json.load(sys.stdin)['Secret'])")"
  kubectl create secret docker-registry agentsmesh-regcred -n "${NS}" \
    --docker-server="${REG}" --docker-username="${u}" --docker-password="${p}" \
    --dry-run=client -o yaml > "${SEC}/agentsmesh-regcred.yaml"
}

# shellcheck shell=bash

umask 077

cluster_secret_data() {
  local secret_name="$1" secret_key="$2" escaped_key
  escaped_key="${secret_key//./\\.}"
  doops -session "${SESSION}" exec --target "${TARGET}" \
    --cmd "printf '__AGENTCLOUD_SECRET__'; kubectl -n ${NS} get secret ${secret_name} --ignore-not-found -o jsonpath='{.data.${escaped_key}}'" |
    sed -n 's/^__AGENTCLOUD_SECRET__//p' |
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

encode_secret_value() {
  printf '%s' "$1" | base64 | tr -d '\r\n'
}

encode_secret_file() {
  base64 < "$1" | tr -d '\r\n'
}

write_encoded_secret_manifest() {
  local destination="$1" name="$2" type="$3" key encoded
  shift 3
  {
    printf 'apiVersion: v1\nkind: Secret\nmetadata:\n'
    printf '  name: %s\n  namespace: %s\ntype: %s\ndata:\n' "${name}" "${NS}" "${type}"
    while (( $# > 0 )); do
      key="$1"
      encoded="$2"
      shift 2
      printf '  %s: %s\n' "${key}" "${encoded}"
    done
  } > "${destination}"
  chmod 600 "${destination}"
}

docker_config_json() {
  jq -c --arg registry "$1" '
    . as $credential |
    (($credential.Username + ":" + $credential.Secret) | @base64) as $auth |
    {auths: {
      ($registry): {
        username: $credential.Username,
        password: $credential.Secret,
        auth: $auth
      }
    }}
  '
}

restore_app_secret_env() {
  local db_password jwt_secret internal_secret minio_password
  db_password="$(cluster_secret_data agentcloud-secrets DB_PASSWORD)"
  jwt_secret="$(cluster_secret_data agentcloud-secrets JWT_SECRET)"
  internal_secret="$(cluster_secret_data agentcloud-secrets INTERNAL_API_SECRET)"
  minio_password="$(cluster_secret_data agentcloud-secrets MINIO_ROOT_PASSWORD)"

  if [[ -z "${db_password}${jwt_secret}${internal_secret}${minio_password}" ]]; then
    return 0
  fi
  [[ -n "${db_password}" && -n "${jwt_secret}" &&
      -n "${internal_secret}" && -n "${minio_password}" ]] || {
    echo "agentcloud-secrets is incomplete" >&2
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
  restore_secret_file agentcloud-pki-ca ca.crt "${GEN}/ca.crt"
  restore_secret_file agentcloud-pki-ca ca.key "${GEN}/ca.key"
  [[ -f "${GEN}/ca.crt" && -f "${GEN}/ca.key" ]] || {
    echo "==> generating runner mTLS CA"
    openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:prime256v1 -out "${GEN}/ca.key"
    openssl req -x509 -new -key "${GEN}/ca.key" -days 3650 -out "${GEN}/ca.crt" \
      -subj "/CN=Agent Cloud Runner CA/O=agentcloud"
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
  restore_secret_file agentcloud-access-token private.pem "${GEN}/access-token-private.pem"
  restore_secret_file agentcloud-access-token public.pem "${GEN}/access-token-public.pem"
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
  MARKETPLACE_DATABASE_URL="postgres://agentcloud:${DB_PASSWORD}@postgres:5432/agentcloud?sslmode=disable"
  MARKETPLACE_MIGRATION_DATABASE_URL="${MARKETPLACE_DATABASE_URL}&x-migrations-table=marketplace_schema_migrations"

  write_encoded_secret_manifest "${SEC}/agentcloud-secrets.yaml" agentcloud-secrets Opaque \
    DB_PASSWORD "$(encode_secret_value "${DB_PASSWORD}")" \
    JWT_SECRET "$(encode_secret_value "${JWT_SECRET}")" \
    ACCESS_TOKEN_KEY_ID "$(encode_secret_value "${ACCESS_TOKEN_KEY_ID}")" \
    MARKETPLACE_DATABASE_URL "$(encode_secret_value "${MARKETPLACE_DATABASE_URL}")" \
    MARKETPLACE_MIGRATION_DATABASE_URL "$(encode_secret_value "${MARKETPLACE_MIGRATION_DATABASE_URL}")" \
    INTERNAL_API_SECRET "$(encode_secret_value "${INTERNAL_API_SECRET}")" \
    MINIO_ROOT_PASSWORD "$(encode_secret_value "${MINIO_ROOT_PASSWORD}")" \
    STORAGE_SECRET_KEY "$(encode_secret_value "${MINIO_ROOT_PASSWORD}")"

  write_encoded_secret_manifest "${SEC}/agentcloud-pki-ca.yaml" agentcloud-pki-ca Opaque \
    ca.crt "$(encode_secret_file "${GEN}/ca.crt")" \
    ca.key "$(encode_secret_file "${GEN}/ca.key")"

  write_encoded_secret_manifest "${SEC}/agentcloud-access-token.yaml" agentcloud-access-token Opaque \
    private.pem "$(encode_secret_file "${GEN}/access-token-private.pem")" \
    public.pem "$(encode_secret_file "${GEN}/access-token-public.pem")"

  local store cred docker_config
  store="$(jq -r '.credsStore // empty' "${HOME}/.docker/config.json")"
  [[ -n "${store}" ]] || {
    echo "docker credential store is not configured" >&2
    return 1
  }
  cred="$(echo "${REG}" | "docker-credential-${store}" get)"
  docker_config="$(printf '%s' "${cred}" | docker_config_json "${REG}")"
  write_encoded_secret_manifest "${SEC}/agentcloud-regcred.yaml" agentcloud-regcred \
    kubernetes.io/dockerconfigjson .dockerconfigjson "$(encode_secret_value "${docker_config}")"
}

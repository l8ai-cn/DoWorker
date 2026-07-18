#!/usr/bin/env bash
set -euo pipefail

DOOPS_SESSION="${DOOPS_SESSION:?DOOPS_SESSION is required}"
DOOPS_TARGET="${DOOPS_TARGET:-gw-oilan-node}"
MINUTES=120

read -r -d '' remote_command <<EOF || true
set -euo pipefail
desired=${MINUTES}
password="\$(kubectl -n harbor-system get secret harbor-core \
  -o jsonpath='{.data.HARBOR_ADMIN_PASSWORD}' | base64 -d)"
current="\$(curl -fsS -u "admin:\${password}" \
  https://repo.aiedulab.cn:8443/api/v2.0/configurations \
  | jq -er '.token_expiration.value')"
if (( current < desired )); then
  payload="\$(jq -nc --argjson value "\${desired}" '{token_expiration: \$value}')"
  curl -fsS -u "admin:\${password}" \
    -X PUT \
    -H 'Content-Type: application/json' \
    -d "\${payload}" \
    https://repo.aiedulab.cn:8443/api/v2.0/configurations \
    >/dev/null
fi
observed="\$(curl -fsS -u "admin:\${password}" \
  https://repo.aiedulab.cn:8443/api/v2.0/configurations \
  | jq -er '.token_expiration.value')"
unset password
(( observed >= desired )) || {
  echo "Harbor token expiration remains below \${desired} minutes: \${observed}" >&2
  exit 1
}
printf 'Harbor upload token lifetime: %s minutes\n' "\${observed}"
EOF

doops -session "${DOOPS_SESSION}" exec \
  -target "${DOOPS_TARGET}" \
  -cmd "${remote_command}"

#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
SCRIPT="configure-harbor-upload-token.sh"

[[ -x "${SCRIPT}" ]]
grep -Fq 'DOOPS_SESSION="${DOOPS_SESSION:?DOOPS_SESSION is required}"' "${SCRIPT}"
grep -Fq 'DOOPS_TARGET="${DOOPS_TARGET:-gw-oilan-node}"' "${SCRIPT}"
grep -Fq 'MINUTES=120' "${SCRIPT}"
grep -Fq 'get secret harbor-core' "${SCRIPT}"
grep -Fq '/api/v2.0/configurations' "${SCRIPT}"
grep -Fq 'token_expiration' "${SCRIPT}"
grep -Fq 'unset password' "${SCRIPT}"
grep -Fq 'doops -session "${DOOPS_SESSION}" exec' "${SCRIPT}"

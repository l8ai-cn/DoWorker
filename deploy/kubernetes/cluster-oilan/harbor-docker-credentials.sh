#!/usr/bin/env bash

harbor_creds() {
  local registry="${1:-${REG:?registry required}}"
  local config="${DOCKER_CONFIG:-${HOME}/.docker}/config.json"
  local mode value decoded username password
  IFS=$'\t' read -r mode value < <(
    HARBOR_REGISTRY="${registry}" python3 - "${config}" <<'PY'
import json
import os
import sys

def normalized(value):
    value = value.removeprefix("https://").removeprefix("http://")
    return value.removesuffix("/v1/").rstrip("/")

with open(sys.argv[1], encoding="utf-8") as handle:
    config = json.load(handle)

registry = normalized(os.environ["HARBOR_REGISTRY"])
for server, helper in config.get("credHelpers", {}).items():
    if normalized(server) == registry:
        print(f"helper\t{helper}")
        raise SystemExit

store = config.get("credsStore")
if store:
    print(f"helper\t{store}")
    raise SystemExit

for server, credential in config.get("auths", {}).items():
    if normalized(server) == registry and credential.get("auth"):
        print(f"auth\t{credential['auth']}")
        raise SystemExit

raise SystemExit(f"no Docker credential found for {registry}")
PY
  ) || return 1
  if [[ "${mode}" == "helper" ]]; then
    printf '%s' "${registry}" | "docker-credential-${value}" get
    return
  fi
  [[ "${mode}" == "auth" ]] || return 1
  decoded="$(
    CREDENTIAL_AUTH="${value}" python3 - <<'PY'
import base64
import os

print(base64.b64decode(os.environ["CREDENTIAL_AUTH"]).decode(), end="")
PY
  )" || return 1
  username="${decoded%%:*}"
  password="${decoded#*:}"
  [[ "${decoded}" == *:* && -n "${username}" ]] || return 1
  jq -cn --arg username "${username}" --arg password "${password}" \
    '{Username: $username, Secret: $password}'
}

harbor_curl() {
  if [[ -n "${HARBOR_CA_CERT:-}" ]]; then
    [[ -f "${HARBOR_CA_CERT}" ]] || {
      echo "Harbor CA certificate not found: ${HARBOR_CA_CERT}" >&2
      return 1
    }
    curl --cacert "${HARBOR_CA_CERT}" "$@"
    return
  fi
  curl "$@"
}

#!/usr/bin/env bash

# shellcheck disable=SC1091
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/harbor-docker-credentials.sh"

harbor_load_credentials() {
  local registry="${1:?registry required}" credentials
  credentials="$(harbor_creds "${registry}")" || return 1
  HARBOR_USERNAME="$(jq -er '.Username' <<<"${credentials}")" || return 1
  HARBOR_PASSWORD="$(jq -er '.Secret' <<<"${credentials}")" || return 1
}

harbor_immutable_rules() {
  local registry="${1:?registry required}" project="${2:?project required}"
  harbor_curl -fsS \
    --connect-timeout 10 \
    --max-time 30 \
    --retry 3 \
    --retry-all-errors \
    -u "${HARBOR_USERNAME}:${HARBOR_PASSWORD}" \
    "https://${registry}/api/v2.0/projects/${project}/immutabletagrules"
}

harbor_registry_token() {
  local registry="${1:?registry required}"
  harbor_curl -fsS \
    --connect-timeout 10 \
    --max-time 30 \
    --retry 3 \
    --retry-all-errors \
    -u "${HARBOR_USERNAME}:${HARBOR_PASSWORD}" \
    --get \
    --data-urlencode "service=harbor-registry" \
    --data-urlencode "scope=repository:agentcloud/release-preflight:push,pull" \
    "https://${registry}/service/token" \
    | jq -er '.token'
}

harbor_token_lifetime_minutes() {
  local token="${1:?token required}"
  HARBOR_TOKEN="${token}" python3 - <<'PY'
import base64
import json
import os

payload = os.environ["HARBOR_TOKEN"].split(".")[1]
payload += "=" * (-len(payload) % 4)
claims = json.loads(base64.urlsafe_b64decode(payload))
issued_at = claims["iat"]
expires_at = claims["exp"]
if not isinstance(issued_at, int) or not isinstance(expires_at, int) or expires_at <= issued_at:
    raise SystemExit("invalid Harbor token timestamps")
print((expires_at - issued_at) // 60)
PY
}

harbor_require_upload_token_expiration() {
  local registry="${1:?registry required}" minimum="${2:?minimum required}"
  local HARBOR_USERNAME HARBOR_PASSWORD token lifetime
  harbor_load_credentials "${registry}"
  token="$(harbor_registry_token "${registry}")"
  lifetime="$(harbor_token_lifetime_minutes "${token}")"
  [[ "${lifetime}" =~ ^[0-9]+$ && "${lifetime}" -ge "${minimum}" ]] || {
    echo "Harbor upload token lifetime is ${lifetime} minutes; require at least ${minimum}" >&2
    echo "run configure-harbor-upload-token.sh before publishing images" >&2
    return 1
  }
}

harbor_has_immutable_tag() {
  local rules="$1" repository="$2" tag="$3"
  jq -e --arg repository "${repository}" --arg tag "${tag}" '
    any(.[];
      .disabled != true and
      any(.tag_selectors[]?; .decoration == "matches" and .pattern == $tag) and
      any(.scope_selectors.repository[]?;
        .decoration == "repoMatches" and .pattern == $repository
      )
    )
  ' <<<"${rules}" >/dev/null
}

harbor_ensure_immutable_tag() {
  local registry="$1" project="$2" repository="$3" tag="$4" rules payload
  local HARBOR_USERNAME HARBOR_PASSWORD
  harbor_load_credentials "${registry}"
  rules="$(harbor_immutable_rules "${registry}" "${project}")"
  if harbor_has_immutable_tag "${rules}" "${repository}" "${tag}"; then
    return
  fi
  payload="$(jq -nc --arg repository "${repository}" --arg tag "${tag}" '{
    disabled: false,
    action: "immutable",
    template: "immutable_template",
    tag_selectors: [{
      kind: "doublestar",
      decoration: "matches",
      pattern: $tag
    }],
    scope_selectors: {
      repository: [{
        kind: "doublestar",
        decoration: "repoMatches",
        pattern: $repository
      }]
    }
  }')"
  harbor_curl -fsS \
    --connect-timeout 10 \
    --max-time 30 \
    --retry 3 \
    --retry-all-errors \
    -u "${HARBOR_USERNAME}:${HARBOR_PASSWORD}" \
    -H "Content-Type: application/json" \
    -d "${payload}" \
    "https://${registry}/api/v2.0/projects/${project}/immutabletagrules" \
    >/dev/null
  rules="$(harbor_immutable_rules "${registry}" "${project}")"
  harbor_has_immutable_tag "${rules}" "${repository}" "${tag}"
}

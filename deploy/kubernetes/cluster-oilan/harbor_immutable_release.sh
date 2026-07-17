#!/usr/bin/env bash

harbor_load_credentials() {
  local registry="${1:?registry required}" config helper credentials auth
  config="${DOCKER_CONFIG:-${HOME}/.docker}/config.json"
  helper="$(python3 - "${config}" "${registry}" <<'PY'
import json
import pathlib
import sys

config = json.loads(pathlib.Path(sys.argv[1]).read_text())
print(config.get("credHelpers", {}).get(sys.argv[2], config.get("credsStore", "")))
PY
)"
  if [[ -n "${helper}" ]]; then
    credentials="$(printf '%s' "${registry}" | "docker-credential-${helper}" get)"
    HARBOR_USERNAME="$(printf '%s' "${credentials}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["Username"])')"
    HARBOR_PASSWORD="$(printf '%s' "${credentials}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["Secret"])')"
    return
  fi
  auth="$(python3 - "${config}" "${registry}" <<'PY'
import base64
import json
import pathlib
import sys

config = json.loads(pathlib.Path(sys.argv[1]).read_text())
print(base64.b64decode(config["auths"][sys.argv[2]]["auth"]).decode())
PY
)"
  HARBOR_USERNAME="${auth%%:*}"
  HARBOR_PASSWORD="${auth#*:}"
}

harbor_immutable_rules() {
  local registry="${1:?registry required}" project="${2:?project required}"
  curl -fsS \
    --connect-timeout 10 \
    --max-time 30 \
    --retry 3 \
    --retry-all-errors \
    -u "${HARBOR_USERNAME}:${HARBOR_PASSWORD}" \
    "https://${registry}/api/v2.0/projects/${project}/immutabletagrules"
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
  curl -fsS \
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

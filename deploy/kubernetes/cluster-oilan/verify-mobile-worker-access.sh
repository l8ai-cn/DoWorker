#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${MOBILE_SMOKE_API_BASE_URL:-https://dowork.l8ai.cn}"
MOBILE_BASE_URL="${MOBILE_SMOKE_MOBILE_BASE_URL:-https://mobile.l8ai.cn}"
ORG_SLUG="${MOBILE_SMOKE_ORG_SLUG:-dev-org}"
AGENT_ID="${MOBILE_SMOKE_AGENT_ID:-codex-cli}"
RUN_INTERACTIONS="${MOBILE_SMOKE_RUN_INTERACTIONS:-false}"
USERNAME="${MOBILE_SMOKE_USERNAME:?MOBILE_SMOKE_USERNAME is required}"
PASSWORD="${MOBILE_SMOKE_PASSWORD:?MOBILE_SMOKE_PASSWORD is required}"

require_command() {
  command -v "$1" >/dev/null || {
    printf 'required command not found: %s\n' "$1" >&2
    exit 1
  }
}

require_command curl
require_command jq

login_payload="$(curl -fsS -X POST "${API_BASE_URL}/auth/login" \
  -H 'Content-Type: application/json' \
  --data "$(jq -n --arg username "$USERNAME" --arg password "$PASSWORD" \
    '{username: $username, password: $password}')")"
TOKEN="$(printf '%s' "$login_payload" | jq -er '.token')"

api() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local args=(
    -fsS
    -X "$method"
    "${API_BASE_URL}${path}"
    -H "Authorization: Bearer ${TOKEN}"
    -H "X-Organization-Slug: ${ORG_SLUG}"
  )
  if [[ -n "$body" ]]; then
    args+=(-H 'Content-Type: application/json' --data "$body")
  fi
  curl "${args[@]}"
}

assert_status() {
  local expected="$1"
  local url="$2"
  local actual
  actual="$(curl -sS -o /dev/null -w '%{http_code}' "$url")"
  [[ "$actual" == "$expected" ]] || {
    printf 'unexpected status for %s: expected %s, got %s\n' "$url" "$expected" "$actual" >&2
    exit 1
  }
}

assert_status 200 "${MOBILE_BASE_URL}/"
assert_status 200 "${MOBILE_BASE_URL}/login"

agents="$(api GET '/v1/agents')"
printf '%s' "$agents" | jq -e --arg agent "$AGENT_ID" '
  .data[]
  | select(.id == $agent)
  | (.supported_modes
      | if type == "array" then (sort == ["acp", "pty"]) else false end)
    and (.requires_model_resource == true)
' >/dev/null || {
  printf 'worker %s does not expose the required ACP/PTY and model-resource contract\n' "$AGENT_ID" >&2
  exit 1
}

model_resource_id="$(api GET '/v1/model-resources' | jq -er '
  [.data[] | select(.is_default == true and .id > 0)]
  | if length == 1 then .[0].id else error("exactly one default model resource is required") end
')"

printf 'contract smoke passed: mobile=%s api=%s org=%s worker=%s model_resource=%s\n' \
  "$MOBILE_BASE_URL" "$API_BASE_URL" "$ORG_SLUG" "$AGENT_ID" "$model_resource_id"

[[ "$RUN_INTERACTIONS" == "true" ]] || exit 0

acp_session_id=""
pty_session_id=""
cleanup() {
  set +e
  for session_id in "$acp_session_id" "$pty_session_id"; do
    [[ -n "$session_id" ]] || continue
    status="$(curl -sS -o /dev/null -w '%{http_code}' -X DELETE \
      "${API_BASE_URL}/v1/sessions/${session_id}" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "X-Organization-Slug: ${ORG_SLUG}")"
    if [[ "$status" != "204" && "$status" != "404" ]]; then
      printf 'cleanup failed for session %s: HTTP %s\n' "$session_id" "$status" >&2
    fi
  done
}
trap cleanup EXIT

marker="MOBILE_RELEASE_SMOKE_$(date +%s)_$$"
acp_body="$(jq -n \
  --arg agent "$AGENT_ID" \
  --arg marker "$marker" \
  --argjson model_resource_id "$model_resource_id" \
  '{
    agent_id: $agent,
    model_resource_id: $model_resource_id,
    title: "mobile-acp-release-smoke",
    initial_items: [{
      type: "message",
      data: {
        role: "user",
        content: [{type: "input_text", text: ("Reply with exactly " + $marker + ".")}]
      }
    }]
  }')"
acp_session_id="$(api POST '/v1/sessions' "$acp_body" | jq -er '.id')"

for attempt in $(seq 1 90); do
  items="$(api GET "/v1/sessions/${acp_session_id}/items?limit=100&order=desc")"
  if printf '%s' "$items" | jq -e --arg marker "$marker" \
    '[.data[]? | tostring | select(contains($marker))] | length > 0' >/dev/null; then
    break
  fi
  [[ "$attempt" -lt 90 ]] || {
    printf 'ACP response marker was not observed for session %s\n' "$acp_session_id" >&2
    exit 1
  }
  sleep 1
done

pty_body="$(jq -n \
  --arg agent "$AGENT_ID" \
  --argjson model_resource_id "$model_resource_id" \
  '{
    agent_id: $agent,
    model_resource_id: $model_resource_id,
    title: "mobile-pty-release-smoke",
    pty_only: true
  }')"
pty_session_id="$(api POST '/v1/sessions' "$pty_body" | jq -er '.id')"

relay_file="$(mktemp)"
trap 'rm -f "$relay_file"; cleanup' EXIT
for attempt in $(seq 1 30); do
  relay_status="$(curl -sS -o "$relay_file" -w '%{http_code}' \
    "${API_BASE_URL}/v1/sessions/${pty_session_id}/relay-connection" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "X-Organization-Slug: ${ORG_SLUG}")"
  if [[ "$relay_status" == "200" ]]; then
    jq -e '.relay_url and .token and .pod_key' "$relay_file" >/dev/null
    break
  fi
  [[ "$attempt" -lt 30 ]] || {
    printf 'PTY relay connection did not become available: HTTP %s\n' "$relay_status" >&2
    exit 1
  }
  sleep 1
done

acp_relay_status="$(curl -sS -o /dev/null -w '%{http_code}' \
  "${API_BASE_URL}/v1/sessions/${acp_session_id}/relay-connection" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-Organization-Slug: ${ORG_SLUG}")"
[[ "$acp_relay_status" == "400" ]] || {
  printf 'ACP relay connection must be rejected, got HTTP %s\n' "$acp_relay_status" >&2
  exit 1
}

printf 'interaction smoke passed: ACP reply, PTY relay token, and ACP relay rejection verified\n'

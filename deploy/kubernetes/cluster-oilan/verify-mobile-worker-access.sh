#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
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
require_command node

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

pod_connection() {
  local pod_key="$1"
  curl -fsS -X POST "${API_BASE_URL}/proto.pod.v1.PodService/GetPodConnection" \
    -H 'Connect-Protocol-Version: 1' \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "X-Organization-Slug: ${ORG_SLUG}" \
    --data "$(jq -n --arg org_slug "$ORG_SLUG" --arg pod_key "$pod_key" \
      '{orgSlug: $org_slug, podKey: $pod_key}')"
}

wait_for_pod_key() {
  local session_id="$1"
  local pod_key=""
  for attempt in $(seq 1 90); do
    pod_key="$(api GET "/v1/sessions/${session_id}" | jq -r '.pod_key // empty')"
    [[ -n "$pod_key" ]] && {
      printf '%s' "$pod_key"
      return
    }
    sleep 1
  done
  printf 'Pod key was not assigned to session %s\n' "$session_id" >&2
  exit 1
}

wait_for_pod_connection() {
  local pod_key="$1"
  local response=""
  for attempt in $(seq 1 90); do
    if response="$(pod_connection "$pod_key")"; then
      printf '%s' "$response" | jq -e '.relayUrl and .token and .podKey' >/dev/null
      printf '%s' "$response"
      return
    fi
    sleep 1
  done
  printf 'GetPodConnection did not become available for Worker %s\n' "$pod_key" >&2
  exit 1
}

run_relay_smoke() {
  local mode="$1"
  local connection="$2"
  local marker="${3:-}"
  printf '%s' "$connection" |
    jq -c --arg mode "$mode" --arg marker "$marker" \
      '{mode: $mode, marker: $marker, relayUrl: .relayUrl, token: .token}' |
    node --experimental-websocket "${ROOT}/mobile-relay-data-plane-smoke.mjs"
  printf '%s Relay data-plane smoke passed\n' "$mode"
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

acp_body="$(jq -n \
  --arg agent "$AGENT_ID" \
  --argjson model_resource_id "$model_resource_id" \
  '{
    agent_id: $agent,
    model_resource_id: $model_resource_id,
    title: "mobile-acp-release-smoke"
  }')"
acp_session_id="$(api POST '/v1/sessions' "$acp_body" | jq -er '.id')"
acp_pod_key="$(wait_for_pod_key "$acp_session_id")"
acp_connection="$(wait_for_pod_connection "$acp_pod_key")"
run_relay_smoke acp "$acp_connection" "MOBILE_ACP_RELAY_$(date +%s)_$$"

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
pty_pod_key="$(wait_for_pod_key "$pty_session_id")"
pty_connection="$(wait_for_pod_connection "$pty_pod_key")"
run_relay_smoke pty "$pty_connection"

printf 'interaction smoke passed: direct ACP reply and PTY Relay control verified\n'

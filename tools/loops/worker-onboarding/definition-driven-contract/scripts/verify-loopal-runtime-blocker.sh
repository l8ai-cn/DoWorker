#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(git -C "$LOOP_ROOT" rev-parse --show-toplevel)"
ENV_FILE="${REPO_ROOT}/deploy/dev/.env"
PREPARE="${REPO_ROOT}/docker/agent-runtime/prepare_binaries.sh"
LOCAL_BINARY="${REPO_ROOT}/deploy/dev/loopal-binary"

[[ -f "$ENV_FILE" ]]
[[ -x "$PREPARE" ]]
project="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' "$ENV_FILE")"
[[ -n "$project" ]]

if docker image inspect "${project}-runner-loopal:latest" >/dev/null 2>&1; then
  echo "Loopal image now exists; replace blocker evidence with image verification" >&2
  exit 1
fi

docker compose --env-file "$ENV_FILE" \
  -f "${REPO_ROOT}/deploy/dev/docker-compose.yml" \
  -f "${REPO_ROOT}/deploy/dev/docker-compose.runners.yml" \
  config --format json \
  | jq -e '.services["runner-loopal"].build.args.AGENT_RUNTIME == "loopal"' >/dev/null

[[ -x "$LOCAL_BINARY" ]]
git -C "$REPO_ROOT" check-ignore -q "deploy/dev/loopal-binary"
grep -aFq "runner/internal/agents/mockagent" "$LOCAL_BINARY"

expect_rejection() {
  local expected="$1" binary="$2"
  local staging log rc
  staging="$(mktemp -d)"
  log="$(mktemp)"
  set +e
  LOOPAL_BINARY="$binary" "$PREPARE" "$staging" loopal >"$log" 2>&1
  rc=$?
  set -e
  rm -rf "$staging"
  grep -Fq "$expected" "$log"
  rm -f "$log"
  [[ "$rc" -eq 1 ]]
}

expect_rejection \
  "loopal requires LOOPAL_BINARY to point to a real Loopal CLI artifact" \
  ""
expect_rejection \
  "loopal artifact is an E2E mock binary, not a real Loopal CLI" \
  "$LOCAL_BINARY"

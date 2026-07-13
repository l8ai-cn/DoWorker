#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -n "${TEST_SRCDIR:-}" ]]; then
  REPO_ROOT="${TEST_SRCDIR}/${TEST_WORKSPACE}"
else
  REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
fi

API="${SESSION_COMPAT_API_URL:-http://localhost:10015}"

preflight_cleanup_pods() {
  # Real terminate first: frees the runner's in-memory pod slots
  # (max_concurrent_pods) — a DB UPDATE alone leaks them until 503.
  node "${REPO_ROOT}/tests/session-compat-smoke/terminate-all-pods.mjs" || true

  local pg_container
  if [[ -f "${SCRIPT_DIR}/.env" ]]; then
    # shellcheck disable=SC1091
    source "${SCRIPT_DIR}/.env"
  fi
  pg_container="${COMPOSE_PROJECT_NAME:-agentsmesh-main}-postgres-1"
  if ! docker ps --format '{{.Names}}' | grep -qx "$pg_container"; then
    return 0
  fi
  # DB fallback for rows the API path can't reach (runner offline, stale rows).
  docker exec "$pg_container" psql -U agentsmesh -d agentsmesh -c \
    "UPDATE pods SET status='terminated', agent_status='idle'
     WHERE organization_id=(SELECT id FROM organizations WHERE slug='dev-org')
       AND status NOT IN ('terminated','completed','orphaned','error');" \
    >/dev/null 2>&1 || true
}

# golang-migrate keeps exactly one row in schema_migrations; a second row or
# dirty=t means someone applied schema manually and `migrate up` will refuse
# to run — fail fast with a fix hint instead of debugging opaque SQL errors.
preflight_check_migrations() {
  local pg_container rows
  pg_container="${COMPOSE_PROJECT_NAME:-agentsmesh-main}-postgres-1"
  if ! docker ps --format '{{.Names}}' | grep -qx "$pg_container"; then
    return 0
  fi
  rows="$(docker exec "$pg_container" psql -U agentsmesh -d agentsmesh -tAc \
    'SELECT count(*), bool_or(dirty) FROM schema_migrations;' 2>/dev/null || true)"
  if [[ -n "$rows" && "$rows" != "1|f" ]]; then
    echo "session compatibility smoke: schema_migrations is inconsistent (count|dirty = ${rows})"
    echo "Fix: docker compose run --rm --no-deps migrate ... force <version> && ... up"
    exit 1
  fi
}

preflight_cleanup_pods
preflight_check_migrations

if ! curl -sf "${API}/health" >/dev/null 2>&1; then
  echo "session compatibility smoke: backend not reachable at ${API}"
  echo "Start dev stack: ./deploy/dev/dev.sh --backend-only"
  exit 1
fi

run_suite() {
  local label="$1"
  local script="$2"
  echo ""
  echo "=========================================="
  echo "  ${label}"
  echo "=========================================="
  node "${REPO_ROOT}/tests/session-compat-smoke/${script}"
}

run_suite "S0 message round-trip" "api-integration-smoke.mjs"
run_suite "S1 session wire + elicitation + terminal" "session-s1-smoke.mjs"
run_suite "S2 compat API" "session-s2-smoke.mjs"
run_suite "S3 platform mechanisms" "session-s3-smoke.mjs"
run_suite "S4 P4 extensions" "session-s4-smoke.mjs"
run_suite "S5 full parity" "session-s5-smoke.mjs"

echo ""
echo "session compatibility smoke: all suites passed"

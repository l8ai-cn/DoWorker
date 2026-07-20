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
  node "${REPO_ROOT}/tests/session-compat-smoke/terminate-all-pods.mjs"
}

preflight_check_migrations() {
  return 0
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

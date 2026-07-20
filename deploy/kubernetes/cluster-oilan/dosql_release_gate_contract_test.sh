#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

latest_version() {
  find "$ROOT/../../../backend/migrations" -name '*.up.sql' -exec basename {} \; |
    awk -F_ '{ print $1 }' |
    sort -n |
    tail -1
}

create_fixture() {
  local fixture_case="$1"
  local fixture_root="$TMP/$fixture_case"
  node "$ROOT/dosql_release_gate_fixture.mjs" \
    --root "$fixture_root" \
    --case "$fixture_case" \
    --target oilan-postgres \
    --mode production \
    --session dosql-contract \
    --change-id change-contract \
    --operation-id dbop-contract \
    --version "$(latest_version)"
  printf '%s\n' "$fixture_root"
}

run_verifier() {
  local fixture_root="$1"
  DOSQL_RELEASE_GATE_TEST_MODE=1 \
  node "$ROOT/dosql_release_evidence_validate.mjs" \
    --target oilan-postgres \
    --mode production \
    --session dosql-contract \
    --change-id change-contract \
    --operation-id dbop-contract \
    --expected-version "$(latest_version)" \
    --test-journal "$fixture_root/journal.jsonl" \
    --test-evidence "$fixture_root/evidence.json"
}

if DOSQL_RELEASE_DB_TARGET=oilan-postgres \
  DOSQL_RELEASE_DB_MODE=production \
  DOSQL_RELEASE_DB_SESSION=dosql-contract \
  DOSQL_RELEASE_CHANGE_ID=change-contract \
  DOSQL_RELEASE_OPERATION_ID=dbop-contract \
  DOSQL_RELEASE_MIGRATION_VERSION="$(latest_version)" \
  DOSQL_RELEASE_EVIDENCE_BIN=/bin/true \
  bash -c '
    set -euo pipefail
    DIR="$1"
    source "$1/dosql_release_gate.sh"
    require_dosql_database_evidence
  ' bash "$ROOT" >/dev/null 2>&1; then
  echo "production deploy accepted unavailable DoSql audit evidence" >&2
  exit 1
fi

run_verifier "$(create_fixture valid)" >/dev/null
for fixture_case in target-mismatch mode-mismatch session-mismatch change-mismatch stale-version missing-running fingerprint-mismatch broken-chain; do
  if run_verifier "$(create_fixture "$fixture_case")" >/dev/null 2>&1; then
    echo "invalid DoSql release audit evidence was accepted: ${fixture_case}" >&2
    exit 1
  fi
done

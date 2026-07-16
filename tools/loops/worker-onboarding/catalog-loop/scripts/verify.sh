#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$LOOP_ROOT"

test -s loop.json
test -s loops.json
test -s tasks.json
test -s agents.json
test -s context-policy.json
test -s ACCEPTANCE.md
test -s DECISIONS.md
test -s monitoring-plan.json
test -s catalog/worker-evidence-matrix.json
test -s evidence/revocations/2026-07-12-invalid-shared-contract.md

bash scripts/verify-rebuild-state.sh
bash scripts/verify-inventory.sh
bash scripts/verify-catalog-contract.sh
bash scripts/verify-worker-runs.sh --all-verified
jq -e '.status == "complete"' state.json >/dev/null
! grep -Eq -- '- \[ \]' ACCEPTANCE.md
jq -e '
  .schema_version == 1 and
  .status == "accepted" and
  .terminal_verifier == "catalog-terminal" and
  (.verified_at | type == "string" and length > 0)
' evidence/catalog-terminal.json >/dev/null

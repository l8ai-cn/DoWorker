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

bash scripts/verify-research.sh
bash scripts/verify-contract.sh
bash scripts/verify-api-contract.sh
bash scripts/verify-runner-runtime.sh
jq -e '.status == "complete"' state.json >/dev/null
! grep -Eq -- '- \[ \]' ACCEPTANCE.md
jq -e '
  .schema_version == 1 and
  .status == "accepted" and
  .terminal_verifier == "worker-terminal" and
  (.verified_at | type == "string" and length > 0)
' evidence/tests/terminal.json >/dev/null
jq -e '
  .schema_version == 1 and
  .status == "passed" and
  ([.states[]] | sort) == ["disabled", "error", "incompatible", "loading", "success"] and
  (.screenshot_path | type == "string" and length > 0)
' evidence/browser/worker-flow.json >/dev/null
jq -e '
  .schema_version == 1 and
  .status == "accepted" and
  .reviewer == "reviewer" and
  .terminal_verifier == "worker-terminal" and
  (.reviewed_at | type == "string" and length > 0)
' evidence/review/acceptance.json >/dev/null

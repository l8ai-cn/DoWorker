#!/usr/bin/env bash
set -euo pipefail

# Generated structural verifier. Replace or extend with the real deterministic
# checks for this loop, but do not weaken existing checks without human approval.
loop_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$loop_root"

test -s ACCEPTANCE.md
test -s DECISIONS.md
test -s monitoring-plan.json
grep -Eq -- '- \[[ xX]\]' ACCEPTANCE.md
if grep -Eq '"status"[[:space:]]*:[[:space:]]*"complete"' state.json; then ! grep -Eq -- '- \[ \]' ACCEPTANCE.md; fi

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORKFLOW="$ROOT/.github/workflows/ci.yml"

python3 - "$WORKFLOW" <<'PY'
import sys

workflow = sys.argv[1]
lines = open(workflow, encoding="utf-8").read().splitlines()
for index, line in enumerate(lines):
    if "Swatinem/rust-cache@v2" not in line:
        continue
    window = lines[max(0, index - 8):index]
    if not any("Seed Rust proto stubs" in candidate for candidate in window):
        print(f"rust-cache at {workflow}:{index + 1} runs before proto stubs are seeded", file=sys.stderr)
        sys.exit(1)
PY

bash "$ROOT/scripts/seed-rust-proto-stubs.sh" >/dev/null
test -f "$ROOT/clients/core/crates/proto/acp_state/src/lib.rs"

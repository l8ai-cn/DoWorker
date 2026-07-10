#!/usr/bin/env bash
# Seed empty src/lib.rs for clients/core/crates/proto/* so Cargo can load the
# workspace before do_worker_proto_gen overwrites them with real prost output.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_ROOT="$ROOT/clients/core/crates/proto"

shopt -s nullglob
for dir in "$PROTO_ROOT"/*/; do
  [[ -f "${dir}Cargo.toml" ]] || continue
  mkdir -p "${dir}src"
  if [[ ! -f "${dir}src/lib.rs" ]]; then
    printf '%s\n' '// seeded stub — overwritten by do_worker_proto_gen' > "${dir}src/lib.rs"
  fi
done
echo "seeded rust proto stubs under clients/core/crates/proto/"

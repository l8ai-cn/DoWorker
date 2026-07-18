#!/usr/bin/env bash
# Build clients/core wasm via Cargo + wasm-pack.
# Output: packages/do-worker-wasm/
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CORE="$ROOT/clients/core"
OUT="$ROOT/packages/do-worker-wasm"
STAGE="$(mktemp -d "$ROOT/packages/.do-worker-wasm-build.XXXXXX")"

export PATH="${HOME}/.local/bin:${HOME}/.cargo/bin:${PATH}"

cleanup() {
  rm -rf "$STAGE"
}

trap cleanup EXIT

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required tool not found: $1" >&2
    exit 1
  fi
}

need protoc
need wasm-pack
need rustc
need cargo

if ! rustup target list --installed 2>/dev/null | grep -qx 'wasm32-unknown-unknown'; then
  echo "adding rustup target wasm32-unknown-unknown"
  rustup target add wasm32-unknown-unknown
fi

export RUSTFLAGS="${RUSTFLAGS:-} --cfg getrandom_backend=\"wasm_js\""

cd "$CORE"

# Cargo refuses to load workspace members with no targets. Seed empty
# lib.rs so `cargo run -p do_worker_proto_gen` can start, then overwrite.
bash "$ROOT/scripts/seed-rust-proto-stubs.sh"

cargo run -p do_worker_proto_gen --bin gen-proto

cd "$CORE/crates/wasm"
wasm-pack build \
  --target web \
  --out-name wasm_pkg \
  --out-dir "$STAGE" \
  --release

# Optional binaryen pass. Safe no-op when wasm-opt is absent.
if command -v wasm-opt >/dev/null 2>&1; then
  echo "wasm-opt: optimizing $STAGE/wasm_pkg_bg.wasm"
  wasm-opt -Oz --enable-bulk-memory --enable-mutable-globals \
    -o "$STAGE/wasm_pkg_bg.wasm" "$STAGE/wasm_pkg_bg.wasm"
else
  echo "wasm-opt: not found (optional). Install binaryen for extra size wins."
fi

# wasm-pack overwrites package.json; restore published package identity.
python3 -c "
import json
from pathlib import Path
out = Path(r'''$STAGE''')
pkg = {
    'name': 'do-worker-wasm',
    'type': 'module',
    'version': '0.1.0',
    'main': 'wasm_pkg.js',
    'types': 'wasm_pkg.d.ts',
    'files': ['wasm_pkg_bg.wasm', 'wasm_pkg.js', 'wasm_pkg.d.ts'],
    'sideEffects': ['./snippets/*'],
    'private': True,
}
(out / 'package.json').write_text(json.dumps(pkg, indent=2) + '\n')
print('ok: wrote', out)
"

# Restore gitignore after wasm-pack
cat > "$STAGE/.gitignore" <<'EOF'
/wasm_pkg.js
/wasm_pkg.d.ts
/wasm_pkg_bg.wasm
/wasm_pkg_bg.wasm.d.ts
/snippets/
EOF

mkdir -p "$OUT"
mv "$STAGE/wasm_pkg_bg.wasm" "$OUT/wasm_pkg_bg.wasm"
mv "$STAGE/wasm_pkg_bg.wasm.d.ts" "$OUT/wasm_pkg_bg.wasm.d.ts"
mv "$STAGE/wasm_pkg.d.ts" "$OUT/wasm_pkg.d.ts"
mv "$STAGE/package.json" "$OUT/package.json"
mv "$STAGE/.gitignore" "$OUT/.gitignore"
mv "$STAGE/wasm_pkg.js" "$OUT/wasm_pkg.js"

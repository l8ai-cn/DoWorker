#!/usr/bin/env bash
# Verify that pre-dashboard chunks don't pull in wasm. Run after
# `pnpm run web:build` (or a local `next build` under clients/web).
# Exits non-zero if any leak.
#
# Architectural invariant:
#   Marketing routes AND the entire (auth) route group must not load the
#   wasm bundle — they go through @/lib/light-auth + light-session.
#   Only (dashboard) and popout are allowed to boot wasm.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
# Prefer production build output; allow override for custom paths.
NEXT_DIR="${NEXT_DIR:-$ROOT/clients/web/.next}"
CHUNKS_DIR="${NEXT_DIR}/static/chunks"

if [[ ! -d "${CHUNKS_DIR}" ]]; then
  echo "FAIL: ${CHUNKS_DIR} not found. Run: pnpm run web:build"
  exit 2
fi

# Pre-dashboard route chunks: marketing + (auth). Must not contain
# WasmProvider / initWasmCore / any wasm-bindgen class names.
LEAKS=$(
  find "${CHUNKS_DIR}/app" -maxdepth 4 -name "*.js" \
    -not -path "*\(dashboard\)*" \
    -not -path "*popout*" \
    -print0 \
  | xargs -0 grep -l 'WasmProvider\|initWasmCore\|WasmApiClient\|WasmAuthManager\|wasm_pkg\|do-worker-wasm' 2>/dev/null \
  || true
)

if [[ -n "${LEAKS}" ]]; then
  echo "FAIL: pre-dashboard chunks contain wasm symbols:"
  echo "${LEAKS}" | sed 's/^/  /'
  exit 1
fi

echo "PASS: no wasm symbols in marketing or (auth) chunks"

# Confirm the source dependency chain still boots wasm. Production minification
# may move WasmProvider into a shared chunk and remove the component name.
WEB_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
for layout in \
  "${WEB_ROOT}/src/app/(dashboard)/layout.tsx" \
  "${WEB_ROOT}/src/app/popout/layout.tsx"; do
  if ! grep -F 'AuthBootstrap' "${layout}" >/dev/null 2>&1; then
    echo "FAIL: ${layout} lost AuthBootstrap"
    exit 1
  fi
done
if ! grep -F 'WasmProvider' \
  "${WEB_ROOT}/src/components/auth/AuthBootstrap.tsx" >/dev/null 2>&1; then
  echo "FAIL: AuthBootstrap lost WasmProvider"
  exit 1
fi

echo "PASS: dashboard / popout layouts retain WasmProvider"

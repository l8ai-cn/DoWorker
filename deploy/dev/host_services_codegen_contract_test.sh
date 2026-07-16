#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="$ROOT/deploy/dev"

source "$SCRIPT_DIR/lib/log.sh"
source "$SCRIPT_DIR/lib/worktree.sh"
source "$SCRIPT_DIR/lib/host_services_lite.sh"

codegen_calls=0
codegen_exit=0

bash() {
  codegen_calls=$((codegen_calls + 1))
  return "$codegen_exit"
}

ensure_go_codegen
ensure_go_codegen
[[ "$codegen_calls" -eq 1 ]] || {
  echo "expected one successful codegen invocation, got $codegen_calls" >&2
  exit 1
}

unset _DEV_GO_CODEGEN_READY
codegen_calls=0
codegen_exit=1
if ensure_go_codegen >/dev/null 2>&1; then
  echo "expected failed codegen to propagate" >&2
  exit 1
fi

codegen_exit=0
ensure_go_codegen
[[ "$codegen_calls" -eq 2 ]] || {
  echo "expected failed codegen to be retried, got $codegen_calls calls" >&2
  exit 1
}

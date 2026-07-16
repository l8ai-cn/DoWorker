#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SMOKE="${ROOT}/verify-mobile-worker-access.sh"

require() {
  grep -F "$1" "$SMOKE" >/dev/null || {
    printf 'missing mobile release smoke contract: %s\n' "$1" >&2
    exit 1
  }
}

require 'GetPodConnection'
require 'mobile-relay-data-plane-smoke.mjs'
require 'run_relay_smoke acp'
require 'run_relay_smoke pty'

if grep -F '/relay-connection' "$SMOKE" >/dev/null; then
  printf 'mobile Worker release smoke must not verify the legacy session relay path\n' >&2
  exit 1
fi

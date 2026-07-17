#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

if grep -Eq 'e2e-mock-agent-binary.*(loopal|do-agent)|_write_do_agent_stub' \
  prepare_binaries.sh ../../deploy/dev/lib/build_do_agent_binary.sh; then
  echo "Worker runtime staging must not substitute mock binaries" >&2
  exit 1
fi

grep -Fq 'COPY --chmod=0755 binaries/' Dockerfile
grep -Fq 'prepare_binaries.sh" "$STAGING" "$rt"' build.sh
grep -Fq 'AGENT_RUNTIME="${2:?agent runtime required}"' prepare_binaries.sh
grep -Fq 'loopal requires LOOPAL_BINARY' prepare_binaries.sh
grep -Fq 'runner/internal/agents/mockagent' prepare_binaries.sh

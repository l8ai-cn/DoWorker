#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DESTINATION="${1:?destination directory required}"
SOURCE="${DO_AGENT_BINARY:-${REPO_ROOT}/deploy/dev/do-agent-binary}"
source "${REPO_ROOT}/docker/agent-runtime/do_agent_release_manifest.sh"
EXPECTED="$(do_agent_release_value artifact.binary_sha256)"

if [[ "${REQUIRE_DO_AGENT_BINARY:-0}" == "1" ]]; then
  [[ -n "${DO_AGENT_BINARY:-}" ]] || {
    echo "DO_AGENT_BINARY must point to the approved linux/amd64 artifact" >&2
    exit 1
  }
fi
do_agent_require_digest "artifact.binary_sha256" "${EXPECTED}"
if [[ -n "${DO_AGENT_BINARY_SHA256:-}" && "${DO_AGENT_BINARY_SHA256}" != "${EXPECTED}" ]]; then
  echo "DO_AGENT_BINARY_SHA256 does not match the trusted release manifest" >&2
  exit 1
fi

[[ -x "${SOURCE}" ]] || {
  echo "do-agent artifact is not executable: ${SOURCE}" >&2
  exit 1
}
file "${SOURCE}" | grep -Eq 'ELF 64-bit.*x86-64' || {
  echo "do-agent artifact must be linux/amd64 ELF: ${SOURCE}" >&2
  exit 1
}

ACTUAL="$(do_agent_sha256 "${SOURCE}")"
if [[ "${ACTUAL}" != "${EXPECTED}" ]]; then
  echo "do-agent artifact checksum mismatch: expected ${EXPECTED}, got ${ACTUAL}" >&2
  exit 1
fi

mkdir -p "${DESTINATION}"
install -m 0755 "${SOURCE}" "${DESTINATION}/do-agent-binary"
echo "staged do-agent artifact ${ACTUAL}"

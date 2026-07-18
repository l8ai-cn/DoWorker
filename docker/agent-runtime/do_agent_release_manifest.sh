#!/usr/bin/env bash

DO_AGENT_RELEASE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DO_AGENT_RELEASE_MANIFEST="${DO_AGENT_RELEASE_ROOT}/docker/agent-runtime/do-agent-release.json"

do_agent_release_value() {
  local path="${1:?release manifest path required}"
  python3 - "${DO_AGENT_RELEASE_MANIFEST}" "${path}" <<'PY'
import json
import pathlib
import sys

value = json.loads(pathlib.Path(sys.argv[1]).read_text())
for segment in sys.argv[2].split("."):
    value = value[segment]
if not isinstance(value, (str, int)):
    raise SystemExit(f"{sys.argv[2]} must be a scalar")
print(value)
PY
}

do_agent_sha256() {
  local artifact="${1:?artifact required}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${artifact}" | awk '{print "sha256:" $1}'
  else
    shasum -a 256 "${artifact}" | awk '{print "sha256:" $1}'
  fi
}

do_agent_require_digest() {
  local name="${1:?digest name required}" value="${2:-}"
  [[ "${value}" =~ ^sha256:[a-f0-9]{64}$ ]] || {
    echo "${name} must be an immutable sha256 digest" >&2
    return 1
  }
}

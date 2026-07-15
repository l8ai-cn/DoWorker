#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/worker-context.sh"
slug="$(worker_slug)"
definition="${LOOP_ROOT}/artifacts/definition.json"
image="${LOOP_ROOT}/evidence/images/runtime-image.json"
test_log="${LOOP_ROOT}/evidence/tests/runner-runtime.txt"
adapter_id="$(jq -r '.adapter_id' "$definition")"

[[ -s "$image" ]] || {
  echo "missing runtime image evidence: $image" >&2
  exit 1
}
[[ -s "$test_log" ]] || {
  echo "missing Runner runtime test log: $test_log" >&2
  exit 1
}
jq -e --arg slug "$slug" --arg adapter_id "$adapter_id" '
  .schema_version == 1 and
  .worker_slug == $slug and
  .adapter_id == $adapter_id and
  (.image_reference | type == "string" and length > 0) and
  (.image_digest | test("^sha256:[a-f0-9]{64}$")) and
  (.upstream_version | type == "string" and length > 0) and
  (.probe_command | type == "string" and length > 0) and
  (.probe_evidence | type == "string" and length > 0) and
  (.adapter_source_path | type == "string" and startswith("runner/"))
' "$image" >/dev/null

adapter_source="$(jq -r '.adapter_source_path' "$image")"
probe_evidence="$(jq -r '.probe_evidence' "$image")"
[[ "$adapter_source" != *".."* && "$probe_evidence" != /* && "$probe_evidence" != *".."* ]]
[[ -s "${REPO_ROOT}/${adapter_source}" ]] || {
  echo "missing adapter source: ${adapter_source}" >&2
  exit 1
}
[[ -s "${LOOP_ROOT}/${probe_evidence}" ]] || {
  echo "missing probe evidence: ${probe_evidence}" >&2
  exit 1
}

if rg -q 'return TransportTypeACP|falling back to ACP' \
  "${REPO_ROOT}/runner/internal/acp/transport.go"; then
  echo "Runner must reject unknown adapters instead of falling back to ACP" >&2
  exit 1
fi
if rg -q 'e2e-mock-agent-binary.*(loopal-binary|do-agent-binary)' \
  "${REPO_ROOT}/docker/agent-runtime/prepare_binaries.sh"; then
  echo "runtime build cannot substitute e2e-mock-agent for a product Worker" >&2
  exit 1
fi

(cd "$REPO_ROOT" && go test ./runner/internal/acp ./runner/internal/runner -count=1)

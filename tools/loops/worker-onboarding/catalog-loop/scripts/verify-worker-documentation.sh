#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"

cd "$REPO_ROOT"

! rg -q 'Any terminal-based agent works|bazel run //deploy/dev:up|agentsmesh-runner' \
  README.md runner/README.md
rg -q './deploy/dev/dev.sh' README.md
rg -q 'do-worker-runner register' README.md runner/README.md
rg -q 'Worker Runtime Status' README.md
rg -q 'No Worker type is formally deployable' README.md
! rg -q '3-step wizard' docs/api/workers-create.md
rg -q 'The current product form has four ordered steps' docs/api/workers-create.md
rg -q 'not a 1:1 API for the Worker wizard' docs/api/workers-create.md
rg -q 'CreatePodRequest.worker_spec' docs/api/workers-create.md
! rg -q 'Full example (matches UI wizard)' docs/api/workers-create.md
rg -q 'config/worker-types/catalog.json' docs/agent-runtime-build-audit.md
rg -q 'No Worker type is formally deployable' docs/agent-runtime-build-audit.md
rg -q 'not pullable release artifacts' docs/agent-runtime-build-audit.md
rg -q '`verified_local_dev` is local-development evidence' \
  docs/agent-runtime-build-audit.md
rg -q 'not be described as built-in runnable agents' \
  docs/integrations/openclaw-hermes.md
rg -q 'but it is not' docs/integrations/do-agent.md
rg -q 'currently selectable because no immutable runtime image digest has been' \
  docs/integrations/do-agent.md

expected="$(jq -r '.worker_types[].slug' config/worker-types/catalog.json | sort)"
actual="$(jq -r '.workers[].slug' clients/web/src/generated/worker-runtime-catalog.json | sort)"
[[ "$actual" == "$expected" ]] || {
  echo "generated Worker documentation catalog does not match formal Definitions" >&2
  exit 1
}

jq -e '
  .workers | length == 12 and
  all(.[]; .validationStatus != "verified_local_dev") and
  ([.[] | select(.validationStatus == "local_evidence_release_blocked")] | length) == 1 and
  ([.[] | select(.validationStatus == "invalid_published_runtime")] | length) == 2
' clients/web/src/generated/worker-runtime-catalog.json >/dev/null

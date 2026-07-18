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
rg -q 'Resource-native Worker' docs/api/workers-create.md
rg -q 'Direct WorkerSpec' docs/api/workers-create.md
rg -q 'External REST' docs/api/workers-create.md
rg -q 'CreateWorkerFromPlan' docs/api/workers-create.md
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
  .workers as $workers |
  all($workers[]; .validationStatus != "verified_local_dev") and
  ([$workers[] | select(.validationStatus == "local_evidence_release_blocked")] | length) == 0 and
  any($workers[]; .slug == "codex-cli" and .validationStatus == "runtime_ready_unverified") and
  any($workers[]; .slug == "video-studio" and .validationStatus == "runtime_ready_unverified")
' clients/web/src/generated/worker-runtime-catalog.json >/dev/null

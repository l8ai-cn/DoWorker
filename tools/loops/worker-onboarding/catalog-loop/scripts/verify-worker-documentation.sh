#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"

cd "$REPO_ROOT"

! rg -q 'Any terminal-based agent works|agentsmesh-runner' \
  README.md runner/README.md
rg -q './deploy/dev/dev.sh' README.md
rg -q 'do-worker-runner register' README.md runner/README.md
rg -q 'Worker Runtime Status' README.md
rg -q 'No Worker type is formally deployable' README.md
! rg -q '3-step wizard' docs/api/workers-create.md
rg -q 'Resource-native Worker' docs/api/workers-create.md
rg -q 'Resource Apply 失败时不会调用 Direct WorkerSpec 或 External REST' \
  docs/api/workers-create.md
rg -q 'Worker 是 create-only' docs/api/workers-create.md
rg -q 'PodService.ListWorkerCreateOptions' docs/api/workers-create.md
rg -q 'PreflightWorker.*resolved spec' docs/api/workers-create.md
rg -q '会重新解析同一 draft' docs/api/workers-create.md
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

expected_worker_slugs="$(jq -c '[.worker_types[].slug] | sort' config/worker-types/catalog.json)"
jq -e --argjson expected_worker_slugs "$expected_worker_slugs" '
  ([.workers[].slug] | sort) == $expected_worker_slugs and
  all(.workers[];
    .validationStatus | IN(
      "invalid_published_runtime",
      "local_evidence_release_blocked",
      "requires_model_resource",
      "runtime_image_unavailable",
      "runtime_ready_unverified"
    )
  )
' clients/web/src/generated/worker-runtime-catalog.json >/dev/null

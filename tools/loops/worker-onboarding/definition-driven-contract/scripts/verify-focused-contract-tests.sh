#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
cd "$repo_root"

go test ./backend/internal/service/workercreation -count=1
go test ./backend/internal/api/connect/pod -count=1
go test ./backend/internal/domain/workerspec -count=1
go test ./backend/internal/service/workerconfigmigration -count=1
go test ./runner/internal/runner -run \
  'TestPrepareAgentHome_(MergesConfigAndRemovesFromFilesToCreate|OpenClawBootstrapsProviderConfig)' \
  -count=1

(cd clients/core && cargo test -p pod_proto --test worker_creation_wire_contract -q)
(cd clients/core && cargo test -p agentcloud_api_client --test orchestration_resource_connect -q)
(cd clients/core && cargo test -p agentcloud_services --test orchestration_resource_service -q)
pnpm run build:wasm
bash deploy/dev/wasm_freshness_contract_test.sh

(cd clients/web && \
VITEST=true NODE_OPTIONS='--experimental-require-module' \
  ../../node_modules/.bin/vitest run --config vitest.config.ts --no-color \
  src/components/resource-editor/ResourceEditorSessionProvider.test.tsx \
  src/components/resource-editor/WorkerTemplateRuntimePanel.test.tsx \
  src/components/resource-editor/worker-template-runtime-options.test.ts \
  src/components/resource-editor/WorkerTemplateConfigDocumentBindingsField.test.tsx \
  src/components/resource-editor/use-resource-reference-options.test.tsx \
  src/lib/api/__tests__/podWorkerCreation.test.ts \
  src/lib/api/connect/orchestrationResourceConnect.test.ts)

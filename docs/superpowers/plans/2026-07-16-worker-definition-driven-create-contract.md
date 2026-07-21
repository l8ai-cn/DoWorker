# Definition-Driven Worker Create Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Worker creation derive compatible model resources, credential references, and named configuration-document bindings from the selected Worker Definition.

**Architecture:** `config/worker-types/<slug>/definition.json` remains the only Worker integration schema. The Backend projects a redacted, typed create contract and validates it again during preflight/create; Rust Core transports the Connect contract; Web renders the projection and submits references only. Existing anonymous `config_bundle_ids` are migrated before removal rather than treated as a fallback.

**Tech Stack:** Go, Connect-RPC/protobuf, PostgreSQL migrations, Rust WASM services, Next.js/TypeScript, Vitest, Go tests, browser E2E.

---

### Task 1: Name Configuration-Document Bindings

**Files:**
- Modify: `proto/pod/v1/worker_creation.proto`
- Modify: `backend/internal/domain/workerspec/workspace.go`
- Modify: `backend/internal/service/workercreation/workspace.go`
- Modify: `clients/core/crates/services/src/pod.rs`
- Modify: `clients/core/crates/api-client/src/modules/pod.rs`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationTypes.ts`
- Modify: `clients/web/src/lib/api/connect/podWorkerDraftProto.ts`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationConnect.ts`
- Test: `backend/internal/service/workercreation/workspace_config_documents_test.go`
- Test: `clients/web/src/lib/api/__tests__/podWorkerCreation.test.ts`

- [ ] Write a Backend test with a Definition containing `settings` and `openclaw-json`, and assert that `{document_id, config_bundle_id}` resolves only when the document ID exists, the bundle is `config`, and the declared document format parses.
- [ ] Run `go test ./backend/internal/service/workercreation -run ConfigDocument`; confirm it fails because the draft carries only raw IDs.
- [ ] Add `WorkerConfigDocumentBinding { document_id, config_bundle_id }` to `WorkerSpecDraft`; replace `config_bundle_ids` in the canonical WorkerSpec workspace with `ConfigDocumentBindings`.
- [ ] Resolve each binding against `Definition.ConfigDocuments`; preserve the Definition-owned ID, format, target path, bundle revision, and content hash for the compiled execution manifest.
- [ ] Regenerate protobuf code with `pnpm proto:gen-go-all`, then update Rust/TypeScript wire adapters without introducing an old-field read path.
- [ ] Run the focused Go and TypeScript tests; assert an unknown ID, duplicate ID, wrong bundle kind, and malformed JSON each fail with a field-specific error.

### Task 2: Project Definition-Specific Create Fields

**Files:**
- Modify: `proto/pod/v1/worker_creation.proto`
- Modify: `backend/internal/service/workercreation/options.go`
- Modify: `backend/internal/api/connect/pod/worker_creation_options.go`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationTypes.ts`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationConnect.ts`
- Modify: `clients/web/src/lib/api/facade/podConnect.ts`
- Test: `backend/internal/api/connect/pod/worker_creation_test.go`
- Test: `clients/web/src/lib/api/__tests__/podWorkerCreation.test.ts`

- [ ] Write a Connect handler test for Do Agent and OpenClaw that expects named config documents and credential-reference fields from their Definitions, and expects OpenCode to return neither.
- [ ] Run `go test ./backend/internal/api/connect/pod -run WorkerCreateOptions`; confirm the expected projection fields do not exist.
- [ ] Extend `WorkerTypeOption` with redacted credential-reference requirements and configuration-document declarations. Return target environment field, source kind, required bundle reference, document ID, format, and target path; never return secret values.
- [ ] Regenerate protobufs and map the projection through Rust Core and the Web wire adapter.
- [ ] Run the focused handler and Web adapter tests; assert the projection equals Definition data and contains no credential material.

### Task 3: Return Backend-Filtered Model Choices

**Files:**
- Modify: `backend/internal/service/workercreation/options.go`
- Modify: `backend/internal/service/workercreation/model.go`
- Modify: `backend/internal/api/connect/pod/worker_creation_options.go`
- Modify: `proto/pod/v1/worker_creation.proto`
- Modify: `clients/core/crates/services/src/pod.rs`
- Modify: `clients/core/crates/api-client/src/modules/pod.rs`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationTypes.ts`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationConnect.ts`
- Test: `backend/internal/service/workercreation/options_model_candidates_test.go`
- Test: `backend/internal/api/connect/pod/worker_creation_test.go`

- [ ] Write a service test with OpenAI-compatible, Anthropic, Gemini, and disabled resources. Assert Claude receives only the compatible selectable Anthropic reference, Gemini only Gemini, and OpenCode no primary-model field.
- [ ] Run `go test ./backend/internal/service/workercreation -run ModelCandidates`; confirm it fails because options expose only protocol strings.
- [ ] Add a metadata-only compatible-resource query to the existing AI-resource resolver. It must apply organization visibility, connection/resource enabled state, modality, capability, and Definition protocol adapters.
- [ ] Include only resource ID, revision, display label, selectable state, and blocking reason in `WorkerTypeOption`; the Backend remains authoritative and preflight/create repeat exact resolution.
- [ ] Regenerate protobufs and update Rust/Web mapping. Do not move provider protocol filtering into the Web.
- [ ] Run focused Go tests and `cargo test -p agentcloud-services orchestration_resource_service`; confirm incompatible IDs cannot appear as selectable choices.

### Task 4: Replace Static Web Form Logic

**Files:**
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerPrimaryModelField.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerConfigFileSelect.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerWorkspaceCapabilities.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/workerModelResources.ts`
- Modify: `clients/web/src/components/pod/hooks/useWorkerCreateDependencies.ts`
- Modify: `clients/web/src/components/pod/hooks/workerCreateDraft.ts`
- Modify: `clients/web/src/messages/en/*.json`
- Modify: `clients/web/src/messages/zh/*.json`
- Test: `clients/web/src/components/pod/CreatePodForm/__tests__/WorkerCreateFlow.test.tsx`
- Test: `clients/web/src/components/pod/CreatePodForm/__tests__/workerModelResources.test.ts`

- [ ] Write a UI test that selects OpenClaw and expects one labelled `openclaw-json` JSON document field, then selects OpenCode and expects no configuration selector.
- [ ] Write a UI test that selects Claude and sees only Backend-returned candidates, without importing or consulting a static Worker-to-protocol map.
- [ ] Run the focused Vitest files; confirm both fail against the current single-JSON selector and `AGENT_PROTOCOLS` map.
- [ ] Render credential bundle references only for Definition `credential_bundle` bindings. Keep secret values in the existing credential-bundle settings flow; Worker creation submits bundle IDs and target fields only.
- [ ] Render one config selector/upload control per Definition document; enforce declared format before creating a config bundle and persist `{document_id, config_bundle_id}` in draft state.
- [ ] Remove the static primary-model Worker map after all callers use Backend-projected candidates.
- [ ] Run the focused Vitest files and `pnpm run web:typecheck`; assert type switches reset stale bindings and no raw key field exists.

### Task 4.1: Remove The Legacy Direct Pod Create Path

**Files:**
- Modify: `clients/web/src/components/ide/CreatePodModal.tsx`
- Modify: `clients/web/src/app/(dashboard)/[org]/workspace/page.tsx`
- Modify: `clients/web/src/app/(dashboard)/[org]/tickets/page.tsx`
- Modify: `clients/web/src/components/ide/IDEShell.tsx`
- Modify: `clients/web/src/components/tickets/SpawnPodButton.tsx`
- Modify: `clients/web/src/components/tickets/SidebarPodSection.tsx`
- Modify: `clients/web/src/components/tickets/TicketPodPanel.tsx`
- Test: `clients/web/src/app/(dashboard)/[org]/workspace/__tests__/page.test.tsx`
- Test: `clients/web/src/components/tickets/__tests__/TicketDetail-actions.test.tsx`

- [ ] Write a failing browser or component test that opens each current
  `CreatePodModal` entry point and asserts it routes to the versioned
  `/{org}/workers/new` flow with its context encoded as a Worker draft, rather
  than rendering a direct Pod form.
- [ ] Remove the user-reachable direct `createPod` creation flow. Do not leave
  it as a hidden compatibility path, because it has no WorkerTemplate snapshot
  and cannot carry named configuration-document bindings.
- [ ] Preserve ticket, repository, prompt, and initial Worker type context by
  passing it to the resource-based Worker creation flow; do not reintroduce
  raw AgentFile or model protocol decisions in the UI.
- [ ] Run Workspace, Ticket, IDE, and mobile navigation tests plus a browser
  smoke test. Assert the legacy dialog no longer renders and no existing action
  can create a direct Pod.

### Task 5: Materialize Explicit Documents and Migrate Snapshots

**Files:**
- Create: `backend/migrations/000222_name_worker_config_document_bindings.up.sql`
- Create: `backend/migrations/000222_name_worker_config_document_bindings.down.sql`
- Modify: `backend/internal/service/workercreation/compiler.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_worker_spec_resources.go`
- Modify: `backend/internal/service/agent/config_builder_configbundle.go`
- Modify: `runner/internal/runner/pod_builder.go`
- Test: `backend/internal/service/workercreation/compiler_test.go`
- Test: `backend/internal/service/agentpod/pod_orchestrator_worker_spec_resources_test.go`
- Test: `runner/internal/runner/pod_builder_config_documents_test.go`

- [ ] Write a compiler test that binds Do Agent `settings` and OpenClaw `openclaw-json` to their Definition paths, and asserts a bundle cannot be rebound to another Worker document.
- [ ] Run the focused compiler test; confirm the current compiler emits anonymous `USE_CONFIG_BUNDLE` declarations only.
- [ ] Migrate all persisted WorkerSpec JSON snapshots atomically from positional IDs to Definition-backed named bindings. Abort migration if a historical snapshot cannot be mapped unambiguously; do not preserve dual read/write behavior.
- [ ] Include document ID, format, target path, content hash, and bundle revision in the redacted execution manifest. Resolve bundle contents only at Runner materialization.
- [ ] Ensure required resource inventory includes configuration bundle IDs and that Runner writes only within Definition-owned resolved targets.
- [ ] Run Go and Runner focused tests plus migration tests against a populated development database snapshot.

### Task 6: Verify Real Create Paths and Release Gates

**Files:**
- Modify: `tools/loops/worker-onboarding/catalog-loop/catalog/worker-evidence-matrix.json`
- Create: `tools/loops/worker-onboarding/catalog-loop/evidence/definition-driven-create-contract-2026-07-16.json`
- Modify: `docs/superpowers/specs/2026-07-16-worker-integration-runtime-architecture.md`

- [ ] Start the development stack and run `pnpm run build:wasm`, `pnpm run web:typecheck`, the focused Go/Runner tests, `go test ./backend/... ./runner/...`, and the runtime catalog contract checks.
- [ ] In a real browser session, create template drafts for Aider, Claude, Cursor, Do Agent, Gemini, Grok, Hermes, MiniMax, OpenClaw, OpenCode, and Seedance. Capture success/field-level failure evidence without entering provider secrets.
- [ ] With explicit non-production credential authorization, run one Pod lifecycle per eligible type: preflight, create, connect, harmless prompt, expected PTY/ACP event, terminate, and cleanup. Mark all unavailable providers as blocked, not supported.
- [ ] Update the evidence matrix only from command, API, Runner, and browser artifacts. A Worker becomes supported only when all its release gates pass.

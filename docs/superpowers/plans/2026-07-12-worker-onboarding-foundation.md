# Worker Onboarding Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make all 12 formal Worker types use one versioned Definition for configuration, credentials, frontend forms, backend dispatch, Runner adapters, and runtime images.

**Architecture:** `config/worker-types/` becomes the Git source of truth and is embedded in Backend. Backend returns immutable Definition version/hash snapshots and sends an explicit `adapter_id` to Runner. Web renders Definition metadata; Runner rejects an unknown adapter; no legacy AgentFile, frontend list, command-name inference, or mock-binary fallback remains on the product path.

**Tech Stack:** Go, Go embed, JSON Schema, Connect/Proto, Rust/WASM, Next.js, Vitest, Playwright, Docker.

---

## Scope And File Map

Formal slugs: `claude-code`, `codex-cli`, `gemini-cli`, `aider`, `opencode`, `loopal`, `cursor-cli`, `do-agent`, `minimax-cli`, `grok-build`, `openclaw`, `hermes`. `e2e-echo` remains test-only.

- Create: `config/worker-types/catalog.json`, `config/worker-types/schema/definition.schema.json`
- Create: one `definition.json` and `AgentFile` under each formal slug directory
- Create: `backend/internal/service/workerdefinition/{catalog_embed,definition,catalog_loader}_test.go`
- Modify: `backend/internal/service/workercreation/worker_type.go`, `backend/internal/service/workerspec/contracts.go`
- Modify: `proto/pod/v1/worker_creation.proto`, `proto/runner_api/v1/runner.proto`
- Modify: `clients/core/crates/api-client/src/modules/pod.rs`, `clients/core/crates/services/src/pod_worker_creation.rs`, `clients/core/crates/wasm/src/service_pod_worker_creation.rs`
- Modify: `clients/web/src/components/pod/CreatePodForm/*`, `clients/web/src/lib/api/connect/podWorkerCreationConnect.ts`
- Modify: `runner/internal/acp/transport.go`, `runner/internal/runner/{pod_builder.go,pod_builder_build.go,message_handler_acp.go}`, `docker/agent-runtime/{Dockerfile,prepare_binaries.sh,build.sh}`
- Delete after Web migration: `clients/web/src/components/settings/envBundleCredentialForms/credentialBuiltinFallbacks.ts`, `credentialUxOverrides.ts`

### Task 1: Define The Canonical Catalog

- [ ] Add a failing table-driven test in `backend/internal/service/workerdefinition/catalog_loader_test.go` that loads the 12 expected slugs, rejects an extra slug, recomputes each catalog file hash, and asserts each Definition has `adapter_id`, modes, credential bindings, configuration documents, and image probe.
- [ ] Create `definition.schema.json` and one `definition.json` per formal slug. Every entry follows this shape:

```json
{"schema_version":1,"slug":"codex-cli","definition_version":"1","adapter_id":"codex-app-server","interaction_modes":["pty","acp"],"credential_bindings":[],"config_documents":[],"image":{"runtime":"codex-cli","version_probe":["codex","--version"]}}
```

- [ ] Copy each current AgentFile into its Definition directory, record its SHA-256 in `definition.json`, and make `catalog.json` list only `{slug,definition_path,definition_hash}`.
- [ ] Run `go test ./backend/internal/service/workerdefinition -count=1` and `bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-catalog-contract.sh`; expected result is pass only after all 12 Definition files exist.

### Task 2: Load Definitions And Replace Legacy Resolution

- [ ] Add a failing test proving `workercreation.ResolveWorkerType` gets its schema, version, hash, adapter ID, modes, credentials, and config documents from the embedded catalog rather than `agents.AgentfileSource`.
- [ ] Implement an embedded, immutable `workerdefinition.Catalog` with `Get(slug)` and `List()`; validate schema, slug, file hashes, duplicate entries, and catalog completeness during service startup.

```go
type Definition struct {
    Slug, Version, Hash, AdapterID string
    InteractionModes []workerspec.InteractionMode
    CredentialBindings []CredentialBinding
    ConfigDocuments []ConfigDocument
}
```

- [ ] Extend `workerspec.WorkerType` and `WorkerTypeResolution` with Definition version, hash, and adapter ID. Remove the Agent service and `.json` path heuristic from the Worker creation resolution path.
- [ ] Run `go test ./backend/internal/service/workerdefinition ./backend/internal/service/workercreation ./backend/internal/service/workerspec -count=1`.

### Task 3: Keep Credentials As References

- [ ] Add failing tests in `backend/internal/service/workercreation` that reject a binding with a direct value, reject an unbound secret reference, and assert `Snapshot.SpecJSON()` and preflight responses never contain a resolved credential.
- [ ] Define each binding as `{id, source:{kind,ref}, target:{kind,name}, required_when, display}`. `source.kind` is exactly `model_resource` or `credential_bundle`; `target.kind` is exactly `env` or `config_document`.
- [ ] Map the existing model-resource and environment-bundle resolvers by binding ID, materialize values only immediately before the authenticated Runner command, and keep the persisted WorkerSpec as IDs/revisions/refs.
- [ ] Run `go test ./backend/internal/service/workercreation ./backend/internal/service/workerspec -count=1`.

### Task 4: Carry Definition Snapshots Through APIs

- [ ] Add failing Connect tests for `GetWorkerTypeDefinition`, stale Definition hash rejection in `PreflightWorker`, and exact snapshot propagation in create-pod dispatch.
- [ ] Extend `proto/pod/v1/worker_creation.proto` with Definition version/hash and a `GetWorkerTypeDefinition` operation. Extend `proto/runner_api/v1/runner.proto` create-pod payload with `worker_type_slug`, `definition_hash`, and `adapter_id`.
- [ ] Regenerate with `pnpm proto:gen-go-all`; update Backend converters, Rust API client/service/WASM bindings, and Web facade types in the same change.
- [ ] Run `go test ./backend/internal/api/connect/... ./backend/internal/service/workercreation/... -count=1`, `cd clients/core && cargo test --workspace`, and `pnpm run build:wasm`.

### Task 5: Render Forms From Definitions

- [ ] Add Vitest coverage for Definition-driven required, conditional, loading, disabled, incompatible, and validation-error form states in `clients/web/src/components/pod/CreatePodForm/__tests__/`.
- [ ] Replace `workerTypeConfigSchema.ts` input with the typed Definition response; render credential binding display metadata and configuration document fields without per-slug switches.
- [ ] Delete `credentialBuiltinFallbacks.ts` and `credentialUxOverrides.ts`; a missing Definition is an API error state, not an empty or fallback form.
- [ ] Run `pnpm run web:test`, `pnpm run web:typecheck`, and `bash clients/web/scripts/check-no-wasm-in-marketing.sh`.

### Task 6: Enforce Exact Runner Adapters And Real Images

- [ ] Add failing tests in `runner/internal/acp` that unknown adapter IDs return an error and command names cannot select a transport. Add `pod_builder` tests that verify the Definition snapshot adapter ID reaches Runner unchanged.
- [ ] Change `NewTransport` to return `(Transport, error)` and delete `TransportTypeForCommand` fallback behavior. Register standard ACP explicitly as `acp-jsonrpc` and assign it in each Definition where applicable.
- [ ] Remove `e2e-mock-agent` substitutions from `prepare_binaries.sh`; Loopal and Do Agent builds must fail when their real artifact is absent. Add `cursor-cli` to the image build matrix or block it in Definition with an explicit unavailable reason.
- [ ] Run `go test ./runner/internal/acp ./runner/internal/runner -count=1`, `bash docker/agent-runtime/contract_test.sh`, and `bash docker/agent-runtime/build_extension_contract_test.sh`.

### Task 7: Run Two Pilots And Start The Remaining Queue

- [ ] Instantiate `runs/codex-cli` and `runs/gemini-cli` with `tools/loops/worker-onboarding/catalog-loop/scripts/instantiate-worker-loop.sh`.
- [ ] Complete their seven Worker Loop tasks, including upstream evidence, Definition hashes, real image digest/probe logs, Runner smoke results, browser screenshots, and independent review JSON.
- [ ] Run browser E2E against `http://localhost:10007`: create a Codex Worker successfully; verify Gemini loading, invalid credential reference, incompatible model, disabled image, and backend validation failures. Save evidence under each run’s `evidence/browser/`.
- [ ] Run each Worker terminal verifier, mark only verifier-backed checklist entries, then run the Catalog verifier. Do not publish images, push, merge, deploy, or use real credentials without the recorded human gate.

## Completion Checks

- [ ] `bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-inventory.sh` exits 0.
- [ ] `bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-catalog-contract.sh` exits 0.
- [ ] Both pilot Worker terminal verifiers exit 0 with non-mock image evidence.
- [ ] All Go, Rust, Web, Docker contract, and browser checks above pass.
- [ ] The catalog loop remains `running` until every formal Worker run is accepted and the independent human review gate is recorded.

# Unified AI Resource Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the approved Unified Resource Center, make provider connections and model resources the only model-credential path, and require Workers to submit one exact compatible resource.

**Architecture:** `provider_connections` owns encrypted account credentials; `model_resources` owns selectable chat/image/audio/video/embedding capabilities. Connect-RPC exposes catalog, scoped CRUD, effective reads, validation, and safe usage metadata through Rust/WASM. Worker orchestration resolves only `model_resource_id`; an explicit migration moves `ai_models` and credential EnvBundles before legacy paths are removed.

**Tech Stack:** Go/GORM/PostgreSQL, Connect-RPC/Protobuf, Rust/Cargo/WASM core, Next.js/TypeScript/Tailwind/shadcn, pnpm/Vitest/Playwright.

---

### Task 1: Provider Catalog and Resource Domain

**Files:** Create `backend/internal/domain/airesource/{catalog.go,connection.go,resource.go,repository.go,catalog_test.go}`.

- [x] Write `catalog_test.go` asserting `openai`, `anthropic`, `gemini`, `minimax`, `dashscope`, `doubao`, `deepseek`, `elevenlabs`, `runway`, `kling`, `hailuo`, `luma`, `replicate`, and `fal` exist and expose non-empty modalities and credential fields.
- [x] Run `go test ./backend/internal/domain/airesource`; expect missing package failure.
- [x] Implement `ProviderDefinition`, `CredentialField`, `Modality`, `Connection`, `ModelResource`, `UsageSummary`, owner-scope constants, and `Provider(key)`/`Providers()` over an immutable registry.
- [x] Add `ValidateIdentifier` checks for provider/resource identifiers and keep every source file below 200 lines.
- [x] Re-run the target; expect PASS. Commit `feat: add AI resource provider catalog`.

### Task 2: Canonical Schema and Repository

**Files:** Create `backend/migrations/000190_ai_resources.{up,down}.sql`, `backend/internal/infra/ai_resource_repo.go`, `backend/internal/infra/ai_resource_repo_test.go`, `backend/internal/testkit/schema_ai_resource.go`; modify `backend/internal/testkit/db.go`.

- [x] Write repository tests for personal/org effective visibility, disabled filtering, scoped uniqueness, one default per `(owner_scope, owner_id, modality)`, and personal-over-organization effective default precedence.
- [x] Run `go test ./backend/internal/infra -run AIResource`; expect missing tables/repository failure.
- [x] Create `provider_connections`, `model_resources`, `model_resource_defaults`, and `ai_resource_migration_map` with owner checks, identifier checks, foreign keys, indexes, and scoped uniqueness. Do not drop legacy tables.
- [x] Implement transactional repository CRUD, `ListEffective(userID, orgID, modalities)`, per-modality default promotion, personal-over-organization default projection, and ownership predicates.
- [x] Re-run focused tests; expect PASS. Commit `feat: persist provider connections and model resources`.

### Task 3: Encrypted Service, Permissions, and Validation

**Files:** Create `backend/internal/service/airesource/{service.go,connections.go,resources.go,effective.go,validation.go,service_test.go}`; modify `backend/pkg/audit/audit.go`.

- [x] Write service tests proving write-only credentials, owner/admin mutation rules, member safe reads, exact resource resolution, incompatible modality rejection, disabled/invalid rejection, and audit event envelopes.
- [x] Run `go test ./backend/internal/service/airesource`; expect missing implementation failure.
- [x] Implement encryption, configured-field extraction, typed errors, scope policy, effective views, `ResolveExact`, and protocol validators for OpenAI-compatible, Anthropic-compatible, Gemini, and registry-declared media endpoints.
- [x] Enforce outbound validation policy that rejects loopback, private, link-local, and metadata-service targets; connections are selectable only after successful validation.
- [x] Re-run tests; expect PASS. Commit `feat: add scoped AI resource service`.

### Task 4: Connect API and Server Wiring

**Files:** Create `proto/ai_resource/v1/ai_resource.proto`, `backend/internal/api/connect/ai_resource/{server.go,queries.go,mutations.go,wire.go,server_test.go}`; modify `backend/cmd/server/{services_init.go,connect_init.go}`.

- [x] Write handler tests for catalog, effective list, personal/org CRUD, validation, permissions, and typed Connect codes.
- [x] Run `go test ./backend/internal/api/connect/ai_resource`; expect missing package failure.
- [x] Define `AIResourceService` messages with safe metadata only; implement handlers and mount them behind existing auth/org interceptors.
- [x] Run `pnpm proto:gen-ts` and `pnpm proto:gen-go-all`; review generated/staged output so ignored local Go mirrors are not committed.
- [x] Re-run handler tests and `go test ./backend/cmd/server`; expect PASS. Direct-commit the resolved foundation on `main`.

### Task 5: Rust/WASM and TypeScript Client Boundary

**Files:** Create `clients/core/crates/api-client/src/modules/ai_resource.rs`, `clients/core/crates/services/src/ai_resource.rs`, `clients/core/crates/wasm/src/service_ai_resource.rs`, `clients/web/src/lib/api/connect/aiResourceConnect.ts`, `clients/web/src/lib/api/facade/aiResource.ts`; modify corresponding `lib.rs`, BUILD files, WASM API/getters, and `packages/service-runtime/src/{service-getters.ts,index.ts}`.

- [x] Write Rust wire tests and TS adapter tests proving protobuf request fields, response decoding, and no secret-valued response field.
- [x] Run the focused client-boundary tests; expected missing service/getter failures before implementation.
- [x] Implement Connect calls through `ApiClient`, service wrapper, WASM exports, runtime getter, and typed TS facade. Do not add direct `fetch` business calls.
- [x] Re-run focused adapter tests and offline Cargo WASM checks; commit `feat: add AI resource client boundary`.

### Task 6: Unified Resource Center

**Files:** Create `clients/web/src/components/settings/AIResourcesSettings/` with `AIResourcesSettings.tsx`, `ProviderConnectionCard.tsx`, `ModelResourceRow.tsx`, `ProviderConnectionDialog.tsx`, `ResourceSummary.tsx`, `useAIResources.ts`, `types.ts`, and tests; modify settings navigation, settings route switch, exports, and all locale message files.

- [x] Write UI tests for personal/org scopes, owner/admin/member actions, provider onboarding, validation, capability filters, loading/empty/error/invalid/disabled states, and absent usage displaying `未接入`.
- [x] Run targeted Vitest and confirm the missing component boundary before implementation.
- [x] Implement approved layout A using existing tokens/components, named state components, capability filters, safe metadata, manage actions, and truthful usage placeholders.
- [x] Keep production files below 200 lines and ensure keyboard/focus labels on dialogs, filters, and actions.
- [x] Re-run focused tests, scoped lint, and browser acceptance; record unrelated full-typecheck baseline failures separately. Commit `feat: add unified AI resource center`.


#### Task 5-6 execution evidence — 2026-07-10

- Task 5 compiles with offline Cargo WASM checks and its adapter test passes.
- Task 6 focused Vitest passed 34/34; scoped lint has no errors (one pre-existing unused-import warning in the existing settings sidebar).
- Browser acceptance covered personal create/edit/rotate/default/enable/delete, validation failure, loading/error, member read-only, desktop/mobile, console/network, and cleanup.
- Full web typecheck has 285 unrelated baseline errors; none reference AI Resource settings or dialog code.

### Task 7: Exact Worker Resource Selection

**Files:** Modify `proto/pod/v1/pod.proto`, pod Connect adapters, `clients/web/src/components/pod/CreatePodForm/WorkerCredentialModelSection.tsx`, hooks/types/tests, `backend/internal/service/agentpod/{pod_orchestrator.go,pod_orchestrator_create.go,pod_orchestrator_worker_model.go}`; create `WorkerModelResourceSelect.tsx`; delete `CredentialBundleSelect.tsx` after callers move.

- [x] Write failing Web tests: no default-auth option, only compatible chat resources, no-resource blocking state, exact `model_resource_id` submission, and load error visibility.
- [x] Write failing Go tests: explicit resource B resolves B, default A is never appended, missing/inaccessible/incompatible/invalid resource rejects creation, and no runtime bundle carries model credentials.
- [x] Run focused Web and agentpod tests; verify expected failures.
- [x] Implement explicit resource selection and `ResolveExact`; remove legacy automatic model-credential mounting and model auto-resolution. Backend generates ephemeral harness config from the selected connection/resource.
- [x] Regenerate proto mirrors, run focused tests/builds, expect PASS. Commit `fix: make Worker model resource selection explicit`.

### Task 8: Fail-Closed Legacy Migration and Virtual-Key Remap

**Files:** Create `backend/internal/service/airesource/{migration.go,migration_test.go}`, `backend/internal/service/envbundle/migration_export.go`, `backend/cmd/migrate-ai-resources/{main.go,BUILD.bazel}`, `backend/migrations/000191_ai_resource_cutover.{up,down}.sql`; modify virtual-key domain/service/repository and deployment migration docs.

- [x] Write migration tests for `ai_models`, mapped credential EnvBundles, idempotency, exact owner/field parity, unknown agent/provider failure, corrupt ciphertext failure, and unchanged sources after failure.
- [x] Run migration tests; expect missing migrator failure.
- [x] Implement the explicit application migrator using the production encryptor and mapping table. Preserve `ai_models.id` as `model_resources.id` so virtual keys can remap deterministically.
- [x] Add a verification command that exits non-zero on count, field, scope, decrypt, or mapping mismatch; make cutover migration require a clean report.
- [x] Re-run tests and a local seeded migration dry run; expect PASS. Commit `feat: migrate legacy model credentials fail closed`.

### Task 9: Remove Legacy Credential Paths and Update Documentation

**Files:** Delete old agent-credential proto/client/WASM facades, credential-profile fields, legacy automatic mount helpers, obsolete tests, and credential EnvBundle settings UI; modify EnvBundle docs/tests, Worker docs, AgentFile docs, and API docs.

- [x] Add negative source-contract tests asserting no legacy service, credential-profile field, default-auth flag, or credential auto-mount remains in active/generated contracts.
- [x] Run the contract tests; expect failures listing legacy symbols.
- [x] Remove legacy code only after Tasks 7-8 are green; keep runtime/config EnvBundles and their UI.
- [x] Run `rg` contract probes, affected Go/Rust/Web tests, and docs link checks; expect no forbidden active symbols outside historical migrations.
- [x] Commit `refactor: remove legacy model credential paths`.


#### Task 7-9 execution evidence — 2026-07-10

- Worker and Loop creation now submit exact `model_resource_id`; Go tests cover explicit selection, inaccessible/incompatible/missing rejection, and absence of model credentials in runtime EnvBundles.
- Legacy model credential migration is fail-closed: the application migrator and SQL cutover verify counts, owner/scope parity, decryptability, mappings, and virtual-key remap before old columns are removed.
- Old agent-credential protocol/client/UI/WASM paths were deleted. The remaining credential-form helpers are runtime EnvBundle maintenance utilities under `envBundleCredentialForms`, not a model credential selection UI.
- Focused verification passed: backend AI-resource/agentpod/loop/session/virtual-key/migration packages, Web AI-resource and Worker selection Vitest suites, `web-user` model-pool picker Vitest, `web-user` typecheck, proto Go/TS generation, and legacy source-contract tests.
- Current repository baseline still has unrelated full web typecheck failures and an unavailable Bazel workspace root; neither failure references the AI Resource or Worker model-resource changes.

### Task 10: Full Verification and Browser Acceptance

**Files:** Update E2E helpers/specs under `clients/web/e2e-playwright/tests/{settings,pod}` and retain screenshots under the project test-output convention.

- [x] Add E2E scenarios for personal connection creation, org permission denial, validation failure, chat/image/video filtering, exact Worker selection, no-resource blocking, and `未接入` quota/usage.
- [x] Run available backend/domain/service/API tests, generated-code checks, Web unit/typecheck probes, migration tests, and `git diff --check`; record existing full-typecheck/Bazel baseline blockers separately.
- [x] Start the app and execute browser QA for desktop/mobile, light/dark, loading/empty/error/invalid/permission/success states; inspect console and failed network requests.
- [x] Compare every acceptance scenario in the design document with test/browser evidence and record any residual baseline failures.
- [x] Commit `test: verify unified AI resource management` and only then mark the active goal complete.


#### Task 10 execution evidence — 2026-07-10

- Services: backend `http://127.0.0.1:10015/health` returned 200 `ok`; relay `10017` listened; frontend `17557` returned 200.
- Browser QA used an isolated Chrome context. Login succeeded as `dev@agentsmesh.local`; the stale `admin-workspace` redirect was corrected by opening the real `dev-org` route.
- AI Resource Center desktop and mobile/dark routes rendered personal scope, empty state, capability tabs, provider catalog, add-connection dialog, write-only API key field, and `未接入` usage placeholder. AI Resource Connect RPCs returned 200.
- Validation failure path: creating an OpenAI connection with a loopback base URL was rejected by `CreatePersonalConnection` with HTTP 400, left the connection count at 0, and did not echo the API key in the DOM.
- Worker creation route `/dev-org/workers/new` rendered exact-resource copy, `暂无兼容的模型资源` blocking state, disabled create buttons, and no old default-auth or API credential Bundle copy. `GetCatalog` and `ListOrganizationEffectiveResources` returned 200.
- Browser console was clean of AI Resource/Worker model-resource errors after fixing the duplicate `workerSlash` translation key usage; remaining DevTools issue advisories are generic form-label accessibility findings on the Worker page.
- Verification passed: focused Go packages, legacy source-contract test, proto Go/TS generation, focused Web Vitest 129/129, `web-user` model-pool Vitest, `web-user` typecheck, filtered Web typecheck for AI Resource/Worker symbols, and `git diff --check`.
- Residual baseline: full `clients/web` typecheck still reports unrelated existing errors, currently including `clients/web/src/lib/expert-form-prefill.ts` importing a moved hook type; current checkout has no Bazel `MODULE.bazel`/`WORKSPACE`, so Bazel Rust/WASM build is unavailable in this local baseline.

## Mainline follow-up verification — 2026-07-11

**Purpose:** Verify the merged Task 7–10 chain on current `main` without
absorbing unrelated in-progress Worker, Pod lifecycle, or Rust Core changes.

- [x] Clear only verified stale local resources: no live `.git/index.lock`;
  terminate the orphaned Playwright daemon and its headless Chrome children;
  remove their unused browser profiles.
- [x] Run the fail-closed source-contract and migration suites. Any legacy
  credential-path reference in an active contract, migration parity mismatch,
  or virtual-key remap error is a blocking defect.
- [x] Run exact-resource regression suites for AgentPod, Loop, REST session,
  and Connect Pod creation. Missing, invisible, disabled, unhealthy, or
  incompatible `model_resource_id` must reject creation.
- [x] Run focused Web unit/type probes and current-browser acceptance for the
  Resource Center and Worker creation success, empty, error, and permission
  states. Do not mask a test failure with a fallback or stale fixture.
- [x] Review migration and public documentation against the executed behavior,
  check the AI-resource commit range for whitespace/errors, and report only
  reproducible residual blockers.

**Execution limits:** Stop and report after three no-progress attempts with the
same environmental root cause. Do not stage, discard, or reformat unrelated
dirty files. Completion requires all focused checks to pass or each unrelated
baseline failure to be reproduced and evidenced separately.

#### Mainline follow-up execution evidence — 2026-07-11

- There was no Git index lock. The only lock risk was a short-lived AutoSync
  Git process; it exited normally. An orphaned Playwright daemon, six headless
  Chrome children, and two unused profiles were removed after process checks.
- Browser QA initially found `ListWorkerCreateOptions` returning 404. Source
  registration was correct; the host backend binary had been running since
  July 10, before the route existed. Restarting it through
  `deploy/dev/lib/host_services.sh:start_backend_host` loaded current `main`;
  the procedure now returns 401 without auth and 200 in the logged-in browser.
- Resource Center personal-empty and Worker no-resource states render with
  successful AI Resource/Pod Connect RPCs and no browser console errors.
  The Worker form now shows the existing localized blocking explanation instead
  of an empty model-resource listbox. Desktop and 390×844 mobile screenshots
  are under `output/playwright/`.
- Passed: AI resource migration/legacy contract suite, cutover migration test,
  AgentPod/Loop/REST/Connect exact-resource suites, virtual-key suite, focused
  Web tests (11/11), scoped ESLint, and whitespace/source-contract checks.
- `pnpm run web:typecheck` still fails on unrelated repository-wide WASM,
  i18n, E2E, and legacy component errors. Its output contains no error for the
  touched Worker resource files.
- The empty-state fix and request-id test correction are part of the current
  untracked Worker creation file group. They remain unstaged because that group
  also contains unrelated in-progress creation changes and must be committed
  cohesively by its owner.

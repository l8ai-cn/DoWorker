# Resource-Native Orchestration Verification

This file stores durable verification evidence for the
[resource-native orchestration goal](./2026-07-14-resource-native-orchestration-goal.md).

Audit date: 2026-07-18.

## Verification Snapshot

- Expert service/REST guard packages pass their full focused suites and two
  independent reviews approved the behavior and code quality.
- Workflow Connect, GoalLoop Connect, Workflow REST, Runner MCP, Workflow
  service, AgentPod inventory, and focused Infra tests pass after legacy
  definition/runtime cutover.
- Quick Task consumes only a Worker Plan, reuses the durable Worker launch, and
  reports the applied Pod's current status; it no longer selects agent, Runner,
  repository, prompt, alias, AgentFile, or queue TTL at runtime.
- Runner MCP `create_pod` accepts only a Worker `resource` manifest. The Backend
  executes Validate, Plan, caller authorization, and typed Worker Apply, then
  returns the applied resource revision and WorkerSpec snapshot. The Runner
  keeps automatic Pod read/write binding after creation and reports validation,
  apply, or binding failures as redacted tool errors.
- Runner MCP full package and build-tag integration suites pass. An independent
  two-pass review found and then verified closure of one P1 and three P2
  contract/documentation issues; no P0/P1/P2 remain in that cutover.
- Session create, fork, and import now materialize deferred `queued` Pods and
  atomically commit the Session, all initial/copied/imported items, and the
  serialized create command in `pending_runner_commands`. Local Session events
  publish before drain is triggered, so an asynchronous `initializing` event
  cannot be followed by a stale handler-emitted `queued` event.
- A failure before that transaction commits rolls back Session/items/command and
  terminates the materialized Pod with a detached, bounded compensation
  context. A crash or Runner send failure after commit preserves the owner and
  durable command; the existing backlog sweep and Runner reconnect drain retry
  it, with a CAS `queued -> initializing` claim before every send.
- Session, Prompt outbox, and ordinary pending-command admission now share the
  same Runner advisory lock and `MaxPerRunner` boundary. Offline Session
  admission is rejected when the queue feature is disabled; an already
  connected Runner may still use the row as a short-lived transactional outbox
  so a process crash cannot separate owner persistence from delivery.
- Switch persists the new binding before synchronous dispatch, keeps the old
  Pod alive until dispatch succeeds, and restores the old binding if dispatch
  fails. Crash-consistent old/new replacement ownership still requires the
  frozen migration work and is not claimed complete.
- Coordinator task ownership now wraps active-execution lookup, external
  link/Ticket materialization, and claimed-execution creation in one PostgreSQL
  transaction-level advisory lock. The first-discovery concurrency test proves
  one Ticket and one Pod across overlapping runs.
- Full Session, AgentPod, Coordinator, Infra, and server focused suites pass.
  Focused race suites pass for Session lifecycle, Coordinator concurrency, and
  deferred AgentPod dispatch.
- An independent follow-up review found one P1 queue-admission bypass in the
  first durable Session implementation. The unified Runner lock, enabled-state
  admission, atomic capacity check, 429/503 behavior, and cross-entry tests
  closed it; the reviewer found no remaining non-migration-blocking P0/P1/P2.
- The full Runner workspace suite reached one existing Autopilot logger timeout;
  the failing test passed three isolated reruns, while every Runner MCP package
  test passed in the full invocation.
- Apply authorization unit tests and orchestration-control service tests pass.
- Frontend resource-editor review found a P1 that left successful and failed
  Apply Plans replayable. Apply terminal states now disable the old Plan, and
  starting a new Plan clears the terminal Apply state. Typed WASM errors render
  only the service message, without their JSON envelope or internal request URL.
- Reference catalogs now isolate failures by Kind, so one forbidden or
  unavailable dependency does not clear successfully loaded choices. Mobile
  Validate/Plan actions use 44-pixel targets and Apply spans the action row.
- A final Definition-binding review found that the orchestration resource layer
  incorrectly applied slug rules to Worker config and credential target names.
  It now reuses the Workerspec field contract, so real names such as
  `CURSOR_API_KEY` remain intact through Plan and compilation. Worker option
  errors use the same URL redaction, and credential required state comes from
  the projected Worker schema.
- The resource-editor and Connect adapter suites pass 75 tests across 19 files.
  `web:typecheck` passes, and Playwright statically lists six resource-editor
  scenarios: desktop Apply, mobile invalid YAML, permission denial, blocking
  issue, Plan expiry/recovery, and Apply conflict/re-plan.
- EnvironmentBundle reference options now use a typed ListResources filter.
  The repository joins each resource's immutable active binding revision to
  current EnvBundle ownership, active state, purpose, and Worker compatibility.
  Worker Definition is the server-side field-policy SSOT: runtime catalogs
  exclude bundles containing model-managed fields, and each credential target
  receives a separate catalog that requires the exact declared key. Config
  remains a purpose-only catalog. Unsupported database dialects and corrupt
  active bindings fail explicitly. The response echoes purpose, Worker type,
  and target name, so a client cannot accept an old control plane's unfiltered
  response. Integrity validation, count, and page reads share one consistent
  transaction snapshot. YAML remains editable and Plan remains authoritative
  for changes after candidate loading.
- Go control-plane tests, 75 Web tests, and four filtered Rust protocol/service/
  WASM tests pass for the Definition-aware EnvironmentBundle contract.
- An independent read-only review found no P0/P1/P2 in the Definition policy,
  permission boundary, Secret handling, SQL predicates, stale-request guards,
  old-server acknowledgement rejection, or cross-language protocol wiring.
  The final delivery reran the migration suites against real PostgreSQL,
  started the control plane from an empty worktree database, and exercised the
  browser paths. Mixed-version production rollout remains release-owner work.
- Full Web lint passes with zero errors and 180 existing warnings.
  `ResourceEditorSessionProvider.test.tsx` and the files changed by the
  EnvironmentBundle slice pass their focused checks.
- A fresh read-only execution audit confirmed that none of the remaining six
  legacy constructors has a safely deployable no-migration cutover. Session
  create/import/switch/host-bind and Coordinator need persisted Plan/snapshot
  ownership; same-Worker fork is only mechanically possible after Sessions
  carry that immutable binding.
- Source-lineage resume still accepts a Pod without `WorkerSpecSnapshotID`
  because the six migration-gated constructors can currently produce those
  Pods and Session message recovery depends on them. This is an explicit
  unresolved cutover blocker, not an accepted compatibility contract. The
  final migration phase must convert those constructors first, then make
  missing-snapshot lineage fail closed and update the legacy resume tests.
- PodOrchestrator now resolves plan, snapshot, and source lineage through one
  `ExecutionSource` contract before any preparation or persistence and rejects
  all three pairwise source conflicts. Persistent Workflow runs use lineage
  only, while fresh runs use their pinned snapshot; the current run prompt is
  appended as invocation metadata without changing the inherited snapshot.
  AgentPod and Workflow full package suites pass with 16 inventory entries.
- Formal release mainline is confirmed through
  `000225_agent_workbench_stream`. This delivery uses the coordinated sequence
  `000226_enforce_orchestration_domain_snapshot_consistency`,
  `000227_workflow_run_execution_manifest`, and
  `000228_worker_spec_optional_model_binding`.
- AI resource configuration revisions now distinguish runtime facts from
  operational metadata. Validation state, display names, enable state, and
  credential rotation preserve resource/connection revisions; BaseURL, Model
  ID, modalities, and capabilities advance them. Validation commits compare
  the encrypted credential identity under a row lock, so a probe started before
  rotation cannot approve the new credential. Runtime updates and deletes also
  compare `updated_at`, so an older request cannot overwrite or delete a newer
  credential, validation, enable-state, or display-name write. Focused domain,
  repository, service, race, Connect API, WorkerCreation, and AgentPod model
  binding suites pass, including rotation/revalidation, disable/re-enable, and
  stale update/delete paths.
- An independent review found two P1 lost-update/delete races after operational
  metadata stopped advancing revisions. The revision-plus-`updated_at` CAS and
  rollback assertions closed both; the follow-up review found no remaining
  P0/P1 in this slice.
- `worker_spec_snapshots` is a strict V1 JSON contract with PostgreSQL
  constraints and strict Go decoding. Adding resolved dependencies as optional
  V1 data would create a second runtime path; changing the required V1 shape
  would break version semantics. The next shared migration must introduce a
  versioned artifact or WorkerSpec V2, deterministically backfill what can be
  proved, and fail closed for rows that require guessing.
- Artifact V1 plus `WorkerTemplateBuild` now revalidate raw Definition JSON and
  AgentFile, materialize typed resolver facts internally, enforce exact
  reference closure and pre-allocation budgets for collection cardinality,
  strings, enum aliases, recursive JSON trees, and custom marshalers, and expose
  a pure `DecodeApplyPlan` boundary for future transactional Secret
  authorization and persistence.
- The dependency artifact, Definition snapshot, and build artifact packages pass
  focused tests, race tests, vet, and full Backend compile. Final independent
  review found no remaining module-local P0/P1/P2.
- Resource-managed Experts now export their active declaration and open a
  locked-identity `ResourceEditorShell`; legacy Experts retain the direct edit
  drawer. The UI no longer offers the legacy delete path that the Backend
  rejects for resource-managed Experts.
- Plan success now parses and validates `canonical_json`, then atomically adopts
  the canonical Draft, YAML source, and Plan. Missing or malformed canonical
  content fails the Plan instead of retaining a stale local Draft.
- Apply is enabled only for protocol `PLAN_STATUS_PENDING`. Applied, cancelled,
  expired, and unspecified Plan states remain visible in the badge and Plan
  review but cannot be consumed.
- Editor identity now includes organization, Kind, session key, and locked
  resource identity. Local and persisted editor sessions reset or isolate state
  across organization changes.
- Worker options are keyed and fetched by explicit organization. Reference
  fields only accept catalog choices and remain visible but read-only while
  loading, forbidden, empty, or unresolved. List/map structure, map keys, and
  handlers obey the same read-only contract.
- WorkerTemplate YAML Plan reloads organization-scoped Worker options and
  relevant reference catalogs before the RPC. Stale type, runtime,
  `optionsRevision`, catalog errors, and unresolved Draft references fail closed.
- The YAML editor and localized user guide now state the 256 KiB document and
  64 KiB physical-line limits and link directly to the YAML guide section.
- `GetResourceCapabilities` is wired through Proto, Go Connect, Rust services,
  WASM, and Web. Revision dialogs require both live source and Plan permission
  before export and do not construct an editable Draft when denied.
- Canonical Draft adoption recursively rejects unknown fields and typed `null`
  values for every Kind and uses exact Backend CPU wire names.
- Final review found and verified closure of a custom `json.Unmarshaler` null
  bypass across runtime resources and nested references; no P0/P1/P2 remain.
- Web passes 324 files and 2343 tests; lint, typecheck, Rust workspace, WASM,
  production Web build, Worker docs/runtime checks, and full Go packages pass.
- Production Playwright passes 16 resource-editor and WASM route scenarios.
  Desktop/mobile visual QA found no overflow, framework overlay, console error,
  failed response, or inaccessible action state.
- A serial production-browser regression passes 121 tests covering Autopilot
  realtime control, Runner capacity scheduling, Workflow reference protection,
  GoalLoop Apply gating, Mesh, and Channel paths.
- Runner MCP passes its complete live suite in 33 seconds, including automatic
  Pod placement, immutable applied identity, cross-Runner routing, Workflow
  resource Apply, authorization isolation, and redacted errors. Binding failure
  returns a tool error, and Runner capacity uses an atomic database claim.
- CI-equivalent `postgres:16` passes migration lineage, dirty-version rejection,
  `000226` revision/snapshot consistency, `000227` execution manifests, and
  `000228` optional model binding, including fail-closed and rollback cases.
- Production chunk scanning confirms marketing and auth chunks contain no WASM
  symbols while authenticated dashboard routes load the WASM runtime.
- Production deployment remains governed by the release owner and GitOps
  rollout policy; no manual production mutation is claimed here.

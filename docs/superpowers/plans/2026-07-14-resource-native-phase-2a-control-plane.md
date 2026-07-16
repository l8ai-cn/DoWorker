# Resource-Native Phase 2A Control Plane Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `subagent-driven-development` task-by-task.

**Goal:** Persist resource identities, immutable revisions, and expiring Plans
without adding a second runtime path.

**Architecture:** `orchestrationresource` remains the document contract.
`orchestrationcontrol` owns tenant scope, hashes, Plan lifecycle, and repository
ports. Target planners compile typed Specs into opaque, non-secret artifacts;
target-specific Apply remains outside the generic service. Browser clients use
the existing Connect-first HTTP path and Rust Core/WASM service boundary;
Runner gRPC remains exclusively the runtime control plane.

**Exit:** Validate and Plan are deterministic, tenant-scoped, concurrency-safe,
and persisted. No Worker Pod is launched in this sub-phase.

## Invariants

- Tenant organization and user come from authenticated scope.
- Manifest namespace must equal the authenticated organization slug.
- Resource identity is `(organization_id, api_version, kind, namespace, name)`.
- UID is server-generated UUID and never changes.
- Revision increments on every accepted Apply; generation increments only when
  canonical Spec changes.
- Resource version changes on every accepted head mutation.
- Revisions and Plans are immutable.
- Plans are bound to organization, actor, target identity, base resource
  version, draft hash, resolved refs, compiled artifact, and expiry.
- Plan artifacts, diffs, refs, errors, events, and logs never contain secret
  values.
- Generic code never creates Pods or reconstructs runtime state.

## BDD Scenarios

```text
Given a member submits a valid new resource draft
When ValidateResource runs
Then no row is written and typed validation issues are deterministic

Given semantic validation succeeds
When PlanResource runs
Then one immutable Plan stores exact hashes, refs, diff, base state, and expiry

Given the same canonical draft is planned twice
When no dependency or base state changed
Then both plans have equal draft/artifact hashes but distinct plan IDs

Given a resource head changes after planning
When target Apply later consumes the Plan
Then Apply fails stale-plan before writing a revision or snapshot

Given a plan belongs to another organization or actor
When it is read or consumed
Then the operation is denied without revealing plan contents
```

### Task 1: Migration And Database Contract

**Files:**
- Create: `backend/migrations/000211_orchestration_resources.up.sql`
- Create: `backend/migrations/000211_orchestration_resources.down.sql`
- Create: `backend/migrations/000212_orchestration_resource_integrity.up.sql`
- Create: `backend/migrations/000212_orchestration_resource_integrity.down.sql`
- Create: `backend/migrations/orchestration_resources_test.go`
- Create: `backend/migrations/orchestration_resources_postgres_test.go`

- [ ] Add `orchestration_resources` head table with organization FK, UUID UID,
  strict API/kind/name/namespace checks, display metadata, generation,
  resource_version, active revision, creator/updater, status JSONB, timestamps,
  and scoped unique identity.
- [ ] Add immutable `orchestration_resource_revisions` with canonical manifest,
  canonical Spec, resolved refs, digest, optional WorkerSpec snapshot FK,
  actor, revision, generation, and unique `(resource_id, revision)`.
- [ ] Add immutable `orchestration_resource_plans` with UUID plan ID, scope,
  operation, target identity, optional base head identity/version, draft hash,
  plan hash, canonical manifest, resolved refs, semantic diff, issues,
  artifact kind/JSON, options revision, expiry, consumption result, and actor.
- [ ] Add DB checks for JSON object/array shapes, positive counters, SHA-256
  digests, expiry ordering, one consumption result, and scoped FKs.
- [ ] Add update/delete prevention triggers for revisions and plan payloads;
  only consumption columns may transition once.
- [ ] Verify SQLite migration tests and PostgreSQL-specific constraints.

### Task 2: Domain State And Canonical Hashes

**Files:**
- Create: `backend/internal/domain/orchestrationcontrol/resource.go`
- Create: `backend/internal/domain/orchestrationcontrol/plan.go`
- Create: `backend/internal/domain/orchestrationcontrol/errors.go`
- Create: `backend/internal/domain/orchestrationcontrol/hash.go`
- Test: matching `_test.go` files

- [ ] Define `Scope`, `ResourceHead`, `ResourceRevision`, `Plan`, `PlanIssue`,
  `SemanticChange`, `ResolvedReference`, and operation/status enums.
- [ ] Validate identifiers, UUIDs, revisions, generations, timestamps, digest
  grammar, issue paths, and immutable consumption transitions.
- [ ] Hash canonical compact JSON with lowercase `sha256:` output.
- [ ] Compute plan hash from operation, scope, target, base state, draft hash,
  sorted resolved refs, artifact digest, and options revision.
- [ ] Prove map ordering, input ordering, whitespace, and YAML source formatting
  cannot change canonical hashes.

### Task 3: Repository Ports And GORM Implementation

**Files:**
- Create: `backend/internal/service/orchestrationcontrol/contracts.go`
- Create: `backend/internal/infra/orchestration_resource_repo.go`
- Create: focused read/write/record mapping files
- Test: `backend/internal/infra/orchestration_resource_repo_test.go`

- [ ] Add tenant-scoped Get/List head and Get/List revision operations.
- [ ] Add CreatePlan/GetPlan with exact actor and organization scope.
- [ ] Add a transaction entry point that locks the target head and plan for a
  later target Apply without exposing raw `*gorm.DB` to services.
- [ ] Return typed not-found, conflict, stale, expired, consumed, and corrupt
  record errors; never translate corruption to empty results.
- [ ] Test cross-tenant lookup, duplicate identity, immutable rows, lock order,
  stale base state, plan replay, and corrupt JSON rejection.

### Task 4: Validate And Plan Service

**Files:**
- Create: `backend/internal/service/orchestrationcontrol/service.go`
- Create: `validation.go`, `planning.go`, `authorization.go`
- Test: focused service tests

- [ ] Register target planners by exact `TypeMeta`; reject duplicate or missing
  planners at startup.
- [ ] Decode JSON/YAML submissions through the Phase 1 Registry.
- [ ] Enforce authenticated namespace and create/update authorization before
  resolving references.
- [ ] Validate without persistence and return field-addressed blocking issues.
- [ ] Plan by resolving every reference to immutable identity, compiling a
  non-secret artifact, computing semantic diff and hashes, and persisting one
  immutable expiring Plan.
- [ ] Re-read the existing head after target planning and reject races before
  persisting the Plan.
- [ ] Test dependency changes, unauthorized refs, stale options, creator/admin
  update rules, secret redaction, and deterministic issue ordering.

### Task 5: Connect API And Wiring

**Files:**
- Create: `proto/orchestration_resource/v1/orchestration_resource.proto`
- Create: `backend/internal/api/connect/orchestration_resource/`
- Modify: `backend/cmd/server/connect_mount.go`
- Modify: service-container initialization files
- Modify: generated Go/TypeScript protobuf mirrors
- Create: Rust API client, service bridge, and WASM export files
- Test: Connect handler, interceptor, Rust, and tenant-scope tests

- [ ] Add `ValidateResource`, `PlanResource`, `GetResource`,
  `ListResources`, `ExportResource`, and `GetResourcePlan`.
- [ ] Accept either JSON or YAML source with an explicit format enum.
- [ ] Return canonical JSON/YAML, typed issues, semantic diff, plan identity,
  expiry, hashes, and redacted resolved refs.
- [ ] Map validation to InvalidArgument, authorization to PermissionDenied,
  missing rows to NotFound, stale/conflict to Aborted, and unavailable
  dependencies to Unavailable.
- [ ] Mount manual unary Connect handlers under
  `/proto.orchestration_resource.v1.OrchestrationResourceService/*`, using the
  existing auth interceptor and `ResolveOrgScope`; do not add a REST fallback or
  expose this browser API through Runner gRPC.
- [ ] Add Rust Core `api-client` calls, services wire bridge, and WASM exports so
  the web client does not bypass the Rust business-logic SSOT.
- [ ] Verify cross-organization and actor-bound Plan rejection through the real
  interceptor stack and run Go/TypeScript proto generation checks.

### Task 6: Phase Gate

- [ ] Run domain, repository, migration, service, Connect, race, vet, proto,
  and diff checks.
- [ ] Independent spec and quality reviews have no unresolved P0/P1 findings.
- [ ] Record exact commands and hashes in the durable goal document.

# Worker Platform Redesign Progress

## Goal

Implement the approved Worker creation and publishing design from domain contract through browser verification directly on the shared `main` branch without overwriting concurrent task changes.

## Controller

- Trigger: active Codex goal requested by the user.
- Observation per cycle: first unchecked task in the current implementation plan plus its deterministic test result.
- Action per cycle: complete one TDD task, then run spec and quality review before advancing.
- Durable memory: this file, phase implementation plans, git commits, and test logs.
- Context reconstruction: read this file and the active phase plan at the start of each continuation.

## Guardrails

- Maximum implementation cycles: 48 reviewed tasks.
- Soft token ceiling: 3,000,000 goal tokens; inspect goal usage after every phase.
- Active wall-clock ceiling before human checkpoint: 10 hours.
- No-progress exit: stop after three consecutive attempts with the same root cause and no new verified state.
- Integrity rule: tests, coverage expectations, CI checks, and schema validators may not be weakened to obtain green output.
- External action boundary: do not merge, deploy, publish packages, or modify production configuration without a separate user instruction.
- Escalation: stop for missing credentials, destructive migration ambiguity, or a product decision that changes the approved object model.

## Machine-Checkable Completion

All conditions must hold:

1. Every phase plan checkbox is complete and each implementation task has spec and quality approval.
2. Scoped backend tests for runner eligibility, resource visibility, WorkerSpec persistence, Expert publishing, and Skill publishing pass.
3. Proto generation and contract tests pass with no uncommitted generated drift.
4. Rust Core Worker/Expert state tests pass.
5. Web unit, lint, and type checks pass for affected targets.
6. Browser E2E covers create success, incompatibility, loading/error, publish Expert, and publish selected Skills.
7. Browser console has no relevant errors and create/publish network requests match WorkerSpec.
8. Full relevant Bazel suites pass, or every pre-existing unrelated failure is reproduced from `origin/main` and documented.
9. The final diff contains no file over project limits due to this change and no unrelated formatting/refactor churn.

## Phase Status

| Phase | Deliverable | Status |
| --- | --- | --- |
| 0 | Authorization and current-flow correctness | Complete |
| 1 | WorkerSpec V1 contract and immutable snapshot | Complete foundation; Pod/Expert linkage is tracked in the active implementation plan |
| 2 | Runtime image, compute target, deployment, resource profile | Runtime catalog, scoped resolution, preflight, immutable snapshot persistence, and Pod resume linkage complete |
| 3 | Canonical four-step web create workflow | Pending |
| 4 | Runtime Expert and Skill publishing | Pending |
| 5 | Migration, full regression, browser QA, documentation | Pending |

## Environment State

- Repository: `/Users/wwyz/Documents/code/AgentsMesh`
- Branch: `main`
- Current committed base: `aca06b3c6` (`feat(worker): define creation wire contract`)
- Local `main` is synchronized with `origin/main`.
- Shared worktree contains concurrent AI Resource, Loop, Grok, and Marketplace changes. Every commit must use an exact allowlist and a reviewed index.
- Real PostgreSQL migration tests use an isolated schema in the main dev database at `127.0.0.1:10002`.
- Browser verification remains required for the mobile Pod path and the final four-step Worker workflow.

## Integration Dependency

All implementation now occurs on the shared `main` worktree. Do not create or switch branches/worktrees. Preserve file ownership across concurrent tasks, never reset unrelated edits, and inspect the exact staged file list before every commit.

## Change Log

- 2026-07-10: Goal created, isolated worktree created, design approved and persisted.
- 2026-07-10: Environment initialization attempted; deterministic external mirror blocker recorded.
- 2026-07-10: Explicit Runner resolution completed with TDD, full Runner regression, spec review, and quality review.
- 2026-07-10: PodOrchestrator explicit placement gate completed with TDD, full AgentPod regression, spec review, and quality review.
- 2026-07-10: Repository ID and slug access resolvers completed; ambiguous multi-provider slugs now fail explicitly after TDD and two review loops.
- 2026-07-10: Worker repository enforcement completed: explicit/auto/resume paths resolve once before persistence, preserve empty AgentFile overrides, and passed cross-task spec and quality reviews.
- 2026-07-10: Repository validation transport mappings completed: REST, Connect, and MCP return fixed client errors; wrapped and joined errors remain redacted; all transport regressions and independent reviews passed.
- 2026-07-10: AI-model visibility boundary completed: scoped rows are selected before credential decryption, with service and real GORM coverage plus independent reviews.
- 2026-07-10: Token-budget checkpoint inspected at 1.53M; the user explicitly accepted high token use, so the soft ceiling was raised to 3M while verification and time limits remain unchanged.
- 2026-07-10: Full backend build exposed 15 pre-existing Connect BUILD files that duplicate generated `*.amesh.go` outputs as literal sources; Phase 0D tracks the surgical repair.
- 2026-07-10: Connect generated-source repair completed: all 15 duplicate literals were removed, all 16 converter targets remain in the server graph, and the full backend server build plus independent reviews passed.
- 2026-07-10: Explicit Worker and Session model selection now propagates authenticated user and organization scope; old unscoped model lookup is no longer called, with independent reviews passed.
- 2026-07-10: Virtual-key create and scoped resolution boundaries completed: model visibility precedes minting, key ownership is exact, model visibility is rechecked, and usage-touch failures are not swallowed.
- 2026-07-10: Worker Virtual Key binding now propagates exact key, organization, and user scope; the obsolete unscoped credential resolver was removed after spec and quality approval.
- 2026-07-10: Shared Pod protocol landed as `5a52ced6b`; mobile access Plan B landed as `25ad3fe15`.
- 2026-07-10: WorkerSpec validation was hardened to reject missing required fields, validate type schemas, and persist immutable model resource/connection revisions plus provider/model identity.
- 2026-07-10: Runtime selection was reduced from six repository reads to one atomic repository operation, with focused tests passing.
- 2026-07-10: Migration `000197_worker_spec_model_binding` passed static and real PostgreSQL up/down tests, including empty-table guards and invalid binding rejection.
- 2026-07-10: Unified AI Resource and credential cutover completed and was pushed through `fda060c83`.
- 2026-07-10: Worker execution moved permanently to shared `main`; no Worker worktree may be created or used for writes.
- 2026-07-10: The approved Worker creation/publishing spec was converted into `2026-07-10-worker-creation-publishing.md`; execution is inline on `main` with commit-level checkpoints.
- 2026-07-10: Task 1 completed: immutable Codex/Claude/Gemini image catalog, runner-pool and managed-Kubernetes target capabilities, server-owned resource profiles, four-step Worker draft/preflight/fill/publish wire contract, and Go/Rust/TypeScript generation checks all passed.
- 2026-07-10: Task 2 completed: organization-scoped WorkerSpec resolution, exact model and runtime metadata, removal of model-managed fields from Worker type and runtime EnvBundle contracts, interaction-mode and package preflight, deterministic AgentFile compilation, atomic snapshot/Pod/config persistence, same-organization database constraints, exact EnvBundle and Skill runtime loading, and fail-closed fresh-create/resume definition validation all passed focused, full package, Proto generation, and real PostgreSQL tests.

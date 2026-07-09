# Worker Platform Redesign Progress

## Goal

Implement the approved Worker creation and publishing design from domain contract through browser verification without overwriting the dirty `feat/worker-config-lifecycle` worktree.

## Controller

- Trigger: active Codex goal requested by the user.
- Observation per cycle: first unchecked task in the current implementation plan plus its deterministic test result.
- Action per cycle: complete one TDD task, then run spec and quality review before advancing.
- Durable memory: this file, phase implementation plans, git commits, and test logs.
- Context reconstruction: read this file and the active phase plan at the start of each continuation.

## Guardrails

- Maximum implementation cycles: 48 reviewed tasks.
- Soft token ceiling: 1,000,000 goal tokens; inspect goal usage after every phase.
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
| 0 | Authorization and current-flow correctness | In progress (Runner resolver complete) |
| 1 | WorkerSpec V1 contract and immutable snapshot | Pending |
| 2 | Runtime image, compute target, deployment, resource profile | Pending |
| 3 | Canonical four-step web create workflow | Pending |
| 4 | Runtime Expert and Skill publishing | Pending |
| 5 | Migration, full regression, browser QA, documentation | Pending |

## Environment State

- Worktree: `/Users/wwyz/Documents/code/AgentsMesh-Worktrees/codex-worker-creation-redesign`
- Branch: `codex/worker-creation-redesign`
- Base: `origin/main` at `a7067af2e68da9a3908901249f5b49847a6d5a7c`
- Runner and mock-agent Bazel builds: passed during initialization.
- DoAgent cache: reused from the same base worktree after architecture and SHA-256 verification.
- Full dev startup: blocked after three retries by Debian package mirror HTTP 502 while building the Aider image.
- E2E environment retry point: before Phase 3 browser work and again before terminal verification.

## Integration Dependency

The main worktree has uncommitted lifecycle, Proto, Runner ACP, and frontend changes on `feat/worker-config-lifecycle`. Do not copy, reset, stage, or modify them. Before starting any lifecycle implementation, inspect whether that branch has a new commit and integrate by reviewed commit rather than filesystem copying.

## Change Log

- 2026-07-10: Goal created, isolated worktree created, design approved and persisted.
- 2026-07-10: Environment initialization attempted; deterministic external mirror blocker recorded.
- 2026-07-10: Explicit Runner resolution completed with TDD, full Runner regression, spec review, and quality review.
- 2026-07-10: PodOrchestrator explicit placement gate completed with TDD, full AgentPod regression, spec review, and quality review.

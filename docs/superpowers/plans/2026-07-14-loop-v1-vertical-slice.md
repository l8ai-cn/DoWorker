# Loop V1 Vertical Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a production-integrated Loop workbench where Blockly and LoopScript edit the same GoalLoop V1 program and the backend can compile, create, and start the real GoalLoop.

**Architecture:** A restricted LoopScript parser/compiler in Go is authoritative. A typed flat V1 AST is transported through the existing GoalLoop Connect service, stored in Rust Core `AppState`, and projected into CodeMirror and Blockly views. Running always recompiles source on the backend before using the existing GoalLoop service.

**Tech Stack:** Go, ConnectRPC, protobuf, Rust/WASM, React 19, Next.js, CodeMirror 6, Blockly 13, Vitest, Playwright.

---

## Scope

Included:

- one Loop, one Worker snapshot, one explicit Repeat, one Agent task, one command Verifier;
- hard iteration/token/time/no-progress/same-error limits and pause/fail policy;
- stable `@id` metadata, canonical formatting, diagnostics, source/AST round-trip;
- block-to-code and code-to-block synchronization with code-error lockout;
- backend compile and compile-then-create/start using existing GoalLoop runtime;
- desktop/mobile browser validation and an authenticated integration run.

Excluded:

- ProgramVersion persistence, custom block publishing, collaboration, V2 StepRun Controller;
- branch, approval, Skill/MCP calls, parallelism and arbitrary TypeScript.

## Acceptance Scenarios

- Given a valid block program, when a field or connection changes, then Rust Core stores the new semantic draft and the code view shows canonical LoopScript with unchanged node IDs.
- Given valid LoopScript, when the user edits the task or limits, then backend compilation succeeds and Blockly projects the equivalent tree.
- Given invalid LoopScript, when compilation returns diagnostics, then source remains editable, blocks keep the last valid AST read-only, and Run is disabled.
- Given a valid program and available Worker snapshot, when Run is clicked, then the backend recompiles source, creates a GoalLoop, starts it, and returns its real slug/status.
- Given unsupported syntax or a stale/unknown Worker snapshot, when Compile or Run is called, then the request fails explicitly without creating a GoalLoop.

### Task 1: Authoritative LoopScript Core

**Files:**
- Create: `backend/internal/loopscript/token.go`
- Create: `backend/internal/loopscript/lexer.go`
- Create: `backend/internal/loopscript/ast.go`
- Create: `backend/internal/loopscript/parser.go`
- Create: `backend/internal/loopscript/formatter.go`
- Create: `backend/internal/loopscript/compiler.go`
- Test: `backend/internal/loopscript/*_test.go`

- [x] Add failing lexer/parser tests for the canonical source and stable diagnostic codes.
- [x] Implement tokens, parser and typed AST without permissive recovery.
- [x] Add failing formatter round-trip and compiler boundary tests.
- [x] Implement canonical formatting and GoalLoop launch compilation.
- [x] Run `go test ./backend/internal/loopscript/...`.

### Task 2: GoalLoop Connect Contract

**Files:**
- Modify: `proto/goalloop/v1/goalloop.proto`
- Modify: `backend/internal/api/connect/goalloop/goalloop.go`
- Create: `backend/internal/api/connect/goalloop/goalloop_loop_program.go`
- Test: `backend/internal/api/connect/goalloop/goalloop_loop_program_test.go`
- Generated: `proto/gen/go/goalloop/v1/*`
- Generated: `proto/gen/ts/goalloop/v1/goalloop_pb.ts`
- Generated: `clients/core/crates/proto/goalloop/src/lib.rs`

- [x] Add failing handler tests for compile, run, invalid source and unavailable Worker.
- [x] Add `CompileLoopProgram` and `RunLoopProgram` messages and RPCs.
- [x] Implement compile and compile-create-start handlers using tenant identity.
- [x] Regenerate Go, TypeScript and Rust protobuf code.
- [x] Run focused Go handler and proto contract tests.

### Task 3: Rust Core Loop State

**Files:**
- Create: `clients/core/crates/state/src/loop_builder_state.rs`
- Create: `clients/core/crates/state/src/loop_builder_state_tests.rs`
- Modify: `clients/core/crates/state/src/lib.rs`
- Modify: `clients/core/crates/state/src/app_state.rs`
- Modify: `clients/core/crates/api-client/src/modules/goalloop_connect.rs`
- Modify: `clients/core/crates/services/src/goal_loop_service.rs`
- Modify: `clients/core/crates/wasm/src/api.rs`
- Modify: `clients/core/crates/wasm/src/service_goal_loop.rs`
- Create: `clients/core/crates/wasm/src/state_loop_builder.rs`
- Modify: `clients/core/crates/wasm/src/lib.rs`
- Modify: `packages/service-interface/src/index.ts`
- Modify: `packages/service-runtime/src/service-getters.ts`
- Modify: `packages/service-runtime/src/index.ts`

- [x] Add failing tests for last-valid AST, invalid source lockout and reset-on-org-switch.
- [x] Implement `LoopState` as the semantic draft owner in `AppState`.
- [x] Add API client/service methods for compile and run.
- [x] Expose WASM state selectors and mutation methods.
- [x] Register the Loop state view in the cross-platform service provider.
- [x] Run focused Cargo tests and `pnpm run build:wasm`.

### Task 4: Web Projection And Editors

**Files:**
- Create: `clients/web/src/components/loop-builder/*`
- Create: `clients/web/src/lib/api/connect/loopProgramConnect.ts`
- Create: `clients/web/src/lib/viewModels/loop-program.ts`
- Create: `clients/web/src/app/(dashboard)/[org]/loops/workbench/page.tsx`
- Modify: `clients/web/src/components/goal-loops/GoalLoopPage.tsx`
- Modify: `clients/web/src/lib/wasm-getters.ts`
- Modify: `package.json`
- Modify: `pnpm-lock.yaml`
- Test: `clients/web/src/components/loop-builder/__tests__/*`

- [x] Add Blockly 13 at the workspace root and keep it lazy-loaded on the Loop route.
- [x] Add failing projection tests for AST-to-blocks, blocks-to-source and stable node IDs.
- [x] Implement block catalog, projector, semantic edit adapter and CodeMirror editor.
- [x] Implement single-writer mode, diagnostics, last-valid read-only blocks and run result.
- [x] Add the Loop entry action to the existing GoalLoop page.
- [x] Run focused Vitest, web typecheck and lint.

### Task 5: Integration And Browser Evidence

**Files:**
- Create: `clients/web/e2e-playwright/tests/loops/loop-workbench.spec.ts`
- Update: `docs/superpowers/plans/2026-07-14-loop-v1-progress.md`

- [x] Start the development stack and confirm backend, WASM and Web health.
- [x] Exercise block edit to code, code edit to blocks, invalid-code recovery and Run.
- [x] Verify the created GoalLoop through the list/get API and capture its real status.
- [x] Check console/network errors and take desktop/mobile screenshots.
- [x] Run focused Go, Cargo, Vitest, typecheck, lint and E2E suites.
- [x] Request final spec and code-quality review; fix every blocking finding.
- [ ] Commit, push, fetch the remote branch and verify the full commit SHA is visible.

## Completion Gate

All five acceptance scenarios must have deterministic test or browser evidence. Unsupported syntax, compile errors, missing Worker snapshots, backend failures and browser runtime errors are blocking; none may be hidden by client-side compilation or fallback behavior.

# Loop AI And Generic Block Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development or executing-plans task by task. Every behavior change starts with a failing test.

**Goal:** Complete the approved Loop V1.1/V1.2 scope: host-injected language resources, a reusable block-programming workbench shell, and AI-generated LoopScript proposals that require backend validation and user confirmation.

**Architecture:** The Go LoopScript compiler remains authoritative. AI generation resolves one explicit chat model resource, requests a strict JSON `{source}` proposal, compiles it on the backend, and returns only a valid canonical program. React shows a change preview; confirmation writes the proposal into Rust Core and applies the matching compile response. Worker snapshots remain exclusive to the run dialog.

**Tech Stack:** Go, ConnectRPC, protobuf, Rust/WASM, React 19, Next.js, next-intl, Blockly 12.5, CodeMirror 6, Vitest.

---

### Task 1: AI Loop Proposal Contract

**Files:**
- Modify: `proto/goalloop/v1/goalloop.proto`
- Modify: `backend/internal/api/connect/goalloop/goalloop.go`
- Create: `backend/internal/api/connect/goalloop/goalloop_ai_generation.go`
- Test: `backend/internal/api/connect/goalloop/goalloop_ai_generation_test.go`
- Modify: `backend/cmd/server/worker_creation_init.go`
- Modify: `backend/cmd/server/connect_mount.go`

- [x] Write a failing Connect handler test that requires an explicit model resource, passes no Worker snapshot to the generator, rejects malformed or invalid generated source, and returns a compiled canonical program for valid `{source}` JSON.
- [x] Run the focused Go test and confirm it fails because the RPC and handler are absent.
- [x] Add `GenerateLoopProgramRequest` and `GenerateLoopProgram` to the proto contract.
- [x] Inject the existing provider-neutral generator and AI resource resolver through a server option.
- [x] Resolve only chat/text-generation resources using the explicit model resource ID and supported protocol adapters.
- [x] Build prompts that forbid Worker selection, verifier removal, budget removal, secrets, Blockly JSON, and non-JSON output.
- [x] Strictly decode `{source}`, compile it with the existing parser/formatter, and return a normal compile response without mutating state.
- [x] Regenerate Go, TypeScript, Rust, and amesh mappings; run backend, proto, Rust, and codegen drift checks.

### Task 2: Reusable Multilingual Workbench

**Files:**
- Create: `clients/web/src/components/block-programming/BlockProgrammingWorkbench.tsx`
- Create: `clients/web/src/components/loop-builder/loop-workbench-messages.ts`
- Modify: `clients/web/src/components/loop-builder/loop-workbench.tsx`
- Modify: `clients/web/src/components/loop-builder/loop-block-catalog.ts`
- Modify: `clients/web/src/components/loop-builder/loop-blockly-canvas.tsx`
- Modify: `clients/web/src/components/loop-builder/loop-quick-insert.tsx`
- Modify: `clients/web/src/components/loop-builder/loop-status-panel.tsx`
- Modify: `clients/web/src/components/loop-builder/loop-runtime-dialog.tsx`
- Modify: `clients/web/src/components/loop-builder/loop-workbench-toolbar.tsx`
- Modify: `clients/web/src/messages/en/app.json`
- Modify: `clients/web/src/messages/zh/app.json`
- Test: `clients/web/src/components/loop-builder/__tests__/loop-block-projection.test.ts`
- Test: `clients/web/src/components/loop-builder/__tests__/loop-workbench-messages.test.ts`

- [x] Write failing tests proving English and Chinese labels produce identical block types/source and that no toolbox contains Worker.
- [x] Write a failing render test for the reusable shell regions and injected labels.
- [x] Extract the split canvas/editor/status layout into a host-controlled generic component with no Loop or Worker domain dependency.
- [x] Build block definitions and toolbox categories from injected messages while keeping stable block type IDs and AST node IDs.
- [x] Move toolbar, quick insert, runtime, diagnostics, run status, and empty/error labels to `next-intl`.
- [x] Keep Chinese output for the current Chinese locale and English fallback for other locales.

### Task 3: AI Preview And Confirm

**Files:**
- Create: `clients/web/src/components/loop-builder/loop-ai-assistant-dialog.tsx`
- Create: `clients/web/src/components/loop-builder/loop-ai-proposal-preview.tsx`
- Modify: `clients/web/src/components/loop-builder/use-loop-workbench.ts`
- Modify: `clients/web/src/components/loop-builder/loop-workbench.tsx`
- Modify: `clients/web/src/components/loop-builder/loop-workbench-toolbar.tsx`
- Modify: `clients/web/src/lib/api/connect/loopProgramConnect.ts`
- Test: `clients/web/src/components/loop-builder/__tests__/loop-ai-assistant-dialog.test.tsx`
- Test: `clients/web/src/lib/api/__tests__/loopAIConnect.test.ts`

- [x] Write failing tests for resource loading, missing-resource state, generation error, unchanged proposal, preview, cancel, and confirm.
- [x] Add the generated RPC to the Rust service/WASM path and TypeScript Connect adapter.
- [x] List selectable organization chat resources and require an explicit model selection.
- [x] Generate without mutating the current source; show current and proposed canonical LoopScript in a preview.
- [x] On confirmation, pass the original validated response to Rust Core. Rust applies it atomically only when its revision still matches, then advances the source and semantic revisions.
- [x] Never start a Loop, choose a Worker snapshot, or alter a verifier automatically.
- [x] Run focused tests, full Web tests, typecheck, production build, browser desktop checks, and console/network inspection.

### Task 4: Deterministic Explain And Targeted Repair

**Files:**
- Create: `backend/internal/service/goalloop/draft_repair.go`
- Create: `backend/internal/service/goalloop/draft_repair_target.go`
- Create: `clients/web/src/components/loop-builder/loop-ai-repair-form.tsx`
- Create: `clients/web/src/components/loop-builder/request-loop-ai-proposal.ts`
- Modify: `proto/goalloop/v1/goalloop.proto`
- Modify: `clients/web/src/components/loop-builder/loop-status-panel.tsx`

- [x] Derive explanations from the authoritative compiled AST without a model call.
- [x] Recompute the selected diagnostic on the backend instead of trusting client details.
- [x] Limit repair to one supported integer field and request strict `{"value": integer}` JSON.
- [x] Apply the typed patch, recompile the complete program, and return a preview without mutating state.
- [x] Require explicit user confirmation and matching Rust Core revision before applying.
- [x] Cover missing-model, invalid proposal, stale revision, unsupported diagnostic, preview, cancel, and confirm behavior.

### Task 5: Release Verification

- [x] Run backend, Runner, Relay, Rust workspace, Web tests, typecheck, build, and focused artifact tests.
- [x] Execute an empty PostgreSQL migration through the branch tip and confirm `dirty=false`.
- [x] Verify the local Chinese workbench, deterministic explanation, diagnostic targeting, and disabled no-model states in a real browser.
- [ ] Rebase the release commit onto the latest `origin/main`, resolve overlaps, and repeat regression checks.
- [ ] Push the final branch, pass CI, deploy migration-first through GitOps, and verify production health.
- [ ] Run a new Seedance Worker to MP4 preview and a professional PPT Loop in production.

# Loop Blockly MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an isolated browser MVP that authors, validates, compiles, persists, and simulates a single-Worker Goal Loop through Blockly.

**Architecture:** Blockly owns visual editing only. A workspace adapter creates a typed draft, a pure compiler validates it and emits canonical Loop JSON, and React owns workbench presentation. Custom blocks are deterministic text macros persisted locally.

**Tech Stack:** React 19, TypeScript, Vite, Blockly, Vitest, Testing Library.

---

### Task 1: Standalone Project And Compiler Contracts

**Files:**
- Create: `prototypes/loop-blockly-mvp/package.json`
- Create: `prototypes/loop-blockly-mvp/pnpm-workspace.yaml`
- Create: `prototypes/loop-blockly-mvp/index.html`
- Create: `prototypes/loop-blockly-mvp/tsconfig.json`
- Create: `prototypes/loop-blockly-mvp/vite.config.ts`
- Create: `prototypes/loop-blockly-mvp/src/domain/loop-types.ts`
- Create: `prototypes/loop-blockly-mvp/src/domain/compile-loop.test.ts`
- Create: `prototypes/loop-blockly-mvp/src/domain/compile-loop.ts`

- [x] Write tests for a complete draft, missing required sections, loose blocks, and invalid limits.
- [x] Run `pnpm test` and confirm failure because `compileLoop` is absent.
- [x] Implement typed draft, diagnostics, canonical result, and deterministic compilation.
- [x] Run `pnpm test` and confirm compiler tests pass.

Core API:

```ts
export function compileLoop(draft: LoopDraft): CompileResult;

export interface CompileResult {
  diagnostics: Diagnostic[];
  program?: GoalLoopProgram;
  executionBlockIds: string[];
}
```

### Task 2: Blockly Language And Workspace Adapter

**Files:**
- Create: `prototypes/loop-blockly-mvp/src/blockly/block-catalog.ts`
- Create: `prototypes/loop-blockly-mvp/src/blockly/block-theme.ts`
- Create: `prototypes/loop-blockly-mvp/src/blockly/toolbox.ts`
- Create: `prototypes/loop-blockly-mvp/src/blockly/workspace-to-draft.test.ts`
- Create: `prototypes/loop-blockly-mvp/src/blockly/workspace-to-draft.ts`
- Create: `prototypes/loop-blockly-mvp/src/custom-blocks/custom-block-definition.test.ts`
- Create: `prototypes/loop-blockly-mvp/src/custom-blocks/custom-block-definition.ts`

- [x] Write failing tests for Blockly connection mapping and `{{parameter}}` extraction.
- [x] Define typed Goal Loop, Worker, task, acceptance, verifier, limits, escalation blocks.
- [x] Implement workspace traversal without silently ignoring loose or unknown blocks.
- [x] Implement runtime registration and deterministic expansion of custom macro blocks.
- [x] Run all unit tests.

Required block checks:

```ts
task: "LoopInstruction";
acceptance: "LoopAcceptance";
worker: "LoopWorker";
verifier: "LoopVerifier";
limits: "LoopLimits";
escalation: "LoopEscalation";
```

### Task 3: Workbench UI And Persistence

**Files:**
- Create: `prototypes/loop-blockly-mvp/src/main.tsx`
- Create: `prototypes/loop-blockly-mvp/src/app.tsx`
- Create: `prototypes/loop-blockly-mvp/src/components/loop-workbench.tsx`
- Create: `prototypes/loop-blockly-mvp/src/components/blockly-canvas.tsx`
- Create: `prototypes/loop-blockly-mvp/src/components/block-inspector.tsx`
- Create: `prototypes/loop-blockly-mvp/src/components/custom-block-dialog.tsx`
- Create: `prototypes/loop-blockly-mvp/src/components/quick-insert-menu.tsx`
- Create: `prototypes/loop-blockly-mvp/src/components/output-panel.tsx`
- Create: `prototypes/loop-blockly-mvp/src/hooks/use-loop-workspace.ts`
- Create: `prototypes/loop-blockly-mvp/src/persistence/workspace-storage.ts`

- [x] Render the four-region workbench with stable canvas and panel dimensions.
- [x] Add inline and inspector parameter editing for selected blocks.
- [x] Open quick insert on canvas double-click and place the chosen block.
- [x] Add custom macro creation and refresh the “我的积木” toolbox category.
- [x] Persist workspace, custom definitions, Loop name, and selected output tab.
- [x] Represent empty, invalid, valid, saved, running, and simulation-complete states.

### Task 4: Simulation, Styling, And Browser Verification

**Files:**
- Create: `prototypes/loop-blockly-mvp/src/simulation/run-simulation.ts`
- Create: `prototypes/loop-blockly-mvp/src/styles/base.css`
- Create: `prototypes/loop-blockly-mvp/src/styles/workbench.css`
- Create: `prototypes/loop-blockly-mvp/README.md`

- [x] Implement cancellable simulation over `executionBlockIds`.
- [x] Highlight one block at a time and append structured evidence events.
- [x] Add responsive desktop and mobile workbench layouts using existing product-like restrained colors.
- [x] Run `pnpm test`, `pnpm typecheck`, and `pnpm build`.
- [x] Start Vite and exercise valid, invalid, custom block, persistence, and simulation paths.
- [x] Inspect console and network output and capture desktop/mobile screenshots.

Terminal verification:

```bash
cd prototypes/loop-blockly-mvp
pnpm test
pnpm typecheck
pnpm build
```

Browser acceptance:

- Blank canvas exposes one obvious primary action.
- Double-click opens quick insert at the pointer.
- Invalid programs cannot generate JSON or start simulation.
- Valid programs produce deterministic JSON.
- Custom macro parameters compile into the objective.
- Simulation highlights blocks and ends with verifier evidence.
- Desktop and mobile contain no overlap, clipping, or inaccessible controls.

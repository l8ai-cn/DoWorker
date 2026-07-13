# Loom Blockly MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an isolated browser MVP that authors, validates, compiles, persists, and simulates a single-Worker Goal Loop through Blockly.

**Architecture:** Blockly owns visual editing only. A workspace adapter creates a typed draft, a pure compiler validates it and emits canonical Loop JSON, and React owns workbench presentation. Custom blocks are deterministic text macros persisted locally.

**Tech Stack:** React 19, TypeScript, Vite, Blockly, Vitest, Testing Library.

---

### Task 1: Standalone Project And Compiler Contracts

**Files:**
- Create: `prototypes/loom-blockly-mvp/package.json`
- Create: `prototypes/loom-blockly-mvp/pnpm-workspace.yaml`
- Create: `prototypes/loom-blockly-mvp/index.html`
- Create: `prototypes/loom-blockly-mvp/tsconfig.json`
- Create: `prototypes/loom-blockly-mvp/vite.config.ts`
- Create: `prototypes/loom-blockly-mvp/src/domain/loop-types.ts`
- Create: `prototypes/loom-blockly-mvp/src/domain/compile-loop.test.ts`
- Create: `prototypes/loom-blockly-mvp/src/domain/compile-loop.ts`

- [ ] Write tests for a complete draft, missing required sections, loose blocks, and invalid limits.
- [ ] Run `pnpm test` and confirm failure because `compileLoop` is absent.
- [ ] Implement typed draft, diagnostics, canonical result, and deterministic compilation.
- [ ] Run `pnpm test` and confirm compiler tests pass.

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
- Create: `prototypes/loom-blockly-mvp/src/blockly/block-catalog.ts`
- Create: `prototypes/loom-blockly-mvp/src/blockly/block-theme.ts`
- Create: `prototypes/loom-blockly-mvp/src/blockly/toolbox.ts`
- Create: `prototypes/loom-blockly-mvp/src/blockly/workspace-to-draft.test.ts`
- Create: `prototypes/loom-blockly-mvp/src/blockly/workspace-to-draft.ts`
- Create: `prototypes/loom-blockly-mvp/src/custom-blocks/custom-block-definition.test.ts`
- Create: `prototypes/loom-blockly-mvp/src/custom-blocks/custom-block-definition.ts`

- [ ] Write failing tests for Blockly connection mapping and `{{parameter}}` extraction.
- [ ] Define typed Goal Loop, Worker, task, acceptance, verifier, limits, escalation blocks.
- [ ] Implement workspace traversal without silently ignoring loose or unknown blocks.
- [ ] Implement runtime registration and deterministic expansion of custom macro blocks.
- [ ] Run all unit tests.

Required block checks:

```ts
task: "LoomInstruction";
acceptance: "LoomAcceptance";
worker: "LoomWorker";
verifier: "LoomVerifier";
limits: "LoomLimits";
escalation: "LoomEscalation";
```

### Task 3: Workbench UI And Persistence

**Files:**
- Create: `prototypes/loom-blockly-mvp/src/main.tsx`
- Create: `prototypes/loom-blockly-mvp/src/app.tsx`
- Create: `prototypes/loom-blockly-mvp/src/components/loom-workbench.tsx`
- Create: `prototypes/loom-blockly-mvp/src/components/blockly-canvas.tsx`
- Create: `prototypes/loom-blockly-mvp/src/components/block-inspector.tsx`
- Create: `prototypes/loom-blockly-mvp/src/components/custom-block-dialog.tsx`
- Create: `prototypes/loom-blockly-mvp/src/components/quick-insert-menu.tsx`
- Create: `prototypes/loom-blockly-mvp/src/components/output-panel.tsx`
- Create: `prototypes/loom-blockly-mvp/src/hooks/use-loom-workspace.ts`
- Create: `prototypes/loom-blockly-mvp/src/persistence/workspace-storage.ts`

- [ ] Render the four-region workbench with stable canvas and panel dimensions.
- [ ] Add inline and inspector parameter editing for selected blocks.
- [ ] Open quick insert on canvas double-click and place the chosen block.
- [ ] Add custom macro creation and refresh the “我的积木” toolbox category.
- [ ] Persist workspace, custom definitions, Loop name, and selected output tab.
- [ ] Represent empty, invalid, valid, saved, running, and simulation-complete states.

### Task 4: Simulation, Styling, And Browser Verification

**Files:**
- Create: `prototypes/loom-blockly-mvp/src/simulation/run-simulation.ts`
- Create: `prototypes/loom-blockly-mvp/src/styles/base.css`
- Create: `prototypes/loom-blockly-mvp/src/styles/workbench.css`
- Create: `prototypes/loom-blockly-mvp/README.md`

- [ ] Implement cancellable simulation over `executionBlockIds`.
- [ ] Highlight one block at a time and append structured evidence events.
- [ ] Add responsive desktop and mobile workbench layouts using existing product-like restrained colors.
- [ ] Run `pnpm test`, `pnpm typecheck`, and `pnpm build`.
- [ ] Start Vite and exercise valid, invalid, custom block, persistence, and simulation paths.
- [ ] Inspect console and network output and capture desktop/mobile screenshots.

Terminal verification:

```bash
cd prototypes/loom-blockly-mvp
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


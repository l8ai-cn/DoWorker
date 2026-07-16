# Agent Tool Trace Collapse Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Agent Workspace tool execution read like Codex: contiguous commands and tool calls render as one collapsed activity summary, with individual steps and raw evidence available only after explicit expansion.

**Architecture:** Keep runtime and protocol contracts unchanged. Group contiguous `kind: "tool"` timeline items inside the shared `packages/agent-ui` presentation layer, render the group with a native accessible disclosure, and reuse `ToolActivityCard` for the expanded second level.

**Tech Stack:** React 18/19, TypeScript, Tailwind CSS, Vitest, Testing Library.

---

### Task 1: Pin the collapsed interaction contract

**Files:**
- Modify: `packages/agent-ui/src/ActivityTimeline.test.tsx`
- Modify: `packages/agent-ui/src/AgentWorkspace.test.tsx`

- [x] Add a test with two consecutive shell tools and assert one collapsed summary is rendered.
- [x] Assert command text and output are not visible before expansion.
- [x] Click the group summary and assert both individual tools become visible.
- [x] Update the Chinese workspace test to expect the localized collapsed summary.
- [x] Run `pnpm --dir packages/agent-ui exec vitest run src/ActivityTimeline.test.tsx src/AgentWorkspace.test.tsx` and confirm the new assertions fail before implementation.

### Task 2: Add shared tool-run grouping

**Files:**
- Create: `packages/agent-ui/src/ToolActivityGroup.tsx`
- Create: `packages/agent-ui/src/toolActivityGrouping.ts`
- Modify: `packages/agent-ui/src/ActivityTimeline.tsx`
- Modify: `packages/agent-ui/src/agentWorkspaceText.ts`

- [x] Add a pure grouping function that preserves timeline order and combines only contiguous tool items.
- [x] Add localized count phrases such as `运行了 3 个命令` and `读取了 2 个文件`.
- [x] Count structured file-change entries by their actual `changes` length.
- [x] Render each grouped run as a collapsed disclosure with running and failed status indicators.
- [x] Reuse `ToolActivityCard` inside the expanded group so its existing evidence disclosure remains the second level.
- [x] Keep messages, artifacts, reasoning, errors, and system rows outside tool groups.

### Task 3: Verify shared behavior and mobile rendering

**Files:**
- Test: `packages/agent-ui/src/ActivityTimeline.test.tsx`
- Test: `packages/agent-ui/src/AgentWorkspace.test.tsx`
- Test: `packages/agent-ui/src/ToolActivityCard.test.tsx`

- [x] Run the focused Agent UI tests and lint the touched files.
- [x] Reload `http://127.0.0.1:10030/` at 390×844.
- [x] Confirm completed commands are collapsed, expansion reveals the individual steps, and the composer remains usable.
- [x] Check page identity, horizontal overflow, framework overlays, current console errors, and screenshot evidence.

# Agent Workbench V2 Implementation Roadmap

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver one production-grade Agent Workbench across Web, Web User, React, plain mount, and iframe with real programming and media task proof.

**Architecture:** Security boundaries land first. Generated V2 contracts then unlock independent backend, Rust Core, Web User, and viewer work. The final gate runs real Runner tasks and browser verification in every host.

**Tech Stack:** Go, Rust/WASM, Protobuf, React, Vite, Next.js, SSE, Relay, Playwright.

---

### Task 1: Establish Safe Content Boundaries

**Plan:** `docs/superpowers/plans/2026-07-15-agent-workbench-phase-0a-content-security.md`

- [ ] Remove top-level untrusted HTML Blob navigation.
- [ ] Replace permissive static artifact sandbox with the shared static profile.
- [ ] Block automatic remote Markdown resource fetches.
- [ ] Exit gate: shared, Web, and Web User focused tests pass.

### Task 2: Isolate Live Preview

**Plan:** `docs/superpowers/plans/2026-07-15-agent-workbench-phase-0b-preview-security.md`

- [ ] Configure an exact dedicated preview origin.
- [ ] Add single-use bootstrap redemption and separate session cookie.
- [ ] Apply the `pod-live` frame profile.
- [ ] Exit gate: replay, origin, cookie, popup, navigation, and permissions tests pass.

### Task 3: Land Protocol And Runtime Primitives

**Plan:** `docs/superpowers/plans/2026-07-15-agent-workbench-protocol-runtime.md`

- [ ] Generate Go, Rust, and TypeScript V2 types.
- [ ] Add lossless cross-language fixtures.
- [ ] Add deterministic snapshot/delta reduction.
- [ ] Add exact tool and content registries.
- [ ] Exit gate: codegen is clean and protocol/runtime/registry tests pass.

### Task 4: Persist V2 On The Backend

**Plan:** `docs/superpowers/plans/2026-07-15-agent-workbench-backend-session.md`

- [ ] Persist projection, event stream, and command receipts atomically.
- [ ] Expose snapshot, replayable SSE, and idempotent command APIs.
- [ ] Remove the V1 session event route.
- [ ] Exit gate: database, API, reconnect, authorization, and receipt tests pass.

### Task 5: Build Result And Media Surfaces

**Plan:** `docs/superpowers/plans/2026-07-15-agent-workbench-result-media.md`

- [ ] Extract Web User viewers into shared pure renderers.
- [ ] Add ResultWorkbench, resource rail, and persistent surfaces.
- [ ] Add image comparison/edit, video, and presentation viewers.
- [ ] Exit gate: component, responsive, keepalive, action, and bundle-budget tests pass.

### Task 6: Migrate Web User And Embeds

**Plan:** `docs/superpowers/plans/2026-07-15-agent-workbench-web-user-embed.md`

- [ ] Replace lossy embedded projections with V2 snapshot/delta runtime.
- [ ] Integrate shared artifact and terminal adapters.
- [ ] Publish React, plain mount, and data-only iframe entries.
- [ ] Exit gate: Web User tests and embed production build pass.

### Task 7: Migrate Web To Rust Core SSOT

**Plan:** `docs/superpowers/plans/2026-07-15-agent-workbench-web-rust-core.md`

- [ ] Apply V2 state atomically in Rust and expose WASM byte methods.
- [ ] Replace TypeScript business reconstruction with a revision-only adapter.
- [ ] Remove V1 ACP state and Web projections after consumer search is empty.
- [ ] Exit gate: Cargo workspace, WASM build, Web tests, typecheck, and build pass.

### Task 8: Prove Real Tasks In Every Host

**Plan:** `docs/superpowers/plans/2026-07-15-agent-workbench-real-runner-e2e.md`

- [ ] Run deterministic real programming/HTML, image edit, PPT, and video tasks.
- [ ] Validate artifacts and revisions with machine tools.
- [ ] Verify desktop, 390px mobile, plain mount, and iframe.
- [ ] Exit gate: full acceptance script exits zero with screenshots, network evidence, and no new console errors.

## Execution Rules

The current worktree contains unrelated modified and untracked files. Every task stages only its explicit file list, reviews `git diff --cached`, and commits before the next task. Never use `git add -A`.

Existing design commits are `47c6a60eecb5c874eba5cdc3166ce43f00aec7b6` and `a04f54048`. Existing untracked `packages/agent-ui`, Web adapter, and Web User embed files are inputs to migrate, not evidence that a phase is complete.

No phase may add a V1 fallback, tool-name guess, mock artifact, static media substitution, or unsafe preview compatibility switch. A failed gate blocks the dependent phase.

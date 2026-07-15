# Agent Conversation Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** deliver a reusable Agent conversation core that can mount as a React component, a standalone web island, or a controlled iframe surface.

**Architecture:** `@do-worker/agent-ui` defines framework-neutral conversation
runtime contracts, React surfaces, DOM mounting, and strict postMessage bridge
utilities. `clients/web-user` supplies one explicit adapter over its existing
conversation state and an iframe entry to prove all three consumption forms.
The package boundary is transport-neutral so the future Rust/Connect runtime
replaces the adapter without changing consumer APIs.

**Tech Stack:** TypeScript, React 18/19 peer compatibility, Zustand adapter,
Vite, Vitest, Testing Library.

---

## File Structure

| Path | Responsibility |
| --- | --- |
| `packages/agent-ui/src/conversation-types.ts` | Public runtime and snapshot contracts |
| `packages/agent-ui/src/conversation-context.tsx` | Runtime provider and subscription hook |
| `packages/agent-ui/src/ConversationSurface.tsx` | Reusable transcript, loading/error, composer, and interrupt UI |
| `packages/agent-ui/src/mount-conversation.tsx` | Imperative page/island mounting API |
| `packages/agent-ui/src/iframe-bridge.ts` | Origin-checked parent/frame message protocol |
| `packages/agent-ui/src/index.ts` | Public package exports |
| `packages/agent-ui/src/*.test.tsx` | Package contract and rendering tests |
| `clients/web-user/src/lib/chatStoreConversationRuntime.ts` | Explicit legacy runtime adapter |
| `clients/web-user/src/iframe.tsx` | Iframe-only entry mounted by Vite |
| `clients/web-user/vite.config.ts` | Iframe entry build configuration |
| `clients/web-user/src/iframe-entry.test.tsx` | Adapter/bridge integration coverage |

## Task 1: Shared Runtime Contract

**Files:**
- Create: `packages/agent-ui/package.json`
- Create: `packages/agent-ui/tsconfig.json`
- Create: `packages/agent-ui/src/conversation-types.ts`
- Create: `packages/agent-ui/src/conversation-context.tsx`
- Create: `packages/agent-ui/src/conversation-context.test.tsx`
- Create: `packages/agent-ui/src/index.ts`
- Modify: `package.json`

- [x] Define `ConversationMessage`, `ConversationSnapshot`, `SendConversationMessage`,
  and `ConversationRuntime` with `open`, `send`, `interrupt`, `subscribe`, and
  `getSnapshot`.
- [x] Write a fake runtime test that proves `useConversationSnapshot` updates
  after a subscription notification and releases its listener on unmount.
- [x] Implement the provider and hook with `useSyncExternalStore`; never copy
  runtime state into React state.
- [x] Register `@do-worker/agent-ui` in the root workspace dependencies so
  first-party applications import a package name rather than a relative path.
- [x] Run `pnpm exec vitest run packages/agent-ui/src/conversation-context.test.tsx`
  and `pnpm exec tsc --noEmit -p packages/agent-ui/tsconfig.json`.

## Task 2: Reusable Conversation Surface and Web Mount

**Files:**
- Create: `packages/agent-ui/src/ConversationSurface.tsx`
- Create: `packages/agent-ui/src/mount-conversation.tsx`
- Create: `packages/agent-ui/src/ConversationSurface.test.tsx`
- Modify: `packages/agent-ui/src/index.ts`

- [x] Write rendering tests for loading, empty, error, sent user message,
  active assistant response, disabled composer, send failure, and interrupt.
- [x] Implement `ConversationSurface` only through the runtime context. It
  renders semantic roles and status, invokes `runtime.send`, and calls
  `runtime.interrupt` while a response is active.
- [x] Implement `mountConversation(element, runtime, options)` for non-React
  pages. It creates and returns a React root with an idempotent `unmount`.
- [x] Keep visual styling application-owned and avoid importing application CSS,
  Router, QueryClient, or auth stores. Consumers own visual tokens and brand
  treatment while the shared surface exposes semantic structure.
- [x] Run the package Vitest suite and TypeScript check.

## Task 3: Strict Iframe Bridge

**Files:**
- Create: `packages/agent-ui/src/iframe-bridge.ts`
- Create: `packages/agent-ui/src/iframe-bridge.test.ts`
- Modify: `packages/agent-ui/src/index.ts`

- [x] Define versioned messages: `ready`, `open_session`, `session_changed`,
  `resize`, and `error`. Every message includes a caller-supplied nonce.
- [x] Write tests proving the bridge rejects an unexpected origin, wrong nonce,
  malformed payload, and unknown protocol version.
- [x] Implement `createConversationFrameBridge` for iframe pages and
  `createConversationIframeHost` for parent pages. Both require an explicit
  target origin; neither sends or accepts wildcard-origin messages.
- [x] Ensure the bridge has no credential, token refresh, or transport logic.
  Authentication remains inside the selected runtime or the iframe's normal
  product login flow.
- [x] Run bridge tests and TypeScript check.

## Task 4: Web User Adapter and Iframe Entry

**Files:**
- Create: `clients/web-user/src/lib/chatStoreConversationRuntime.ts`
- Create: `clients/web-user/src/lib/chatStoreConversationRuntime.test.ts`
- Create: `clients/web-user/src/iframe.tsx`
- Create: `clients/web-user/iframe.html`
- Modify: `clients/web-user/vite.config.ts`
- Modify: `clients/web-user/tsconfig.app.json`

- [x] Write adapter tests that map current session status, optimistic user text,
  persisted user text, active assistant text, load errors, `switchTo`, `send`,
  and `stop` into the public runtime contract.
- [x] Implement a single explicit adapter over `useChatStore`. It is a
  migration adapter, not a second store: all mutation continues to delegate to
  the existing actions.
- [x] Add an iframe entry that mounts `ConversationSurface`, creates the
  adapter, and binds the strict bridge. `open_session` selects the requested
  existing session; it does not create a hidden fallback session.
- [x] Configure Vite to emit `iframe.html` for a deployable iframe URL while
  preserving the existing standalone and same-root embed builds.
- [x] Run focused Web User tests, type-check, and a production Vite build.

The full `clients/web-user` project type-check still reports 71 pre-existing
errors outside this delivery. The final filtered check reports zero errors in
`packages/agent-ui`, `chatStoreConversationRuntime`, and `iframe.tsx`.

## Task 5: Consumer Documentation and Browser Verification

**Files:**
- Create: `packages/agent-ui/README.md`
- Modify: `docs/superpowers/specs/2026-07-12-agent-conversation-component-design.md`

- [x] Document React usage, `mountConversation` usage, and strict iframe host
  usage with an explicit target origin and session open request.
- [x] Record that the Web User adapter is temporary and that Rust/Connect
  runtime migration is the next delivery, not an automatic transport fallback.
- [x] Use the running `clients/web-user` Vite server and browser tests to
  verify `iframe.html` mounts, sends `ready`, accepts a valid `open_session`,
  rejects a wrong-origin message, and renders loading/empty state.
- [x] Run `git diff --check` for delivery paths; report test and browser evidence without
  committing because this shared worktree contains unrelated user changes.

## Review Checklist

- [x] The public package has no imports from `clients/web-user` or `clients/web`.
- [x] No message path accepts `*` as an origin or serializes credentials.
- [x] Runtime state is subscribed through `useSyncExternalStore`; no mirror
  Zustand state is added.
- [x] Existing Web User routes and same-root `embed.tsx` behavior remain intact.
- [x] Every new source file stays under 200 lines.
- [x] The implementation preserves the design's future Rust/Connect boundary.

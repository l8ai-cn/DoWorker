# Agent Workbench Web User And Embed Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Web User, React mounts, plain page mounts, and iframe embeds consume the same lossless V2 session runtime and results workbench.

**Architecture:** Web User owns transport/auth adaptation only. Shared runtime owns snapshot/delta consistency. React and imperative mounts may register executable renderers; iframe messages select prebundled renderer IDs and capability-scoped actions.

**Tech Stack:** React 18, Vite, TypeScript, REST/SSE, postMessage, Vitest, Playwright.

---

### Task 1: Add V2 REST And SSE Client Contracts

**Files:**
- Modify: `clients/web-user/src/embed-session-api.ts`
- Modify: `clients/web-user/src/embed-session-api.test.ts`
- Modify: `clients/web-user/src/embed-session-response-parsers.ts`
- Create: `clients/web-user/src/embed-session/v2SessionStream.ts`
- Create: `clients/web-user/src/embed-session/v2SessionStream.test.ts`
- Create: `clients/web-user/src/embed-session/v2SessionHydration.ts`
- Create: `clients/web-user/src/embed-session/v2SessionHydration.test.ts`

- [ ] **Step 1: Write cursor and overflow tests**

```ts
expect(request.headers.get("Last-Event-ID")).toBe("epoch-a:41");
expect(parseUint64("18446744073709551615")).toBe(18446744073709551615n);
expect(buffer.push(eventOver4MiB)).toBe("resync_required");
```

- [ ] **Step 2: Verify current client fails**

Run: `pnpm --dir clients/web-user test -- embed-session-api v2Session`
Expected: FAIL because the V2 snapshot, sequence cursor, and bounded buffer are absent.

- [ ] **Step 3: Implement subscribe-buffer-snapshot-replay**

Open SSE first, buffer at most 5,000 events or 4 MiB, fetch one atomic snapshot, commit it, then replay events after `latestSequence`. Gaps, epoch changes, invalid decimal uint64, and overflow call `requestResync`; no partial state is published.

- [ ] **Step 4: Run and commit**

Run: `pnpm --dir clients/web-user test -- embed-session-api v2Session`
Expected: PASS for reconnect, overflow, epoch change, and exact cursor.

```bash
git add clients/web-user/src/embed-session-api.ts clients/web-user/src/embed-session-api.test.ts clients/web-user/src/embed-session-response-parsers.ts clients/web-user/src/embed-session/v2SessionStream.ts clients/web-user/src/embed-session/v2SessionStream.test.ts clients/web-user/src/embed-session/v2SessionHydration.ts clients/web-user/src/embed-session/v2SessionHydration.test.ts
git commit -m "feat(web-user): consume workbench v2 snapshots and deltas"
```

### Task 2: Replace Lossy Embedded Projections

**Files:**
- Modify: `clients/web-user/src/embed-session/EmbeddedAgentSessionRuntime.ts`
- Modify: `clients/web-user/src/embed-session/EmbeddedAgentSessionRuntime.test.ts`
- Modify: `clients/web-user/src/embed-session/EmbeddedSessionCommands.ts`
- Modify: `clients/web-user/src/embed-session/EmbeddedAgentWorkspace.tsx`
- Delete: `clients/web-user/src/embed-session/embeddedTimelineProjection.ts`
- Delete: `clients/web-user/src/embed-session/embeddedWorkspaceProjection.ts`
- Delete: `clients/web-user/src/embed-session/embeddedWorkspaceProjection.test.ts`
- Delete: `clients/web-user/src/embed-session/embeddedRuntimeState.ts`

- [ ] **Step 1: Add lossless runtime tests**

```ts
expect(runtime.getSnapshot().history.find((item) => item.payload.case === "unsupported")).toBeDefined();
expect(runtime.getSnapshot().artifacts[0].representations).toHaveLength(3);
expect(receipt.state).toBe("accepted");
```

- [ ] **Step 2: Implement shared store delegation**

`EmbeddedAgentSessionRuntime` owns client, stream, artifact, and terminal adapters, then delegates `applySnapshot`, `applyDeltaBatch`, `execute`, receipt lookup, and subscriptions to `AgentSessionStore`. Commands send generated envelopes and never directly remove permissions or optimistic messages.

- [ ] **Step 3: Remove old state path**

Delete the projection and handwritten runtime state files in the same commit. No V1 adapter or transport fallback remains.

- [ ] **Step 4: Run and commit**

Run: `pnpm --dir clients/web-user test -- EmbeddedAgentSessionRuntime EmbeddedAgentWorkspace`
Expected: PASS with unsupported blocks, receipts, artifacts, permissions, and resources preserved.

```bash
git add clients/web-user/src/embed-session
git commit -m "refactor(web-user): replace embedded v1 projections"
```

### Task 3: Integrate Result Workbench And Host Adapters

**Files:**
- Modify: `clients/web-user/src/embed-session/EmbeddedAgentWorkspace.tsx`
- Modify: `clients/web-user/src/embed-session/EmbeddedAgentWorkspace.test.tsx`
- Create: `clients/web-user/src/embed-session/WebUserArtifactRuntime.ts`
- Create: `clients/web-user/src/embed-session/WebUserArtifactRuntime.test.ts`
- Modify: `clients/web-user/src/embed-session/EmbeddedTerminalRuntime.ts`
- Modify: `clients/web-user/src/embed-session/EmbeddedTerminalRuntime.test.ts`
- Modify: `clients/web-user/src/embed-workspace-artifact-api.ts`

- [ ] **Step 1: Add artifact and terminal action tests**

```ts
expect(await artifacts.loadRepresentation("deck", "pdf", 4n)).toBeInstanceOf(Blob);
await expect(artifacts.executeAction(staleEdit)).rejects.toThrow("artifact_revision_conflict");
expect(terminal.write).toHaveBeenCalledWith(expect.objectContaining({ fencingEpoch: 9n }));
```

- [ ] **Step 2: Implement host adapters**

Map Pod-scoped artifact list/read/download/action endpoints into `ArtifactRuntime`. Map terminal connect/output/input/resize/lease into V2 resources and include fencing epoch on every mutation. Render `AgentWorkbench` with shared tool/content registries and Chinese locale by default when browser locale starts with `zh`.

- [ ] **Step 3: Run and commit**

Run: `pnpm --dir clients/web-user test -- WebUserArtifactRuntime EmbeddedTerminalRuntime EmbeddedAgentWorkspace`
Expected: PASS for artifacts, stale edit, observer mode, lease loss, and results selection.

```bash
git add clients/web-user/src/embed-session clients/web-user/src/embed-workspace-artifact-api.ts
git commit -m "feat(web-user): integrate shared results and terminal runtimes"
```

### Task 4: Publish React And Plain Mount Entries

**Files:**
- Modify: `clients/web-user/src/mountEmbeddedAgentWorkspace.tsx`
- Modify: `clients/web-user/src/mountEmbeddedAgentWorkspace.test.tsx`
- Modify: `clients/web-user/src/mount.tsx`
- Modify: `clients/web-user/src/mount.test.tsx`
- Modify: `clients/web-user/src/standalone.tsx`
- Modify: `clients/web-user/src/embed.tsx`
- Modify: `clients/web-user/vite.embed.config.ts`
- Modify: `clients/web-user/vite-embed-css-scope.ts`
- Modify: `clients/web-user/package.json`

- [ ] **Step 1: Test mount isolation**

```ts
const mounted = mountAgentWorkbench(host, { runtime, renderers: [trustedRenderer] });
expect(host.querySelector("[data-agent-workbench]")).not.toBeNull();
mounted.unmount();
expect(host.childElementCount).toBe(0);
```

- [ ] **Step 2: Implement entries**

React exports component props. Plain mount creates and owns a React root, scoped CSS, locale, theme tokens, runtime, and executable renderer registrations. Build emits ESM and CJS mount entries and does not bundle Monaco, PDF, spreadsheet, presentation, or video editor chunks into first load.

- [ ] **Step 3: Verify and commit**

Run: `pnpm --dir clients/web-user test -- mount mountEmbeddedAgentWorkspace && pnpm --dir clients/web-user build:embed`
Expected: PASS and gzip budgets from the renderer design remain green.

```bash
git add clients/web-user/src/mountEmbeddedAgentWorkspace.tsx clients/web-user/src/mountEmbeddedAgentWorkspace.test.tsx clients/web-user/src/mount.tsx clients/web-user/src/mount.test.tsx clients/web-user/src/standalone.tsx clients/web-user/src/embed.tsx clients/web-user/vite.embed.config.ts clients/web-user/vite-embed-css-scope.ts clients/web-user/package.json
git commit -m "feat(web-user): publish workbench component and mount entries"
```

### Task 5: Harden The Iframe Entry

**Files:**
- Modify: `clients/web-user/src/iframe.tsx`
- Modify: `clients/web-user/src/embed-app.tsx`
- Modify: `clients/web-user/src/embed-context.ts`
- Modify: `clients/web-user/src/embed-context.test.ts`
- Modify: `clients/web-user/src/embed-session/EmbeddedSessionIframe.tsx`
- Modify: `clients/web-user/src/embed-session/EmbeddedSessionIframe.test.tsx`
- Modify: `clients/web-user/src/embed-session/embedParentHandshake.ts`
- Modify: `clients/web-user/src/embed-session/embedParentHandshake.test.ts`
- Modify: `clients/web-user/e2e/embed-host.html`

- [ ] **Step 1: Add spoof and executable-config tests**

```ts
expect(handleMessage({ source: sibling, origin: allowedOrigin, data: validMessage })).toBe(false);
expect(parseEmbedOptions({ rendererModuleUrl: "https://evil.test/x.js" })).toThrow("invalid_embed_option");
expect(postMessage).toHaveBeenCalledWith(expect.anything(), allowedOrigin);
```

- [ ] **Step 2: Implement data-only configuration**

Bind `event.source` to the parent window, use exact target origin, nonce, message ID, protocol version, size limit, and replay set. Options allow locale, theme, initial surface, grants, and prebundled renderer IDs only.

- [ ] **Step 3: Run and commit**

Run: `pnpm --dir clients/web-user test -- EmbeddedSessionIframe embedParentHandshake embed-context`
Expected: PASS for wrong source, wrong origin, replay, oversized payload, and renderer-module rejection.

```bash
git add clients/web-user/src/iframe.tsx clients/web-user/src/embed-app.tsx clients/web-user/src/embed-context.ts clients/web-user/src/embed-context.test.ts clients/web-user/src/embed-session/EmbeddedSessionIframe.tsx clients/web-user/src/embed-session/EmbeddedSessionIframe.test.tsx clients/web-user/src/embed-session/embedParentHandshake.ts clients/web-user/src/embed-session/embedParentHandshake.test.ts clients/web-user/e2e/embed-host.html
git commit -m "feat(web-user): publish the capability-scoped iframe entry"
```

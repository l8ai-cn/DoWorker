# Agent Workbench Phase 0A Content Security Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove static HTML and Markdown active-content escape paths before shared Agent renderers ship.

**Architecture:** Static HTML uses a trusted shell around an opaque sandbox. Markdown blocks remote fetches by default and exposes only explicit user-initiated navigation.

**Tech Stack:** React, TypeScript, Vitest.

---

### Task 1: Shared Static HTML And Remote Resource Policy

**Files:**
- Create: `packages/agent-ui/src/security/staticHtmlProfile.ts`
- Create: `packages/agent-ui/src/security/staticHtmlProfile.test.ts`
- Create: `packages/agent-ui/src/security/markdownResourcePolicy.ts`
- Create: `packages/agent-ui/src/security/markdownResourcePolicy.test.ts`
- Modify: `packages/agent-ui/src/index.ts`

- [ ] **Step 1: Write failing policy tests**

```ts
expect(STATIC_HTML_SANDBOX).toBe("");
expect(markdownImageSource("https://tracker.test/p.png")).toBeUndefined();
expect(markdownImageSource("blob:https://app.test/id")).toBe("blob:https://app.test/id");
expect(markdownImageSource("data:image/png;base64,AA==")).toBe("data:image/png;base64,AA==");
```

- [ ] **Step 2: Verify failure**

Run: `pnpm --dir packages/agent-ui exec vitest run src/security`
Expected: FAIL because the policy modules do not exist.

- [ ] **Step 3: Implement exact policy**

```ts
export const STATIC_HTML_SANDBOX = "";
export const STATIC_HTML_REFERRER_POLICY = "no-referrer";

export function markdownImageSource(src?: string): string | undefined {
  if (!src) return undefined;
  return /^(blob:|data:image\/)/i.test(src) ? src : undefined;
}
```

`openStaticHtmlInNewWindow` must open a trusted blank shell, set `opener = null`, create an iframe with the empty sandbox, set `referrerpolicy=no-referrer`, assign `srcdoc`, and never open the untrusted document as a top-level Blob.

- [ ] **Step 4: Run tests and commit**

Run: `pnpm --dir packages/agent-ui exec vitest run src/security`
Expected: PASS.

```bash
git add packages/agent-ui/src/security packages/agent-ui/src/index.ts
git commit -m "fix(agent-ui): isolate static html and markdown resources"
```

### Task 2: Apply Static Policy In Web And Web User

**Files:**
- Modify: `clients/web/src/components/media/HtmlPreviewCard.tsx`
- Modify: `clients/web/src/components/media/__tests__/HtmlPreviewCard.test.tsx`
- Modify: `packages/agent-ui/src/MarkdownMessage.tsx`
- Create: `packages/agent-ui/src/MarkdownMessage.test.tsx`
- Modify: `clients/web-user/src/shell/codeViewerHelpers.ts`
- Modify: `clients/web-user/src/shell/codeViewerHelpers.test.ts`
- Modify: `clients/web-user/src/shell/CodeViewer.tsx`
- Modify: `clients/web-user/src/shell/CodeViewer.test.tsx`

- [ ] **Step 1: Add regression assertions**

```ts
expect(window.open).not.toHaveBeenCalledWith(expect.stringMatching(/^blob:/), "_blank", expect.anything());
expect(frame).toHaveAttribute("sandbox", "");
expect(screen.queryByRole("img", { name: "tracker" })).not.toBeInTheDocument();
```

- [ ] **Step 2: Verify the current implementation fails**

Run: `pnpm run web:test -- HtmlPreviewCard && pnpm --dir clients/web-user test -- codeViewerHelpers CodeViewer`
Expected: FAIL on top-level Blob, permissive sandbox, or remote image loading.

- [ ] **Step 3: Use the shared policy**

Replace both local HTML sandbox/open helpers with the shared exports. `HtmlPreviewCard` no longer accepts an arbitrary remote `src`; live URLs render only through the typed `pod-live` preview descriptor in Phase 0B. Both Markdown renderers map blocked remote images to visible text with an explicit user-initiated open action.

- [ ] **Step 4: Run focused suites and commit**

Run: `pnpm run web:test -- HtmlPreviewCard && pnpm --dir clients/web-user test -- codeViewerHelpers CodeViewer`
Expected: PASS with no top-level untrusted document and no automatic remote fetch.

```bash
git add clients/web/src/components/media packages/agent-ui/src/MarkdownMessage.tsx packages/agent-ui/src/MarkdownMessage.test.tsx clients/web-user/src/shell/codeViewerHelpers.ts clients/web-user/src/shell/codeViewerHelpers.test.ts clients/web-user/src/shell/CodeViewer.tsx clients/web-user/src/shell/CodeViewer.test.tsx
git commit -m "fix(web): enforce shared artifact sandbox policy"
```

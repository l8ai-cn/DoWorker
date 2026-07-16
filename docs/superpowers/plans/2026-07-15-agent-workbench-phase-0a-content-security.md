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

- [x] **Step 1: Write failing policy tests**

```ts
expect(STATIC_HTML_SANDBOX).toBe("");
expect(markdownImageSource("https://tracker.test/p.png")).toBeUndefined();
expect(markdownImageSource("blob:https://app.test/id")).toBe("blob:https://app.test/id");
expect(markdownImageSource("data:image/png;base64,AA==")).toBe("data:image/png;base64,AA==");
```

- [x] **Step 2: Verify failure**

Run: `pnpm --dir packages/agent-ui exec vitest run src/security`
Expected: FAIL because the policy modules do not exist.

- [x] **Step 3: Implement exact policy**

```ts
export const STATIC_HTML_SANDBOX = "";
export const STATIC_HTML_REFERRER_POLICY = "no-referrer";

export function markdownImageSource(src?: string): string | undefined {
  if (!src) return undefined;
  return /^(blob:|data:image\/)/i.test(src) ? src : undefined;
}
```

`openStaticHtmlInNewWindow` must open a trusted blank shell, set `opener = null`, create an iframe with the empty sandbox, set `referrerpolicy=no-referrer`, assign `srcdoc`, and never open the untrusted document as a top-level Blob.
`staticHtmlDocument` must remove artifact-provided base, referrer, and CSP metadata before injecting the platform-owned deny-by-default CSP.

- [x] **Step 4: Run tests and commit**

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
- Modify: `packages/agent-ui/src/ArtifactCard.tsx`
- Modify: `packages/agent-ui/src/ArtifactCard.test.tsx`
- Modify: `packages/agent-ui/src/MarkdownMessage.tsx`
- Create: `packages/agent-ui/src/MarkdownMessage.test.tsx`
- Modify: `clients/web-user/src/shell/codeViewerHelpers.ts`
- Delete: `clients/web-user/src/shell/codeViewerHelpers.test.ts`
- Modify: `clients/web-user/src/shell/HtmlCommentViewer.tsx`
- Modify: `clients/web-user/src/shell/HtmlCommentViewer.test.tsx`
- Modify: `clients/web-user/src/shell/CodeViewer.tsx`
- Modify: `clients/web-user/src/shell/CodeViewer.test.tsx`
- Create: `clients/web-user/src/shell/CodeViewer.preview.test.tsx`
- Create: `clients/web-user/src/shell/MarkdownPreview.tsx`
- Create: `clients/web-user/src/shell/StaticHtmlPreview.tsx`
- Create: `clients/web-user/src/shell/FileImageViewer.tsx`
- Create: `clients/web-user/src/shell/SourceCodeViewer.tsx`
- Create: `clients/web-user/src/shell/SourceCodeLine.tsx`
- Create: `clients/web-user/src/shell/SourceCodeSearchBar.tsx`
- Create: `clients/web-user/src/shell/SourceSelectionActions.tsx`
- Create: `clients/web-user/src/shell/useCodeViewerSourceState.ts`
- Create: `clients/web-user/src/shell/useSourceCodeKeyboard.ts`
- Create: `clients/web-user/src/shell/useSourceCodeSearch.ts`
- Create: `clients/web-user/src/shell/useSourceSelectionActions.ts`
- Create: `clients/web-user/src/shell/fileContentClassification.ts`
- Create: `clients/web-user/src/shell/sourceSelectionOffsets.ts`
- Create: `clients/web-user/src/shell/staticHtmlArtifactPopout.ts`
- Create: `clients/web-user/src/shell/staticHtmlArtifactPopout.test.tsx`
- Modify: `clients/web-user/src/shell/TipTapWorkspaceImage.ts`
- Modify: `clients/web-user/src/shell/TipTapWorkspaceImage.test.ts`
- Create: `clients/web-user/src/shell/WorkspaceImageNodeView.ts`
- Create: `clients/web-user/src/shell/workspaceImagePaths.ts`
- Delete: `clients/web-user/src/shell/htmlCommentBridge.ts`
- Delete: `clients/web-user/src/shell/htmlCommentBridge.test.ts`

- [x] **Step 1: Add regression assertions**

```ts
expect(window.open).not.toHaveBeenCalledWith(expect.stringMatching(/^blob:/), "_blank", expect.anything());
expect(frame).toHaveAttribute("sandbox", "");
expect(screen.queryByRole("img", { name: "tracker" })).not.toBeInTheDocument();
```

- [x] **Step 2: Verify the current implementation fails**

Run: `pnpm run web:test -- HtmlPreviewCard && pnpm --dir clients/web-user test -- codeViewerHelpers CodeViewer`
Expected: FAIL on top-level Blob, permissive sandbox, or remote image loading.

- [x] **Step 3: Use the shared policy**

Replace every static artifact iframe and local HTML sandbox/open helper with the shared exports. `HtmlPreviewCard` no longer accepts an arbitrary remote `src`; live URLs render only through the typed `pod-live` preview descriptor in Phase 0B. Read-only Markdown and the TipTap editor block initial remote image requests and expose an explicit user-initiated load/open action. Static HTML review does not inject an executable comment bridge; interactive review moves to the dedicated preview profile.

- [x] **Step 4: Run focused suites and commit**

Run: focused Web, shared package, and Web User suites plus target lint/typecheck.
Expected: PASS with no top-level untrusted document, no automatic remote fetch, and no target-file type errors.

```bash
git add clients/web/src/components/media packages/agent-ui/src/MarkdownMessage.tsx packages/agent-ui/src/MarkdownMessage.test.tsx clients/web-user/src/shell/codeViewerHelpers.ts clients/web-user/src/shell/codeViewerHelpers.test.ts clients/web-user/src/shell/CodeViewer.tsx clients/web-user/src/shell/CodeViewer.test.tsx
git commit -m "fix(web): enforce shared artifact sandbox policy"
```

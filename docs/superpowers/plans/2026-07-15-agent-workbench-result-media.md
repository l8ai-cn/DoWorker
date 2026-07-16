# Agent Workbench Result And Media Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a reusable results workbench with extracted file viewers, durable media state, image editing actions, video playback, and presentation navigation.

**Architecture:** Artifact data is loaded through a headless runtime. Tool renderers select results; content renderers own viewing/editing. Heavy viewers are lazy chunks, and stateful surfaces remain alive across layout changes.

**Tech Stack:** React, TypeScript, react-resizable-panels, ResizeObserver, Vitest, Testing Library, Playwright.

---

### Task 1: Add Artifact Runtime And Viewer Contracts

**Files:**
- Create: `packages/agent-ui/src/artifacts/ArtifactRuntime.ts`
- Create: `packages/agent-ui/src/artifacts/ArtifactController.ts`
- Create: `packages/agent-ui/src/artifacts/ArtifactController.test.ts`
- Create: `packages/agent-ui/src/artifacts/artifactAction.ts`
- Create: `packages/agent-ui/src/artifacts/artifactAction.test.ts`
- Create: `packages/agent-ui/src/viewers/UnsupportedViewer.tsx`
- Create: `packages/agent-ui/src/viewers/UnsupportedViewer.test.tsx`

- [ ] **Step 1: Write stale-revision and idempotency tests**

```ts
await expect(controller.execute({
  artifactId: "image-1",
  representationId: "source",
  baseRevision: 3n,
  clientActionId: "action-1",
  actionType: "edit_image",
  payload: {},
})).rejects.toThrow("artifact_revision_conflict");
```

- [ ] **Step 2: Implement the headless controller**

Cache payloads by `(artifactId, representationId, revision)`. Mutation never changes the descriptor locally; it waits for the runtime receipt or artifact delta. Unsupported viewer shows exact media type, role, schema version, raw metadata, and original download.

- [ ] **Step 3: Verify and commit**

Run: `pnpm --dir packages/agent-ui exec vitest run src/artifacts src/viewers/UnsupportedViewer.test.tsx`
Expected: PASS.

```bash
git add packages/agent-ui/src/artifacts packages/agent-ui/src/viewers/UnsupportedViewer.tsx packages/agent-ui/src/viewers/UnsupportedViewer.test.tsx
git commit -m "feat(agent-ui): add artifact runtime and unsupported viewer"
```

### Task 2: Extract Web User File Viewers

**Files:**
- Create: `packages/agent-ui/src/viewers/code/CodeArtifactViewer.tsx`
- Create: `packages/agent-ui/src/viewers/markdown/MarkdownArtifactViewer.tsx`
- Create: `packages/agent-ui/src/viewers/diff/DiffArtifactViewer.tsx`
- Create: `packages/agent-ui/src/viewers/image/ImageArtifactViewer.tsx`
- Create: `packages/agent-ui/src/viewers/html/HtmlArtifactViewer.tsx`
- Create: `packages/agent-ui/src/viewers/viewerTypes.ts`
- Create: `packages/agent-ui/src/viewers/builtinContentRenderers.ts`
- Create: `packages/agent-ui/src/viewers/builtinContentRenderers.test.ts`
- Modify: `clients/web-user/src/shell/FileViewer.tsx`
- Modify: `clients/web-user/src/shell/FileViewer.test.tsx`
- Modify: `clients/web-user/src/shell/CodeViewer.tsx`
- Modify: `clients/web-user/src/shell/CodeViewer.test.tsx`
- Modify: `clients/web-user/src/components/blocks/OutputFileArtifact.tsx`
- Modify: `clients/web-user/src/components/blocks/OutputFileArtifact.test.tsx`
- Modify: `clients/web-user/src/components/blocks/BlockRenderer.tsx`
- Modify: `clients/web-user/src/components/blocks/BlockRenderer.file.test.tsx`

- [ ] **Step 1: Add shared viewer fixture tests**

```ts
expect(resolveContentRenderer({ blockKind: "artifact", mediaType: "text/markdown", role: "preview", schemaVersion: "1" })?.id)
  .toBe("builtin.markdown");
expect(resolveContentRenderer({ blockKind: "artifact", mediaType: "application/x-unknown", role: "original", schemaVersion: "1" }))
  .toBeUndefined();
```

- [ ] **Step 2: Extract pure props**

Each shared viewer accepts only `ArtifactPayload`, descriptor, view state, grants, and callbacks. Web User hooks, stores, router, comments panel, and file-save APIs stay in a host adapter. Monaco, Shiki, TipTap, PDF, and spreadsheet code remain lazy.

- [ ] **Step 3: Switch Web User immediately**

`FileViewer.tsx` becomes a shell selecting a shared viewer. `OutputFileArtifact` and the file branch in `BlockRenderer` select the same content registry. Delete replaced local branches in the same change.

- [ ] **Step 4: Run and commit**

Run: `pnpm --dir packages/agent-ui exec vitest run src/viewers && pnpm --dir clients/web-user test -- FileViewer CodeViewer`
Expected: PASS with Web User using shared code, Markdown, diff, image, and HTML viewers.

```bash
git add packages/agent-ui/src/viewers clients/web-user/src/shell/FileViewer.tsx clients/web-user/src/shell/FileViewer.test.tsx clients/web-user/src/shell/CodeViewer.tsx clients/web-user/src/shell/CodeViewer.test.tsx clients/web-user/src/components/blocks
git commit -m "refactor(web-user): extract shared artifact viewers"
```

### Task 3: Build Result Workbench And Persistent Surfaces

**Files:**
- Create: `packages/agent-ui/src/react/ResultWorkbench.tsx`
- Create: `packages/agent-ui/src/react/ResultWorkbench.test.tsx`
- Create: `packages/agent-ui/src/react/ResourceRail.tsx`
- Create: `packages/agent-ui/src/react/ArtifactViewer.tsx`
- Create: `packages/agent-ui/src/react/useWorkbenchContainerMode.ts`
- Create: `packages/agent-ui/src/react/useWorkbenchContainerMode.test.ts`
- Create: `packages/agent-ui/src/runtime/SurfaceLifetimeController.ts`
- Create: `packages/agent-ui/src/runtime/SurfaceLifetimeController.test.ts`
- Modify: `packages/agent-ui/src/AgentWorkspace.tsx`

- [ ] **Step 1: Write container and lifetime tests**

```ts
expect(containerMode(959)).toBe("medium");
expect(containerMode(639)).toBe("narrow");
expect(terminal.disconnect).not.toHaveBeenCalledAfterSwitchingTabs();
expect(video.currentTime).toBe(42);
```

- [ ] **Step 2: Implement composition**

Wide containers use resizable conversation/results panes and a resource rail. Medium uses a stable split or result drawer. Narrow uses Conversation and Results tabs. Pinned terminal, preview, video, image draft, and remote iframe controllers survive tab changes.

- [ ] **Step 3: Verify and commit**

Run: `pnpm --dir packages/agent-ui exec vitest run src/react src/runtime/SurfaceLifetimeController.test.ts`
Expected: PASS for wide, medium, narrow, selection, pinning, and keepalive.

```bash
git add packages/agent-ui/src/react packages/agent-ui/src/runtime/SurfaceLifetimeController.ts packages/agent-ui/src/runtime/SurfaceLifetimeController.test.ts packages/agent-ui/src/AgentWorkspace.tsx
git commit -m "feat(agent-ui): add persistent result workbench"
```

### Task 4: Add Image Comparison And Editing Actions

**Files:**
- Create: `packages/agent-ui/src/viewers/image/ImageComparisonViewer.tsx`
- Create: `packages/agent-ui/src/viewers/image/ImageComparisonViewer.test.tsx`
- Create: `packages/agent-ui/src/viewers/image/imageGeometry.ts`
- Create: `packages/agent-ui/src/viewers/image/imageGeometry.test.ts`
- Create: `packages/agent-ui/src/viewers/image/ImageAnnotationLayer.tsx`
- Create: `packages/agent-ui/src/viewers/image/ImageEditComposer.tsx`

- [ ] **Step 1: Test normalized coordinates**

```ts
expect(normalizePoint({ x: 250, y: 125 }, { width: 1000, height: 500 })).toEqual({ x: 0.25, y: 0.25 });
expect(denormalizePoint({ x: 0.25, y: 0.25 }, { width: 2000, height: 1000 })).toEqual({ x: 500, y: 250 });
```

- [ ] **Step 2: Implement interactions**

Add fit, zoom, pan, source/result, side-by-side, overlay, slider, region, vector annotation, mask attachment, candidate selection, download, and `edit_image` action. Action payload includes source dimensions, orientation, normalized geometry, mask artifact, instruction, and base revision.

- [ ] **Step 3: Verify and commit**

Run: `pnpm --dir packages/agent-ui exec vitest run src/viewers/image`
Expected: PASS for pointer, keyboard, stale revision, and action payload tests.

```bash
git add packages/agent-ui/src/viewers/image
git commit -m "feat(agent-ui): add image comparison and edit actions"
```

### Task 5: Add Video And Presentation Viewers

**Files:**
- Create: `packages/agent-ui/src/viewers/video/VideoArtifactViewer.tsx`
- Create: `packages/agent-ui/src/viewers/video/VideoArtifactViewer.test.tsx`
- Create: `packages/agent-ui/src/viewers/presentation/PresentationArtifactViewer.tsx`
- Create: `packages/agent-ui/src/viewers/presentation/PresentationArtifactViewer.test.tsx`
- Create: `packages/agent-ui/src/viewers/presentation/presentationActions.ts`
- Modify: `packages/agent-ui/src/viewers/builtinContentRenderers.ts`

- [ ] **Step 1: Add viewer behavior tests**

```ts
expect(screen.getByRole("slider", { name: "播放进度" })).toBeEnabled();
expect(screen.getByRole("button", { name: "重新生成当前页" })).toBeDisabled();
expect(action.baseRevision).toBe(7n);
expect(action.slideId).toBe("slide-2");
```

- [ ] **Step 2: Implement media behavior**

Video shows durable generation stages, poster, metadata, seek, volume, speed, fullscreen, versions, timestamp comments, derivatives, and download. Presentation shows stable slide order, thumbnails, selected slide, notes, fit/zoom, fullscreen, revision compare, comments, and grant-gated regenerate/replace/reorder/export actions.

- [ ] **Step 3: Verify chunks and commit**

Run: `pnpm --dir packages/agent-ui exec vitest run src/viewers/video src/viewers/presentation`
Expected: PASS and production build keeps video/presentation viewers outside base chunks.

```bash
git add packages/agent-ui/src/viewers/video packages/agent-ui/src/viewers/presentation packages/agent-ui/src/viewers/builtinContentRenderers.ts
git commit -m "feat(agent-ui): add video and presentation viewers"
```

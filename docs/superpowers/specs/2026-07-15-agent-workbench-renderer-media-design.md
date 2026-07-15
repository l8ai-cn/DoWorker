# Agent Workbench Renderer And Media Design

**Status:** normative appendix
**Date:** 2026-07-15
**Parent:** `2026-07-15-agent-workbench-v2-design.md`

## 1. Two Registry Model

Tool execution presentation and content viewing are separate extension points:

```ts
type ToolRendererKey = {
  namespace: string;
  semanticKey: string;
  schemaVersion: string;
};

interface ToolRendererRegistration {
  key: ToolRendererKey;
  summary?: ToolSummaryRenderer;
  detail?: ToolDetailRenderer;
  workbench?: ToolWorkbenchRenderer;
}

type ContentRendererKey = {
  blockKind: string;
  mediaType?: string;
  role?: string;
  schemaVersion: string;
};

interface ContentRendererRegistration {
  key: ContentRendererKey;
  inline?: ContentInlineRenderer;
  viewer: ContentViewerRenderer;
  editor?: ContentEditorRenderer;
}
```

The tool registry answers how an execution is summarized and controlled. The content registry answers how a result block or artifact representation is viewed or edited. A browser tool may therefore use one tool renderer while producing HTML, image, video, and log blocks handled by independent content renderers.

Unsupported keys show their exact identity, phase or media type, raw structured data, and available download. They do not route to a guessed specialized renderer.

Registering the same exact key twice fails with `renderer_key_conflict` and reports both sources. Hosts that intentionally replace a renderer must call an explicit `replace` API with the expected current source ID; replacement is never last-write-wins.

## 2. Registration By Host Type

React and imperative same-bundle mounts may register executable renderer implementations. Iframe options are serializable and may select only renderer IDs already bundled in the iframe application. `postMessage` never transports JavaScript, module URLs, callbacks, or credentials.

Remote renderer iframes are content descriptors governed by the security profile appendix. They are not registry registrations.

## 3. Artifact Runtime

```ts
interface ArtifactRuntime {
  loadRepresentation(
    artifactId: string,
    representationId: string,
    revision: bigint,
  ): Promise<ArtifactPayload>;
  download(request: ArtifactDownloadRequest): Promise<void>;
  executeAction(command: ArtifactActionCommand): Promise<CommandReceipt>;
  subscribe(artifactId: string, listener: () => void): () => void;
}
```

Artifact descriptors contain stable `artifactId`, revision, filename, media type, role, status, byte size, dimensions or duration, provenance, representations, and allowed actions. Representations have their own ID, media type, role, status, revision, and transport descriptor.

## 4. Mutation And Concurrency

Every edit uses:

```ts
type ArtifactActionCommand = {
  artifactId: string;
  representationId: string;
  baseRevision: bigint;
  clientActionId: string;
  actionType: string;
  payload: unknown;
};
```

The runtime rejects stale `baseRevision` with the current revision and does not silently merge. Retrying the same `clientActionId` and payload is idempotent. A changed payload requires a new ID.

Image geometry includes source width, source height, EXIF orientation, and normalized coordinates in `[0, 1]`. Regions, paths, points, and masks are interpreted against the declared source representation and revision, not the current CSS box.

## 5. Web User Extraction

`clients/web-user/src/shell/FileViewer.tsx` is split by responsibility:

- headless artifact/file loading controllers;
- pure code, Markdown, diff, image, HTML, PDF, presentation, spreadsheet, video, and unsupported viewers;
- host adapters for Web User routing, stores, and actions;
- lazy entry points for heavy syntax, document, spreadsheet, PDF, and media dependencies.

The shared package owns renderer CSS and exports each viewer through explicit subpaths. Host shells own page layout and product navigation. Extraction switches Web User to the shared renderer in the same change so duplicate implementations do not remain.

Initial shared bundle excludes syntax highlighters, PDF engines, spreadsheet engines, presentation engines, and video editing dependencies. Production gzip budgets are:

- protocol entry: at most 40 KiB;
- runtime incremental entry: at most 50 KiB;
- React incremental entry excluding peer dependencies: at most 100 KiB;
- plain mount first-load JavaScript including React: at most 180 KiB;
- iframe first-load JavaScript including React: at most 220 KiB;
- shared base CSS: at most 35 KiB.

Build checks fail above these limits and prove that each heavy viewer remains a separately loaded chunk.

## 6. Surface Lifetime

Switching conversation/results tabs, resizing panes, or changing the selected result must not terminate live resources.

- Terminal sessions and relay subscriptions live in runtime-owned controllers.
- Video playback keeps media element state while its artifact remains selected or pinned.
- Image annotations and unsent edit instructions live in an external draft store keyed by artifact and revision.
- Live previews and remote renderer iframes remain mounted while pinned; hidden surfaces receive visibility messages.
- Unpinned inactive viewers may unmount after persisting serializable view state.

Tests cover tab switching during terminal output, video playback, annotation, live preview refresh, and remote iframe messaging.

## 7. Responsive Composition

Layout is container-driven:

- `wide` at 960px and above: resizable conversation and result panes with a resource rail;
- `medium` from 640px through 959px: fixed conversation/result split or a collapsible result drawer;
- `narrow` below 640px: full-width Conversation and Results tabs with persistent running-state indicators.

Viewport media queries may set outer chrome, but embedded workbench composition uses a `ResizeObserver`-derived container mode. Controls use stable dimensions and do not reflow when status text changes.

## 8. Image Experience

The built-in viewer supports fit, zoom, pan, source/result versions, side-by-side, overlay, slider comparison, region selection, vector annotation, mask attachment, candidate selection, download, and typed continue-edit actions.

Crop, raster painting, layers, filters, and model-specific controls are separate trusted editor registrations. The base viewer does not imitate capabilities the runtime cannot execute.

## 9. Video Experience

Queued, rendering, and transcoding artifacts show durable stage progress. Playable representations expose poster, duration, dimensions, seek, volume, speed, fullscreen, version selection, timestamp comments, derivatives, and download.

Reconnect restores server progress from the artifact revision. Playback position is local view state and never overwrites server progress.

## 10. Presentation Experience

A presentation artifact contains stable slide IDs and order, deck revision, thumbnail and page representations, notes, and read or edit grants. The viewer provides a thumbnail navigator, selected slide, fit/zoom, fullscreen, notes, revision comparison, and anchored comments.

Typed actions include select revision, comment, regenerate slide, replace slide, reorder slides, and export. Editing controls require explicit grants. PPTX conversion to PDF, HTML, thumbnails, or page images is a Runner/backend capability; the browser does not parse arbitrary office files into an invented preview.

## 11. Verification

Component tests verify exact registry lookup, unsupported rendering, stale-revision rejection, normalized image coordinates, and iframe registration limits. Browser tests cover wide, medium, and narrow containers; keyboard and pointer image comparison; video playback; presentation navigation; dynamic chunk loading; and state preservation across surface changes.

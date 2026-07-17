import { useMemo } from "react";
import type React from "react";
import { defaultRemarkPlugins } from "streamdown";
import remarkBreaks from "remark-breaks";
import { MessageResponse } from "@/components/ai-elements/message";
import { ZoomableImage } from "@/components/ImageLightbox";
import { useThrottledValue } from "@/hooks/useThrottledValue";
import { PathologicalMarkdownText, isPathologicalText } from "./PathologicalMarkdownText";
import { WorkspacePathInlineCode } from "./WorkspacePathInlineCode";

// Markdown images open in the shared lightbox on click, matching uploaded and
// generated images. (Remote `src`s are still gated by Streamdown's image
// security; this only adds the zoom affordance to whatever does render.)
function ZoomableMarkdownImage({ src, alt, ...props }: React.ComponentProps<"img">) {
  const resolvedSrc = typeof src === "string" ? src : undefined;
  return <ZoomableImage {...props} src={resolvedSrc} alt={alt ?? ""} />;
}

// Stable module-level override map so MessageResponse's memo (which ignores
// `components` changes) never sees a new identity.
const FILE_PATH_AWARE_COMPONENTS = {
  inlineCode: WorkspacePathInlineCode,
  img: ZoomableMarkdownImage,
};

// How often the live (growing) assistant bubble re-parses its markdown. The
// store pump commits a new, longer text up to once per animation frame (~60/s);
// without this the whole accumulated message is re-parsed on every commit. ~10/s
// is smooth to read and cuts the per-frame parse cost. Trailing-edge, so the
// final text still appears within this window of the last token.
const STREAM_MARKDOWN_THROTTLE_MS = 100;

/**
 * Wraps `MessageResponse` with {@link WorkspacePathInlineCode} via Streamdown's
 * `inlineCode` slot — NOT `code` — so fenced code blocks keep their default
 * `<pre>` wrapper and Shiki highlighting. Overriding `code` here would replace
 * block rendering too, stripping `<pre>` and collapsing whitespace.
 *
 * When `breaks` is set, single newlines render as `<br>` (remark-breaks)
 * instead of collapsing to spaces per CommonMark. Used for user bubbles,
 * where people type multi-line messages without blank-line paragraph
 * separators and expect their line breaks preserved. NOTE: Streamdown's
 * `remarkPlugins` prop *replaces* its defaults rather than merging, so we
 * extend `defaultRemarkPlugins` (which carries remark-gfm) — passing
 * `[remarkBreaks]` alone would silently drop GFM tables / strikethrough.
 */
export function FilePathAwareMessageResponse({
  children,
  breaks = false,
  ...props
}: React.ComponentProps<typeof MessageResponse> & { breaks?: boolean }) {
  const components = FILE_PATH_AWARE_COMPONENTS;

  // Extend (don't replace) Streamdown's defaults so remark-gfm survives;
  // append remark-breaks only when `breaks` is requested. When `breaks` is
  // false we pass `undefined` so Streamdown uses its own defaults unchanged.
  const remarkPlugins = useMemo(
    () => (breaks ? [...Object.values(defaultRemarkPlugins), remarkBreaks] : undefined),
    [breaks],
  );

  // Throttle the markdown so the live (still-growing) bubble re-parses a few
  // times per second instead of on every store commit. `children` is a string
  // at both call sites (a text RenderItem and the user bubble); finalized/static
  // text changes once, which emits immediately, so this is a no-op off the
  // streaming path. The hook must be called unconditionally (rules of hooks), so
  // non-string children (none today) pass an inert "" and bypass the result.
  const isString = typeof children === "string";
  const throttledText = useThrottledValue(
    isString ? (children as string) : "",
    STREAM_MARKDOWN_THROTTLE_MS,
  );

  // Defense-in-depth: a string child that is huge or carries a
  // giant unbroken token (e.g. a base64 data URL serialized into the text
  // stream) would lock the tab in the markdown pipeline + layout. Render it as
  // plain break-anywhere text instead. Both call sites (assistant text blocks
  // and the user bubble) flow through here, so this one guard covers both.
  const pathological = useMemo(
    () => isString && isPathologicalText(children as string),
    [isString, children],
  );
  if (pathological) {
    return <PathologicalMarkdownText text={children as string} />;
  }

  return (
    <MessageResponse {...props} components={components} remarkPlugins={remarkPlugins}>
      {isString ? throttledText : children}
    </MessageResponse>
  );
}

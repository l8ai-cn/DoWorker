import type React from "react";
import { toWorkspaceRelativePath, useWorkspaceFileExists } from "@/hooks/useWorkspaceChangedFiles";
import { cn } from "@/lib/utils";
import {
  useFileViewer,
  useFileViewerConversationId,
  useIsChangedPath,
  useWorkspacePaths,
} from "@/shell/FileViewerContext";

/**
 * Inline-`code` renderer that turns workspace file paths (e.g.
 * `` `src/components/App.tsx` ``) into clickable links opening the FileViewer.
 *
 * The span's text is first collapsed to a workspace-relative path: an
 * absolute (`/home/u/ws/foo.md`) or home-relative (`~/ws/foo.md`) path under
 * the workspace root is stripped down to its relative form so it matches the
 * changed-files list and the filesystem API (both speak relative paths);
 * absolute/`~` paths outside the root resolve to null and never linkify.
 *
 * That relative path is then linkified when it is either (a) a known
 * agent-changed file — resolved synchronously, the fast path, and the only
 * path that may be an uncommitted/deleted file — or (b) a path-shaped string
 * that the filesystem API confirms points at a real file in the workspace.
 * Everything else (prose-y inline code, non-existent paths) falls back to a
 * styled `<code>` matching Streamdown's default inline appearance. The span
 * always *displays* the original text the agent wrote; only the link target
 * uses the resolved relative path.
 *
 * Rendered by Streamdown as a real component (via the `inlineCode` slot), so
 * it may call hooks: the existence query re-renders this span when it settles,
 * independent of whether `MessageResponse` re-renders its parent.
 */
export function WorkspacePathInlineCode({
  children: codeChildren,
  className,
  ...codeProps
}: React.ComponentPropsWithoutRef<"code">) {
  const openFile = useFileViewer();
  const isChangedPath = useIsChangedPath();
  const conversationId = useFileViewerConversationId();
  const { root, home } = useWorkspacePaths();
  const text = typeof codeChildren === "string" ? codeChildren : "";

  // Collapse absolute / "~"-relative forms onto a workspace-relative path so
  // they match the changed-files list and the filesystem API. null = absolute
  // or "~" path outside the workspace (or the root itself) → never a link.
  const linkPath = text ? toWorkspaceRelativePath(text, root, home) : null;
  // "Trusted" means we resolved an absolute/"~" form against the root, so the
  // result is known workspace-relative even if it's a bare basename (no
  // interior slash) that the existence check's path-shape heuristic rejects.
  const trusted = linkPath !== null && linkPath !== text;

  const isChanged = !!linkPath && isChangedPath(linkPath);
  // Only hit the filesystem for path-shaped spans that aren't already known
  // changes; passing null disables the query (keeps hook order stable).
  const existsOnDisk = useWorkspaceFileExists(
    conversationId,
    openFile && linkPath && !isChanged ? linkPath : null,
    trusted,
  );

  if (openFile && linkPath && (isChanged || existsOnDisk)) {
    // Rendered as an inline <code> (not a <button>): a button is laid out as
    // an atomic inline-block, so a long path can't break across lines and
    // drops below the list marker as a whole unit. An inline <code> flows and
    // wraps like the surrounding text; role/tabIndex/keydown restore the
    // button semantics.
    return (
      <code
        role="button"
        tabIndex={0}
        data-streamdown="inline-code"
        // Keep the base inline-code class/props (merge, don't replace) so the
        // link only adds the underline affordance on top of Streamdown's
        // styling and any caller-provided attributes survive.
        className={cn(
          "font-mono text-sm underline decoration-dotted underline-offset-2 hover:text-foreground transition-colors cursor-pointer",
          className,
        )}
        onClick={() => openFile(linkPath)}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            openFile(linkPath);
          }
        }}
        {...codeProps}
      >
        {codeChildren}
      </code>
    );
  }
  // Match Streamdown's default inline-code styling so non-path inline code
  // looks unchanged.
  return (
    <code
      className={cn("rounded bg-muted px-1.5 py-0.5 font-mono text-sm", className)}
      data-streamdown="inline-code"
      {...codeProps}
    >
      {codeChildren}
    </code>
  );
}

import type { Comment } from "@/hooks/useComments";
import { cn } from "@/lib/utils";
import { renderLineTokens } from "./codeViewerRendering";
import type { SourceLineModel } from "./sourceCodeViewModel";

interface SourceCodeLineProps {
  line: SourceLineModel;
  searchQuery: string;
  setLineRef: (index: number, element: HTMLDivElement | null) => void;
  onSelectComment: (comment: Comment) => void;
}

export function SourceCodeLine({
  line,
  searchQuery,
  setLineRef,
  onSelectComment,
}: SourceCodeLineProps) {
  const hasHighlight =
    line.commentOverlays.length > 0 || line.selectionOverlaps;

  return (
    <div
      ref={(element) => setLineRef(line.index, element)}
      className={cn(line.isCurrentMatch && "bg-yellow-200/40 dark:bg-yellow-700/30")}
    >
      <div className="flex items-stretch">
        <div
          data-gutter-comment={line.comment ? true : undefined}
          className={cn(
            "relative flex w-12 shrink-0 select-none items-center justify-end border-r border-border px-2 py-0.5 text-xs leading-5",
            line.comment
              ? "cursor-pointer text-yellow-500 hover:bg-muted/60 dark:text-yellow-400"
              : "text-muted-foreground/50",
            hasHighlight && "bg-yellow-500/10 dark:bg-yellow-400/15",
          )}
          onClick={() => line.comment && onSelectComment(line.comment)}
        >
          <span>{line.lineNumber}</span>
        </div>
        <div
          data-line={line.lineNumber}
          className="relative flex-1 overflow-hidden whitespace-pre-wrap break-all py-0.5 pl-3 leading-5"
        >
          {line.commentOverlays.map((overlay) => (
            <span
              key={overlay.id}
              aria-hidden
              className={cn(
                "pointer-events-none absolute inset-y-0",
                overlay.isSelected
                  ? "bg-yellow-400/25 dark:bg-yellow-400/25"
                  : "bg-yellow-200/40 dark:bg-yellow-400/20",
              )}
              style={{
                left: `calc(0.75rem + ${overlay.startCol}ch)`,
                width: `${overlay.endCol - overlay.startCol}ch`,
              }}
            />
          ))}
          {line.showDraftSelection &&
            line.selectionEndCol > line.selectionStartCol && (
              <span
                aria-hidden
                className="pointer-events-none absolute inset-y-0 bg-yellow-400/25 dark:bg-yellow-400/25"
                style={{
                  left: `calc(0.75rem + ${line.selectionStartCol}ch)`,
                  width: `${line.selectionEndCol - line.selectionStartCol}ch`,
                }}
              />
            )}
          {line.tokens
            ? renderLineTokens(
                line.tokens,
                line.isMatch ? searchQuery : "",
                line.isCurrentMatch,
              )
            : line.rawLine}
        </div>
      </div>
    </div>
  );
}

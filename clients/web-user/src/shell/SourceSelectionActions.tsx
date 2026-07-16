import { createPortal } from "react-dom";
import { AtSignIcon, MessageSquarePlusIcon } from "lucide-react";
import { getEmbedRoot } from "@/lib/host";
import { useChatStore } from "@/store/chatStore";
import type { ActiveSelection } from "./codeViewerTypes";
import { indexToLine } from "./sourceSelectionOffsets";
import type { SourceSelectionAnchor } from "./useSourceSelectionActions";

interface SourceSelectionActionsProps {
  anchor: SourceSelectionAnchor;
  canAttachToAgent: boolean;
  path: string;
  rawLines: string[];
  onSetActiveSelection: (selection: ActiveSelection | null) => void;
  onClose: () => void;
}

export function SourceSelectionActions({
  anchor,
  canAttachToAgent,
  path,
  rawLines,
  onSetActiveSelection,
  onClose,
}: SourceSelectionActionsProps) {
  const estimatedWidth = canAttachToAgent ? 288 : 138;
  return createPortal(
    <div
      className="fixed z-50 flex items-center gap-1"
      style={{
        left: Math.min(anchor.x, Math.max(8, window.innerWidth - estimatedWidth)),
        top: anchor.y,
        transform: "translateY(-100%)",
      }}
    >
      <button
        data-add-comment-btn
        type="button"
        className="flex items-center gap-1.5 rounded-md border border-border bg-popover px-2.5 py-1 text-xs font-medium text-foreground shadow-md backdrop-blur-xl backdrop-saturate-150 transition-colors hover:bg-secondary"
        onClick={() => {
          onSetActiveSelection({
            start_index: anchor.start_index,
            end_index: anchor.end_index,
            anchor_content: anchor.anchor_content,
          });
          onClose();
          window.getSelection()?.removeAllRanges();
        }}
      >
        <MessageSquarePlusIcon className="size-3.5" />
        Add comment
      </button>
      {canAttachToAgent && (
        <button
          data-attach-agent-btn
          type="button"
          className="flex items-center gap-1.5 rounded-md border border-border bg-popover px-2.5 py-1 text-xs font-medium text-foreground shadow-md backdrop-blur-xl backdrop-saturate-150 transition-colors hover:bg-secondary"
          onClick={() => {
            const startLine = indexToLine(anchor.start_index, rawLines);
            const endLine = indexToLine(
              Math.max(anchor.start_index, anchor.end_index - 1),
              rawLines,
            );
            useChatStore.getState().addComposerAttachment({
              path,
              isDir: false,
              lineRange: { start: startLine, end: endLine },
            });
            onClose();
            window.getSelection()?.removeAllRanges();
          }}
        >
          <AtSignIcon className="size-3.5" />
          Attach to agent
        </button>
      )}
    </div>,
    getEmbedRoot() ?? document.body,
  );
}

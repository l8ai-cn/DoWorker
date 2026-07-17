import { useEffect, useRef, useState, type RefObject } from "react";
import type { Comment } from "@/hooks/useComments";
import type { ActiveSelection } from "./codeViewerTypes";
import { getSelectionOffsets } from "./sourceSelectionOffsets";

export interface SourceSelectionAnchor extends ActiveSelection {
  x: number;
  y: number;
}

interface SourceSelectionOptions {
  resetKey: string;
  rawLines: string[];
  comments: Comment[];
  canEdit: boolean;
  codeContainerRef: RefObject<HTMLDivElement | null>;
  onSetActiveSelection: (selection: ActiveSelection | null) => void;
}

export function useSourceSelectionActions({
  resetKey,
  rawLines,
  comments,
  canEdit,
  codeContainerRef,
  onSetActiveSelection,
}: SourceSelectionOptions) {
  const [selectionAnchor, setSelectionAnchor] = useState<SourceSelectionAnchor | null>(null);
  const commentsRef = useRef(comments);
  const canEditRef = useRef(canEdit);
  const onSetActiveSelectionRef = useRef(onSetActiveSelection);
  useEffect(() => {
    commentsRef.current = comments;
  }, [comments]);
  useEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  useEffect(() => {
    onSetActiveSelectionRef.current = onSetActiveSelection;
  }, [onSetActiveSelection]);

  useEffect(() => {
    const container = codeContainerRef.current;
    if (!container) return;
    const handleMouseUp = (event: MouseEvent) => {
      const selection = window.getSelection();
      if (!selection || selection.rangeCount === 0) return;
      const range = selection.getRangeAt(0);

      if (selection.isCollapsed) {
        if ((event.target as Element).closest("[data-gutter-comment]")) return;
        if (container.contains(range.commonAncestorContainer)) {
          const offsets = getSelectionOffsets(range, container, rawLines);
          const clickedComment = offsets
            ? commentsRef.current.find(
                (comment) =>
                  comment.start_index <= offsets.start_index &&
                  offsets.start_index < comment.end_index,
              )
            : undefined;
          if (clickedComment) {
            onSetActiveSelectionRef.current({
              start_index: clickedComment.start_index,
              end_index: clickedComment.end_index,
              anchor_content: clickedComment.anchor_content ?? "",
            });
            return;
          }
        }
        onSetActiveSelectionRef.current(null);
        return;
      }

      if (!canEditRef.current || !container.contains(range.commonAncestorContainer)) return;
      const anchorContent = selection.toString();
      if (!anchorContent.trim()) return;
      const offsets = getSelectionOffsets(range, container, rawLines);
      if (!offsets) return;
      const firstRect = range.getClientRects()[0] ?? range.getBoundingClientRect();
      const containerLeft = container.getBoundingClientRect().left;
      setSelectionAnchor({
        x: Math.max(firstRect.left, containerLeft + 48),
        y: firstRect.top - 6,
        ...offsets,
        anchor_content: anchorContent,
      });
    };
    container.addEventListener("mouseup", handleMouseUp);
    return () => container.removeEventListener("mouseup", handleMouseUp);
  }, [codeContainerRef, rawLines]);

  useEffect(() => {
    const dismiss = (event: MouseEvent) => {
      if (
        !(event.target as HTMLElement).closest(
          "[data-add-comment-btn], [data-attach-agent-btn]",
        )
      ) {
        setSelectionAnchor(null);
      }
    };
    const dismissOnScroll = () => setSelectionAnchor(null);
    document.addEventListener("mousedown", dismiss);
    window.addEventListener("scroll", dismissOnScroll, true);
    return () => {
      document.removeEventListener("mousedown", dismiss);
      window.removeEventListener("scroll", dismissOnScroll, true);
    };
  }, []);

  useEffect(() => setSelectionAnchor(null), [resetKey]);

  return {
    selectionAnchor,
    clearSelectionAnchor: () => setSelectionAnchor(null),
  };
}

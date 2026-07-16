import type { ThemedToken } from "shiki";
import type { Comment } from "@/hooks/useComments";
import type { ActiveSelection } from "./codeViewerTypes";
import { indexToLine, lineOverlapsSelection } from "./sourceSelectionOffsets";

export interface SourceCommentOverlay {
  id: string;
  startCol: number;
  endCol: number;
  isSelected: boolean;
}

export interface SourceLineModel {
  index: number;
  lineNumber: number;
  rawLine: string;
  tokens: ThemedToken[] | null;
  isMatch: boolean;
  isCurrentMatch: boolean;
  comment: Comment | undefined;
  commentOverlays: SourceCommentOverlay[];
  selectionOverlaps: boolean;
  selectionStartCol: number;
  selectionEndCol: number;
  showDraftSelection: boolean;
}

export function sourceMatchLines(rawLines: string[], query: string): number[] {
  if (!query.trim()) return [];
  const normalizedQuery = query.toLowerCase();
  return rawLines.flatMap((line, index) =>
    line.toLowerCase().includes(normalizedQuery) ? [index] : [],
  );
}

export function buildSourceLineModels(
  rawLines: string[],
  tokenLines: ThemedToken[][] | null,
  comments: Comment[],
  activeSelection: ActiveSelection | null,
  searchQuery: string,
  currentMatchLine: number | undefined,
): SourceLineModel[] {
  const commentByLine = new Map<number, Comment>();
  for (const comment of comments) {
    const lineNumber = indexToLine(comment.start_index, rawLines);
    if (!commentByLine.has(lineNumber)) commentByLine.set(lineNumber, comment);
  }

  let lineStart = 0;
  return rawLines.map((rawLine, index) => {
    const lineNumber = index + 1;
    const isMatch =
      searchQuery.trim() !== "" &&
      rawLine.toLowerCase().includes(searchQuery.toLowerCase());
    const selectionOverlaps =
      activeSelection !== null &&
      lineOverlapsSelection(
        index,
        rawLines,
        activeSelection.start_index,
        activeSelection.end_index,
      );
    const selectionStartCol = selectionOverlaps
      ? Math.max(0, activeSelection.start_index - lineStart)
      : 0;
    const selectionEndCol = selectionOverlaps
      ? Math.min(rawLine.length, activeSelection.end_index - lineStart)
      : 0;
    const commentOverlays = comments
      .filter((comment) =>
        lineOverlapsSelection(index, rawLines, comment.start_index, comment.end_index),
      )
      .map((comment) => ({
        id: comment.id,
        startCol: Math.max(0, comment.start_index - lineStart),
        endCol: Math.min(rawLine.length, comment.end_index - lineStart),
        isSelected:
          activeSelection?.start_index === comment.start_index &&
          activeSelection?.end_index === comment.end_index,
      }))
      .filter((overlay) => overlay.endCol > overlay.startCol);
    const hasSavedSelection =
      activeSelection !== null &&
      comments.some(
        (comment) =>
          comment.start_index === activeSelection.start_index &&
          comment.end_index === activeSelection.end_index,
      );
    const model: SourceLineModel = {
      index,
      lineNumber,
      rawLine,
      tokens: tokenLines?.[index] ?? null,
      isMatch,
      isCurrentMatch: isMatch && currentMatchLine === index,
      comment: commentByLine.get(lineNumber),
      commentOverlays,
      selectionOverlaps,
      selectionStartCol,
      selectionEndCol,
      showDraftSelection: selectionOverlaps && !hasSavedSelection,
    };
    lineStart += rawLine.length + 1;
    return model;
  });
}

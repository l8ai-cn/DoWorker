import { useCallback, useMemo, type RefObject } from "react";
import type { ThemedToken } from "shiki";
import type { Comment } from "@/hooks/useComments";
import type { ActiveSelection } from "./codeViewerTypes";
import { SourceCodeLine } from "./SourceCodeLine";
import { SourceCodeSearchBar } from "./SourceCodeSearchBar";
import { SourceSelectionActions } from "./SourceSelectionActions";
import { buildSourceLineModels } from "./sourceCodeViewModel";
import { TruncatedBanner } from "./TruncatedBanner";
import type { CodeViewerSourceState } from "./useCodeViewerSourceState";

interface SourceCodeViewerProps {
  path: string;
  rawLines: string[];
  tokenLines: ThemedToken[][] | null;
  comments: Comment[];
  activeSelection: ActiveSelection | null;
  onSetActiveSelection: (selection: ActiveSelection | null) => void;
  truncated: boolean;
  searchOpen: boolean;
  setSearchOpen: (open: boolean) => void;
  searchInputRef: RefObject<HTMLInputElement | null>;
  sourceState: CodeViewerSourceState;
}

export function SourceCodeViewer({
  path,
  rawLines,
  tokenLines,
  comments,
  activeSelection,
  onSetActiveSelection,
  truncated,
  searchOpen,
  setSearchOpen,
  searchInputRef,
  sourceState,
}: SourceCodeViewerProps) {
  const { codeContainerRef, search, selection, setLineRef, canAttachToAgent } =
    sourceState;
  const lines = useMemo(
    () =>
      buildSourceLineModels(
        rawLines,
        tokenLines,
        comments,
        activeSelection,
        search.query,
        search.currentMatchLine,
      ),
    [
      activeSelection,
      comments,
      rawLines,
      search.currentMatchLine,
      search.query,
      tokenLines,
    ],
  );
  const selectComment = useCallback(
    (comment: Comment) =>
      onSetActiveSelection({
        start_index: comment.start_index,
        end_index: comment.end_index,
        anchor_content: comment.anchor_content ?? "",
      }),
    [onSetActiveSelection],
  );
  const closeSearch = () => {
    setSearchOpen(false);
    search.setQuery("");
  };

  return (
    <>
      {truncated && <TruncatedBanner />}
      {searchOpen && (
        <SourceCodeSearchBar
          inputRef={searchInputRef}
          query={search.query}
          matchCount={search.matches.length}
          currentMatchIndex={search.safeMatchIndex}
          onQueryChange={search.setQuery}
          onPrevious={search.previous}
          onNext={search.next}
          onClose={closeSearch}
        />
      )}
      <div ref={codeContainerRef} className="bg-white font-mono text-xs dark:bg-[#0d1117]">
        {lines.map((line) => (
          <SourceCodeLine
            key={line.lineNumber}
            line={line}
            searchQuery={search.query}
            setLineRef={setLineRef}
            onSelectComment={selectComment}
          />
        ))}
      </div>
      {selection.selectionAnchor && (
        <SourceSelectionActions
          anchor={selection.selectionAnchor}
          canAttachToAgent={canAttachToAgent}
          path={path}
          rawLines={rawLines}
          onSetActiveSelection={onSetActiveSelection}
          onClose={selection.clearSelectionAnchor}
        />
      )}
    </>
  );
}

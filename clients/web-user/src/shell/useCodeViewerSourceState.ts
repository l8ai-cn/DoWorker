import { useCallback, useRef } from "react";
import type { Comment } from "@/hooks/useComments";
import { nativeCodingAgentForHarness } from "@/lib/nativeCodingAgents";
import { useChatStore } from "@/store/chatStore";
import type { ActiveSelection, CodeViewerProps } from "./codeViewerTypes";
import { useSourceCodeKeyboard } from "./useSourceCodeKeyboard";
import { useSourceCodeSearch } from "./useSourceCodeSearch";
import { useSourceSelectionActions } from "./useSourceSelectionActions";

interface CodeViewerSourceStateOptions {
  path: string;
  rendererKey: string;
  content: string;
  rawLines: string[];
  comments: Comment[];
  activeSelection: ActiveSelection | null;
  onSetActiveSelection: (selection: ActiveSelection | null) => void;
  canEdit: boolean;
  keyboardEnabled: boolean;
  panelOpen: boolean;
  searchOpen: boolean;
  setSearchOpen: CodeViewerProps["setSearchOpen"];
  searchInputRef: CodeViewerProps["searchInputRef"];
}

export function useCodeViewerSourceState({
  path,
  rendererKey,
  content,
  rawLines,
  comments,
  activeSelection,
  onSetActiveSelection,
  canEdit,
  keyboardEnabled,
  panelOpen,
  searchOpen,
  setSearchOpen,
  searchInputRef,
}: CodeViewerSourceStateOptions) {
  const codeContainerRef = useRef<HTMLDivElement>(null);
  const lineRefs = useRef(new Map<number, HTMLDivElement>());
  const search = useSourceCodeSearch({
    rawLines,
    searchOpen,
    activeSelection,
    lineRefs,
  });
  const { setQuery } = search;
  const clearSearch = useCallback(() => setQuery(""), [setQuery]);
  useSourceCodeKeyboard({
    enabled: keyboardEnabled,
    panelOpen,
    searchOpen,
    setSearchOpen,
    searchInputRef,
    codeContainerRef,
    content,
    clearSearch,
  });
  const selection = useSourceSelectionActions({
    resetKey: rendererKey,
    rawLines,
    comments,
    canEdit,
    codeContainerRef,
    onSetActiveSelection,
  });
  const sessionHarness = useChatStore((state) => state.sessionHarness);
  const setLineRef = useCallback((index: number, element: HTMLDivElement | null) => {
    if (element) lineRefs.current.set(index, element);
    else lineRefs.current.delete(index);
  }, []);

  return {
    codeContainerRef,
    search,
    selection,
    setLineRef,
    canAttachToAgent:
      !!path && nativeCodingAgentForHarness(sessionHarness) !== undefined,
  };
}

export type CodeViewerSourceState = ReturnType<typeof useCodeViewerSourceState>;

import { useEffect, useRef, type RefObject } from "react";

interface SourceCodeKeyboardOptions {
  enabled: boolean;
  panelOpen: boolean;
  searchOpen: boolean;
  setSearchOpen: (open: boolean) => void;
  searchInputRef: RefObject<HTMLInputElement | null>;
  codeContainerRef: RefObject<HTMLDivElement | null>;
  content: string;
  clearSearch: () => void;
}

export function useSourceCodeKeyboard({
  enabled,
  panelOpen,
  searchOpen,
  setSearchOpen,
  searchInputRef,
  codeContainerRef,
  content,
  clearSearch,
}: SourceCodeKeyboardOptions) {
  const selectAllPendingRef = useRef(false);

  useEffect(() => {
    if (!enabled || !panelOpen) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === "f") {
        event.preventDefault();
        setSearchOpen(true);
        setTimeout(() => searchInputRef.current?.focus(), 0);
        return;
      }
      if ((event.metaKey || event.ctrlKey) && event.key === "a") {
        const container = codeContainerRef.current;
        const active = document.activeElement;
        if (
          !container ||
          active instanceof HTMLInputElement ||
          active instanceof HTMLTextAreaElement
        ) {
          return;
        }
        event.preventDefault();
        const selection = window.getSelection();
        if (!selection) return;
        const range = document.createRange();
        range.selectNodeContents(container);
        selection.removeAllRanges();
        selection.addRange(range);
        selectAllPendingRef.current = true;
        return;
      }
      if (event.key === "Escape" && searchOpen) {
        event.preventDefault();
        setSearchOpen(false);
        clearSearch();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [
    clearSearch,
    codeContainerRef,
    enabled,
    panelOpen,
    searchInputRef,
    searchOpen,
    setSearchOpen,
  ]);

  useEffect(() => {
    const handleCopy = (event: ClipboardEvent) => {
      if (!selectAllPendingRef.current) return;
      selectAllPendingRef.current = false;
      event.preventDefault();
      event.clipboardData?.setData("text/plain", content);
    };
    const clearPending = () => {
      selectAllPendingRef.current = false;
    };
    document.addEventListener("copy", handleCopy);
    document.addEventListener("mousedown", clearPending);
    return () => {
      document.removeEventListener("copy", handleCopy);
      document.removeEventListener("mousedown", clearPending);
    };
  }, [content]);
}

import type { RefObject } from "react";
import {
  ChevronDownIcon,
  ChevronUpIcon,
  SearchIcon,
  XIcon,
} from "lucide-react";

interface SourceCodeSearchBarProps {
  inputRef: RefObject<HTMLInputElement | null>;
  query: string;
  matchCount: number;
  currentMatchIndex: number;
  onQueryChange: (query: string) => void;
  onPrevious: () => void;
  onNext: () => void;
  onClose: () => void;
}

export function SourceCodeSearchBar({
  inputRef,
  query,
  matchCount,
  currentMatchIndex,
  onQueryChange,
  onPrevious,
  onNext,
  onClose,
}: SourceCodeSearchBarProps) {
  const resultLabel = query.trim()
    ? matchCount > 0
      ? `${currentMatchIndex + 1} / ${matchCount}`
      : "No results"
    : "";

  return (
    <div className="sticky top-0 z-10 flex items-center gap-2 border-b border-border bg-card/90 px-3 py-1.5 backdrop-blur">
      <SearchIcon className="size-3.5 shrink-0 text-muted-foreground" />
      <input
        ref={inputRef}
        type="text"
        value={query}
        onChange={(event) => onQueryChange(event.target.value)}
        onKeyDown={(event) => {
          if (event.key !== "Enter" || matchCount === 0) return;
          event.preventDefault();
          if (event.shiftKey) onPrevious();
          else onNext();
        }}
        placeholder="Find…"
        className="min-w-0 flex-1 bg-transparent text-xs outline-none"
      />
      <span className="shrink-0 text-xs text-muted-foreground">{resultLabel}</span>
      <button
        type="button"
        aria-label="Previous match"
        className="rounded p-0.5 text-muted-foreground hover:bg-muted disabled:opacity-40"
        disabled={matchCount === 0}
        onClick={onPrevious}
      >
        <ChevronUpIcon className="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Next match"
        className="rounded p-0.5 text-muted-foreground hover:bg-muted disabled:opacity-40"
        disabled={matchCount === 0}
        onClick={onNext}
      >
        <ChevronDownIcon className="size-3.5" />
      </button>
      <button
        type="button"
        aria-label="Close search"
        className="rounded p-0.5 text-muted-foreground hover:bg-muted"
        onClick={onClose}
      >
        <XIcon className="size-3.5" />
      </button>
    </div>
  );
}

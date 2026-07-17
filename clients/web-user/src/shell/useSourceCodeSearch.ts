import { useEffect, useMemo, useRef, useState, type RefObject } from "react";
import type { ActiveSelection } from "./codeViewerTypes";
import { indexToLine } from "./sourceSelectionOffsets";
import { sourceMatchLines } from "./sourceCodeViewModel";

interface SourceCodeSearchOptions {
  rawLines: string[];
  searchOpen: boolean;
  activeSelection: ActiveSelection | null;
  lineRefs: RefObject<Map<number, HTMLDivElement>>;
}

export function useSourceCodeSearch({
  rawLines,
  searchOpen,
  activeSelection,
  lineRefs,
}: SourceCodeSearchOptions) {
  const [query, setQuery] = useState("");
  const [currentMatchIndex, setCurrentMatchIndex] = useState(0);
  const matches = useMemo(() => sourceMatchLines(rawLines, query), [rawLines, query]);
  const safeMatchIndex = matches.length > 0 ? currentMatchIndex % matches.length : 0;
  const rawLinesRef = useRef(rawLines);
  const matchesRef = useRef(matches);
  rawLinesRef.current = rawLines;
  matchesRef.current = matches;

  useEffect(() => setCurrentMatchIndex(0), [query]);
  useEffect(() => {
    if (!searchOpen) setQuery("");
  }, [searchOpen]);
  useEffect(() => {
    if (activeSelection === null) return;
    const lineNumber = indexToLine(activeSelection.start_index, rawLinesRef.current);
    lineRefs.current.get(lineNumber - 1)?.scrollIntoView({
      block: "center",
      behavior: "smooth",
    });
  }, [activeSelection, lineRefs]);
  useEffect(() => {
    const currentMatches = matchesRef.current;
    const matchIndex =
      currentMatches.length > 0 ? currentMatchIndex % currentMatches.length : 0;
    const lineIndex = currentMatches[matchIndex];
    if (lineIndex === undefined) return;
    lineRefs.current.get(lineIndex)?.scrollIntoView({ block: "center", behavior: "smooth" });
  }, [currentMatchIndex, lineRefs, query]);

  const previous = () => {
    if (matches.length === 0) return;
    setCurrentMatchIndex((index) => (index - 1 + matches.length) % matches.length);
  };
  const next = () => {
    if (matches.length === 0) return;
    setCurrentMatchIndex((index) => (index + 1) % matches.length);
  };

  return {
    query,
    setQuery,
    matches,
    safeMatchIndex,
    currentMatchLine: matches[safeMatchIndex],
    previous,
    next,
  };
}

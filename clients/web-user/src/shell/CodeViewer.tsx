import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import type { BundledLanguage, ThemedToken } from "shiki";
import { highlightCode } from "@/components/ai-elements/code-block";
import { useCanEdit } from "@/hooks/usePermissions";
import type { CodeViewerProps } from "./codeViewerTypes";
import { FileImageViewer } from "./FileImageViewer";
import {
  detectLang,
  isBinaryPath,
  isImageFile,
} from "./fileContentClassification";
import { HtmlCommentViewer } from "./HtmlCommentViewer";
import { MarkdownPreview } from "./MarkdownPreview";
import { MarkdownRichTextViewer } from "./MarkdownRichTextViewer";
import { SourceCodeViewer } from "./SourceCodeViewer";
import { TruncatedBanner } from "./TruncatedBanner";
import { useCodeViewerSourceState } from "./useCodeViewerSourceState";

const MonacoCodeEditor = lazy(() =>
  import("./MonacoCodeEditor").then((module) => ({
    default: module.MonacoCodeEditor,
  })),
);

export type { CodeViewerProps } from "./codeViewerTypes";

export function CodeViewer({
  conversationId,
  path,
  fileQuery,
  comments,
  activeSelection,
  onSetActiveSelection,
  panelOpen,
  searchOpen,
  setSearchOpen,
  searchInputRef,
  viewMode,
  onDirtyChange,
  onSaveStatusChange,
  pendingBodyRef,
}: CodeViewerProps) {
  const canEdit = useCanEdit(conversationId);
  const [tokenLines, setTokenLines] = useState<ThemedToken[][] | null>(null);
  const content = fileQuery.data?.content ?? "";
  const truncated = fileQuery.data?.truncated ?? false;
  const lang = detectLang(path);
  const showMonaco = lang !== "markdown" && viewMode !== "preview";
  const rawLines = useMemo(
    () => (showMonaco ? [] : content.split("\n")),
    [content, showMonaco],
  );
  const sourceState = useCodeViewerSourceState({
    path,
    rendererKey: `${path}:${showMonaco}:${viewMode}`,
    content,
    rawLines,
    comments,
    activeSelection,
    onSetActiveSelection,
    canEdit,
    keyboardEnabled:
      !showMonaco && !(viewMode === "editor" && lang === "markdown"),
    panelOpen,
    searchOpen,
    setSearchOpen,
    searchInputRef,
  });

  useEffect(() => {
    if (showMonaco || (viewMode === "editor" && lang === "markdown")) return;
    let cancelled = false;
    setTokenLines(null);
    if (!content) return;
    const cached = highlightCode(content, lang as BundledLanguage, (result) => {
      if (!cancelled) setTokenLines(result.tokens);
    });
    if (cached) setTokenLines(cached.tokens);
    return () => {
      cancelled = true;
    };
  }, [content, lang, showMonaco, viewMode]);

  const handleSearchHandled = useCallback(
    () => setSearchOpen(false),
    [setSearchOpen],
  );

  if (fileQuery.isLoading) {
    return (
      <div className="flex items-center justify-center p-8 text-muted-foreground text-sm">
        Loading…
      </div>
    );
  }
  if (fileQuery.isError) {
    return (
      <div className="p-8 text-destructive text-sm">
        Error loading file:{" "}
        {fileQuery.error instanceof Error
          ? fileQuery.error.message
          : String(fileQuery.error)}
      </div>
    );
  }
  if (fileQuery.data && isImageFile(path, fileQuery.data.content_type)) {
    return <FileImageViewer data={fileQuery.data} path={path} />;
  }
  if (fileQuery.data?.encoding === "base64" || isBinaryPath(path)) {
    return (
      <div className="flex items-center justify-center p-8 text-muted-foreground text-sm">
        Preview not available for binary files.
      </div>
    );
  }
  if (viewMode === "editor" && lang === "markdown") {
    return (
      <MarkdownRichTextViewer
        content={content}
        conversationId={conversationId}
        path={path}
        isSettled={fileQuery.isSuccess}
        truncated={truncated}
        onDirtyChange={onDirtyChange}
        comments={comments}
        activeSelection={activeSelection}
        onSetActiveSelection={onSetActiveSelection}
        pendingBodyRef={pendingBodyRef}
      />
    );
  }
  if (viewMode === "preview" && lang === "html") {
    return <HtmlCommentViewer content={content} truncated={truncated} />;
  }
  if (viewMode === "preview" && lang === "markdown") {
    const preview = <MarkdownPreview content={content} />;
    if (!truncated) return preview;
    return (
      <div className="flex h-full flex-col">
        <TruncatedBanner />
        <div className="min-h-0 flex-1">{preview}</div>
      </div>
    );
  }
  if (showMonaco) {
    return (
      <Suspense
        fallback={
          <div className="flex items-center justify-center p-8 text-muted-foreground text-sm">
            Loading…
          </div>
        }
      >
        <MonacoCodeEditor
          content={content}
          conversationId={conversationId}
          path={path}
          isSettled={fileQuery.isSuccess}
          truncated={truncated}
          onDirtyChange={onDirtyChange}
          onSaveStatusChange={onSaveStatusChange}
          searchOpen={searchOpen}
          onSearchHandled={handleSearchHandled}
          comments={comments}
          activeSelection={activeSelection}
          onSetActiveSelection={onSetActiveSelection}
          pendingBodyRef={pendingBodyRef}
        />
      </Suspense>
    );
  }
  return (
    <SourceCodeViewer
      path={path}
      rawLines={rawLines}
      tokenLines={tokenLines}
      comments={comments}
      activeSelection={activeSelection}
      onSetActiveSelection={onSetActiveSelection}
      truncated={truncated}
      searchOpen={searchOpen}
      setSearchOpen={setSearchOpen}
      searchInputRef={searchInputRef}
      sourceState={sourceState}
    />
  );
}

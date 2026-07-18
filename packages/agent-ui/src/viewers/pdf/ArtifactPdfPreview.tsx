import { AlertCircle, Loader2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";
import { LazyPdfPageCanvas } from "./LazyPdfPageCanvas";
import type { PDFDocumentLoadingTask, PDFDocumentProxy } from "pdfjs-dist";

type PdfState =
  | { status: "loading" }
  | { status: "ready"; document: PDFDocumentProxy }
  | { status: "error" };

export function ArtifactPdfPreview({
  filename,
  src,
}: {
  filename: string;
  src: string;
}) {
  const text = useAgentWorkspaceText().artifact;
  const [state, setState] = useState<PdfState>({ status: "loading" });
  const scrollRootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let active = true;
    let loadingTask: PDFDocumentLoadingTask | null = null;
    let destroyed = false;
    const destroyLoadingTask = () => {
      if (destroyed || !loadingTask) return;
      destroyed = true;
      void loadingTask.destroy();
    };
    setState({ status: "loading" });

    void createPdfLoadingTask(src)
      .then(async (task) => {
        loadingTask = task;
        if (!active) {
          destroyLoadingTask();
          return;
        }
        const document = await task.promise;
        if (!active) {
          destroyLoadingTask();
          return;
        }
        setState({ status: "ready", document });
      })
      .catch((cause: unknown) => {
        if (!active) return;
        console.error("PDF preview failed", cause);
        setState({ status: "error" });
      });

    return () => {
      active = false;
      destroyLoadingTask();
    };
  }, [src]);

  if (state.status === "loading") {
    return (
      <div
        className="flex h-[32rem] items-center justify-center gap-2 border-b border-border bg-muted/30 text-sm text-muted-foreground"
        role="status"
      >
        <Loader2 className="size-4 animate-spin" />
        {text.loading(filename)}
      </div>
    );
  }
  if (state.status === "error") {
    return (
      <div
        className="flex h-[32rem] items-center justify-center gap-2 border-b border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive"
        role="alert"
      >
        <AlertCircle className="size-4 shrink-0" />
        {text.loadFailed}
      </div>
    );
  }

  return (
    <div
      aria-label={text.pdfPreview(filename)}
      className="flex h-[32rem] flex-col gap-3 overflow-auto border-b border-border bg-muted/40 p-3"
      ref={scrollRootRef}
    >
      {Array.from({ length: state.document.numPages }, (_, index) => (
        <LazyPdfPageCanvas
          document={state.document}
          filename={filename}
          key={index + 1}
          pageNumber={index + 1}
          scrollRootRef={scrollRootRef}
        />
      ))}
    </div>
  );
}

async function createPdfLoadingTask(src: string) {
  const pdfjs = await import("pdfjs-dist");
  pdfjs.GlobalWorkerOptions.workerSrc = new URL(
    "pdfjs-dist/build/pdf.worker.min.mjs",
    import.meta.url,
  ).toString();
  return pdfjs.getDocument({ url: src });
}

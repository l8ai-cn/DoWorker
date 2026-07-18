import { AlertCircle } from "lucide-react";
import { useEffect, useRef, useState, type RefObject } from "react";
import type {
  PDFDocumentProxy,
  PDFPageProxy,
  RenderTask,
} from "pdfjs-dist";

export function LazyPdfPageCanvas({
  document,
  filename,
  pageNumber,
  scrollRootRef,
}: {
  document: PDFDocumentProxy;
  filename: string;
  pageNumber: number;
  scrollRootRef: RefObject<HTMLDivElement | null>;
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const hostRef = useRef<HTMLDivElement>(null);
  const [failed, setFailed] = useState(false);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;
    if (typeof IntersectionObserver === "undefined") {
      setVisible(true);
      return;
    }
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) {
          setVisible(true);
          observer.disconnect();
        }
      },
      { root: scrollRootRef.current, rootMargin: "600px 0px" },
    );
    observer.observe(host);
    return () => observer.disconnect();
  }, [scrollRootRef]);

  useEffect(() => {
    const host = hostRef.current;
    const canvas = canvasRef.current;
    if (!visible || !host || !canvas) return;

    let active = true;
    let page: PDFPageProxy | null = null;
    let pagePromise: Promise<PDFPageProxy> | null = null;
    let renderTask: RenderTask | null = null;
    let renderSequence = 0;

    const getPage = () => {
      pagePromise ??= document.getPage(pageNumber).then((loadedPage) => {
        page = loadedPage;
        return loadedPage;
      });
      return pagePromise;
    };
    const render = async () => {
      const sequence = ++renderSequence;
      const loadedPage = await getPage();
      if (!active || sequence !== renderSequence || host.clientWidth === 0) return;

      renderTask?.cancel();
      const baseViewport = loadedPage.getViewport({ scale: 1 });
      const cssWidth = Math.min(host.clientWidth, baseViewport.width);
      const pixelRatio = Math.min(window.devicePixelRatio || 1, 2);
      const viewport = loadedPage.getViewport({
        scale: (cssWidth / baseViewport.width) * pixelRatio,
      });
      canvas.width = Math.floor(viewport.width);
      canvas.height = Math.floor(viewport.height);
      canvas.style.width = `${cssWidth}px`;
      canvas.style.height = `${viewport.height / pixelRatio}px`;
      renderTask = loadedPage.render({ canvas, viewport });
      await renderTask.promise;
      if (active && sequence === renderSequence) setFailed(false);
    };
    const renderSafely = () => {
      void render().catch((cause: unknown) => {
        if (!active || isCancelledRender(cause)) return;
        console.error(`PDF page ${pageNumber} failed`, cause);
        setFailed(true);
      });
    };
    const observer = new ResizeObserver(renderSafely);
    observer.observe(host);
    renderSafely();

    return () => {
      active = false;
      observer.disconnect();
      renderTask?.cancel();
      page?.cleanup();
    };
  }, [document, pageNumber, visible]);

  return (
    <div
      aria-label={`${filename} · ${pageNumber}`}
      className="flex min-h-24 w-full shrink-0 items-center justify-center"
      ref={hostRef}
    >
      {failed ? (
        <AlertCircle className="size-5 text-destructive" />
      ) : (
        <canvas
          aria-label={`${filename} · ${pageNumber}`}
          className="max-w-full rounded border border-border bg-white"
          ref={canvasRef}
        />
      )}
    </div>
  );
}

function isCancelledRender(cause: unknown) {
  return cause instanceof Error && cause.name === "RenderingCancelledException";
}

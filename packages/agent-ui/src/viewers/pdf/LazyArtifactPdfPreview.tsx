import { AlertCircle, Loader2 } from "lucide-react";
import { useEffect, useState, type ComponentType } from "react";

import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";

type PdfPreviewComponent = ComponentType<{
  filename: string;
  src: string;
}>;

type LazyPdfState =
  | { status: "loading" }
  | { status: "ready"; Preview: PdfPreviewComponent }
  | { status: "error" };

export function LazyArtifactPdfPreview({
  filename,
  src,
}: {
  filename: string;
  src: string;
}) {
  const text = useAgentWorkspaceText().artifact;
  const [state, setState] = useState<LazyPdfState>({ status: "loading" });

  useEffect(() => {
    let active = true;
    void import("./ArtifactPdfPreview").then(
      ({ ArtifactPdfPreview }) => {
        if (active) setState({ status: "ready", Preview: ArtifactPdfPreview });
      },
      (cause: unknown) => {
        if (!active) return;
        console.error("PDF preview failed", cause);
        setState({ status: "error" });
      },
    );
    return () => {
      active = false;
    };
  }, []);

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

  return <state.Preview filename={filename} src={src} />;
}

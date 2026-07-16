import { useEffect, useState } from "react";

import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "./contracts";
import { artifactPresentation } from "./artifactPresentation";

export type ArtifactBlobState =
  | { status: "loading" }
  | {
      status: "ready";
      url: string;
      mimeType: string | null;
      text: string | null;
    }
  | { status: "error"; message: string };

export function useArtifactBlobUrl(
  item: AgentArtifactItem,
  runtime: AgentSessionRuntime,
  sessionId: string,
): ArtifactBlobState {
  const [state, setState] = useState<ArtifactBlobState>({ status: "loading" });

  useEffect(() => {
    if (item.status === "failed") {
      setState({ status: "error", message: "Artifact generation failed" });
      return;
    }
    if (!runtime.loadArtifact) {
      setState({ status: "error", message: "Artifact loading is unavailable" });
      return;
    }

    let active = true;
    let objectUrl: string | null = null;
    setState({ status: "loading" });
    void runtime
      .loadArtifact(sessionId, item.artifactId)
      .then(async (blob) => {
        if (!active) return;
        objectUrl = URL.createObjectURL(blob);
        const mimeType = item.mimeType || blob.type || null;
        const kind = artifactPresentation(mimeType, item.filename).kind;
        const text =
          kind === "code" || kind === "html" || kind === "text"
            ? await readBlobText(blob)
            : null;
        if (!active) return;
        setState({
          status: "ready",
          url: objectUrl,
          mimeType,
          text,
        });
      })
      .catch((cause: unknown) => {
        if (!active) return;
        setState({
          status: "error",
          message: cause instanceof Error ? cause.message : String(cause),
        });
      });

    return () => {
      active = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [item.artifactId, item.mimeType, item.status, runtime, sessionId]);

  return state;
}

function readBlobText(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.addEventListener("load", () => resolve(String(reader.result ?? "")));
    reader.addEventListener("error", () => reject(reader.error));
    reader.readAsText(blob);
  });
}

import { useEffect, useState } from "react";

import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "./contracts";
import { artifactActionAllowed } from "./artifactGrantActions";
import { artifactPresentation } from "./artifactPresentation";

export type ArtifactBlobState =
  | { status: "idle" }
  | { status: "loading" }
  | {
      status: "ready";
      url: string;
      mimeType: string | null;
      text: string | null;
      textTruncated: boolean;
    }
  | {
      status: "error";
      code:
        | "generation_failed"
        | "loading_unavailable"
        | "authorization_required"
        | "load_failed";
      retryable: boolean;
      message: string;
    };

export function useArtifactBlobUrl(
  item: AgentArtifactItem,
  runtime: AgentSessionRuntime,
  sessionId: string,
  enabled = true,
  attempt = 0,
): ArtifactBlobState {
  const [state, setState] = useState<ArtifactBlobState>(
    enabled ? { status: "loading" } : { status: "idle" },
  );

  useEffect(() => {
    if (item.status === "failed") {
      setState({
        status: "error",
        code: "generation_failed",
        retryable: false,
        message: "Artifact generation failed",
      });
      return;
    }
    if (!enabled) {
      setState({ status: "idle" });
      return;
    }
    if (item.status !== "completed") {
      setState({ status: "loading" });
      return;
    }
    if (!artifactActionAllowed(
      item,
      "artifact.download",
      item.selectedRepresentationId ?? undefined,
    )) {
      setState({
        status: "error",
        code: "authorization_required",
        retryable: false,
        message: "artifact_download_not_authorized",
      });
      return;
    }
    if (!runtime.loadArtifact) {
      setState({
        status: "error",
        code: "loading_unavailable",
        retryable: false,
        message: "Artifact loading is unavailable",
      });
      return;
    }

    let active = true;
    let objectUrl: string | null = null;
    setState({ status: "loading" });
    const pending = item.selectedRepresentationId
      ? runtime.loadArtifact(
          sessionId,
          item.artifactId,
          item.selectedRepresentationId,
        )
      : runtime.loadArtifact(sessionId, item.artifactId);
    void pending
      .then(async (blob) => {
        if (!active) return;
        objectUrl = URL.createObjectURL(blob);
        const mimeType = blob.type || item.mimeType || null;
        const kind = artifactPresentation(mimeType, item.filename).kind;
        const readsText =
          kind === "code" ||
          kind === "csv" ||
          kind === "html" ||
          kind === "markdown" ||
          kind === "text";
        const text = readsText
          ? await readBlobText(blob.slice(0, MAX_TEXT_PREVIEW_BYTES))
          : null;
        if (!active) return;
        setState({
          status: "ready",
          url: objectUrl,
          mimeType,
          text,
          textTruncated: readsText && blob.size > MAX_TEXT_PREVIEW_BYTES,
        });
      })
      .catch((cause: unknown) => {
        if (!active) return;
        if (objectUrl) {
          URL.revokeObjectURL(objectUrl);
          objectUrl = null;
        }
        console.error("Artifact loading failed", cause);
        setState({
          status: "error",
          code: "load_failed",
          retryable: true,
          message: cause instanceof Error ? cause.message : String(cause),
        });
      });

    return () => {
      active = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [
    attempt,
    enabled,
    item.artifactId,
    item.filename,
    item.mimeType,
    item.selectedRepresentationId,
    item.status,
    runtime,
    sessionId,
  ]);

  return state;
}

export const MAX_TEXT_PREVIEW_BYTES = 2 << 20;

function readBlobText(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.addEventListener("load", () => resolve(String(reader.result ?? "")));
    reader.addEventListener("error", () => reject(reader.error));
    reader.readAsText(blob);
  });
}

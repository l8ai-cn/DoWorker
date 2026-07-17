import { useEffect, useState } from "react";

import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "./contracts";

export type ArtifactRepresentationUrlState =
  | { status: "loading" }
  | { status: "ready"; mimeType: string; url: string }
  | { status: "error"; message: string };

export function useArtifactRepresentationUrls(
  item: AgentArtifactItem,
  runtime: AgentSessionRuntime,
  sessionId: string,
  representationIds: readonly string[],
): Readonly<Record<string, ArtifactRepresentationUrlState>> {
  const uniqueIds = [...new Set(representationIds.filter(Boolean))];
  const requestedRepresentations = uniqueIds.map((representationId) => {
    const representation = item.representations.find(
      (candidate) => candidate.representationId === representationId,
    );
    return {
      mediaType: representation?.mediaType ?? "",
      representationId,
      revision: representation?.revision.toString() ?? "",
      status: representation?.status ?? "",
    };
  });
  const requestKey = JSON.stringify({
    artifactId: item.artifactId,
    artifactRevision: item.revision.toString(),
    artifactStatus: item.status,
    representations: requestedRepresentations,
  });
  const [states, setStates] = useState<
    Record<string, ArtifactRepresentationUrlState>
  >(() => loadingStates(uniqueIds));

  useEffect(() => {
    setStates(loadingStates(uniqueIds));
    if (uniqueIds.length === 0) return;
    if (item.status === "failed") {
      setStates(errorStates(uniqueIds, "Artifact generation failed"));
      return;
    }
    if (!runtime.loadArtifact) {
      setStates(errorStates(uniqueIds, "Artifact loading is unavailable"));
      return;
    }

    let active = true;
    const urls = new Set<string>();
    for (const representationId of uniqueIds) {
      void runtime
        .loadArtifact(sessionId, item.artifactId, representationId)
        .then((blob) => {
          if (!active) return;
          const url = URL.createObjectURL(blob);
          urls.add(url);
          const representation = requestedRepresentations.find(
            (candidate) => candidate.representationId === representationId,
          );
          setStates((current) => ({
            ...current,
            [representationId]: {
              status: "ready",
              mimeType: representation?.mediaType || blob.type,
              url,
            },
          }));
        })
        .catch((cause: unknown) => {
          if (!active) return;
          setStates((current) => ({
            ...current,
            [representationId]: {
              status: "error",
              message: cause instanceof Error ? cause.message : String(cause),
            },
          }));
        });
    }

    return () => {
      active = false;
      urls.forEach((url) => URL.revokeObjectURL(url));
    };
  }, [requestKey, runtime, sessionId]);

  return states;
}

function loadingStates(
  representationIds: readonly string[],
): Record<string, ArtifactRepresentationUrlState> {
  return Object.fromEntries(
    representationIds.map((representationId) => [
      representationId,
      { status: "loading" },
    ]),
  );
}

function errorStates(
  representationIds: readonly string[],
  message: string,
): Record<string, ArtifactRepresentationUrlState> {
  return Object.fromEntries(
    representationIds.map((representationId) => [
      representationId,
      { status: "error", message },
    ]),
  );
}

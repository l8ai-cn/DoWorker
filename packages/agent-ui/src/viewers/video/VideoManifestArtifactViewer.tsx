import { useEffect, useMemo, useState } from "react";

import type { AgentVideoManifest } from "../../agentArtifactContracts";
import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";
import type { AgentContentRendererProps } from "../../react/contentRendererTypes";
import { useArtifactRepresentationUrls } from "../../useArtifactRepresentationUrls";
import {
  ArtifactViewerError,
  ArtifactViewerLoading,
} from "../ArtifactViewerStatus";
import {
  VideoArtifactViewer,
} from "./VideoArtifactViewer";
import {
  videoPosterRepresentation,
  videoRepresentations,
  videoVersions,
} from "./videoManifestRepresentations";

export function VideoManifestArtifactViewer({
  filename,
  item,
  presentation = "developer",
  runtime,
  sessionId,
}: AgentContentRendererProps) {
  const text = useAgentWorkspaceText().artifact;
  const manifest = item.manifest;
  if (manifest?.kind !== "video") {
    return (
      <ArtifactViewerError
        filename={filename}
        message={
          presentation === "user"
            ? text.loadFailed
            : "video_manifest_missing"
        }
      />
    );
  }
  if (manifest.stage === "unknown") {
    return (
      <ArtifactViewerError
        filename={filename}
        message={
          presentation === "user"
            ? text.loadFailed
            : "video_stage_unknown"
        }
      />
    );
  }
  return (
    <VideoManifestArtifactContent
      filename={filename}
      item={item}
      manifest={manifest}
      presentation={presentation}
      runtime={runtime}
      sessionId={sessionId}
      stage={manifest.stage}
    />
  );
}

function VideoManifestArtifactContent({
  filename,
  item,
  manifest,
  presentation = "developer",
  runtime,
  sessionId,
  stage,
}: AgentContentRendererProps & {
  manifest: AgentVideoManifest;
  stage: Exclude<AgentVideoManifest["stage"], "unknown">;
}) {
  const text = useAgentWorkspaceText().artifact;
  const versionRepresentations = useMemo(
    () => videoRepresentations(item.representations, manifest),
    [item.representations, manifest],
  );
  const posterRepresentation = useMemo(
    () => videoPosterRepresentation(item.representations, manifest),
    [item.representations, manifest],
  );
  const versionIds = new Set(
    versionRepresentations.map((value) => value.representationId),
  );
  const initialId =
    [manifest.playableRepresentationId, manifest.originalRepresentationId]
      .find((id): id is string => Boolean(id && versionIds.has(id))) ??
    versionRepresentations[0]?.representationId ??
    "";
  const [selectedId, setSelectedId] = useState(initialId);
  useEffect(() => {
    if (!versionRepresentations.some((value) => value.representationId === selectedId)) {
      setSelectedId(initialId);
    }
  }, [initialId, selectedId, versionRepresentations]);

  const representationIds = [
    ...(selectedId ? [selectedId] : []),
    ...(posterRepresentation
      ? [posterRepresentation.representationId]
      : []),
  ];
  const resources = useArtifactRepresentationUrls(
    item,
    runtime,
    sessionId,
    representationIds,
  );
  const versions = useMemo(
    () => videoVersions(versionRepresentations, resources),
    [resources, versionRepresentations],
  );
  const selected = versions.find((version) => version.id === selectedId);
  const posterId = posterRepresentation?.representationId;
  const poster = posterId
    ? resources[posterId]
    : undefined;
  if (stage === "ready" && versionRepresentations.length === 0) {
    return (
      <ArtifactViewerError
        filename={filename}
        message={
          presentation === "user"
            ? text.loadFailed
            : "video_playable_representation_missing"
        }
      />
    );
  }
  if (manifest.stage === "ready" && selected?.src === undefined) {
    const failed = selectedId && resources[selectedId]?.status === "error";
    return failed ? (
      <ArtifactViewerError
        filename={filename}
        message={
          presentation === "user"
            ? text.loadFailed
            : resources[selectedId]?.status === "error"
            ? resources[selectedId].message
            : "video_representation_load_failed"
        }
      />
    ) : (
      <ArtifactViewerLoading filename={filename} />
    );
  }

  return (
    <VideoArtifactViewer
      durationSeconds={
        manifest.durationMillis === undefined
          ? undefined
          : Number(manifest.durationMillis) / 1000
      }
      filename={selected?.filename ?? filename}
      mimeType={selected?.mimeType ?? item.mimeType ?? "video/mp4"}
      onDownload={
        selected?.src ? () => download(selected.src!, selected.filename) : undefined
      }
      onSelectVersion={setSelectedId}
      posterSrc={poster?.status === "ready" ? poster.url : undefined}
      progress={
        manifest.progressFraction === undefined
          ? undefined
          : manifest.progressFraction * 100
      }
      selectedVersionId={selectedId}
      src={selected?.src ?? ""}
      status={stage}
      technicalMetadata={presentation === "developer"}
      versions={versions}
    />
  );
}

function download(url: string, filename?: string) {
  const link = document.createElement("a");
  link.href = url;
  link.download = filename || "video";
  link.click();
}

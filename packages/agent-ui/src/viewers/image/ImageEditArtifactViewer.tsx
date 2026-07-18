import { useState } from "react";

import { artifactActionError } from "../../artifactActionError";
import { artifactActionAllowed } from "../../artifactGrantActions";
import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";
import type { AgentImageEditManifest } from "../../contracts";
import type { AgentContentRendererProps } from "../../react/contentRendererTypes";
import { useArtifactRepresentationUrls } from "../../useArtifactRepresentationUrls";
import {
  ArtifactViewerError,
  ArtifactViewerLoading,
} from "../ArtifactViewerStatus";
import { ImageComparisonViewer } from "./ImageComparisonViewer";
import { ImageEditComposer } from "./ImageEditComposer";

export function ImageEditArtifactViewer({
  filename,
  item,
  presentation = "developer",
  runtime,
  sessionId,
}: AgentContentRendererProps) {
  const text = useAgentWorkspaceText();
  const manifest = item.manifest;
  if (manifest?.kind !== "image_edit") {
    return imageEditError(filename, presentation, text, "image_edit_manifest_missing");
  }
  return (
    <ImageEditArtifactContent
      filename={filename}
      item={item}
      manifest={manifest}
      presentation={presentation}
      runtime={runtime}
      sessionId={sessionId}
    />
  );
}
function ImageEditArtifactContent({
  filename,
  item,
  manifest,
  presentation = "developer",
  runtime,
  sessionId,
}: AgentContentRendererProps & { manifest: AgentImageEditManifest }) {
  const text = useAgentWorkspaceText();
  const resultRepresentationId =
    selectResultRepresentationId(item, manifest);
  const representationIds = new Set(
    item.representations.map((representation) => representation.representationId),
  );
  if (!representationIds.has(manifest.sourceRepresentationId)) {
    return imageEditError(
      filename,
      presentation,
      text,
      "image_edit_source_representation_missing",
    );
  }
  if (
    manifest.resultRepresentationId &&
    !representationIds.has(manifest.resultRepresentationId)
  ) {
    return imageEditError(
      filename,
      presentation,
      text,
      "image_edit_result_representation_missing",
    );
  }
  const requestedRepresentationIds = [
    manifest.sourceRepresentationId,
    ...(resultRepresentationId ? [resultRepresentationId] : []),
  ];
  const resources = useArtifactRepresentationUrls(
    item,
    runtime,
    sessionId,
    requestedRepresentationIds,
  );
  const [regionIndex, setRegionIndex] = useState(0);
  const [actionError, setActionError] = useState<string | null>(null);
  const [actionPending, setActionPending] = useState(false);
  const source = resources[manifest.sourceRepresentationId];
  const result = resultRepresentationId
    ? resources[resultRepresentationId]
    : undefined;
  const error = requestedRepresentationIds
    .map((id) => resources[id])
    .find((state) => state?.status === "error");
  const loading = requestedRepresentationIds.some(
    (id) => resources[id]?.status !== "ready",
  );
  if (error?.status === "error") {
    return imageEditError(filename, presentation, text, error.message);
  }
  if (loading || source?.status !== "ready") {
    return <ArtifactViewerLoading filename={filename} />;
  }
  const canEdit =
    artifactActionAllowed(
      item,
      "image.edit",
      manifest.sourceRepresentationId,
    ) &&
    runtime.executeArtifactAction !== undefined;
  const normalizedRegion = manifest.regions[regionIndex];
  return (
    <section className="space-y-3">
      {result?.status === "ready" ? (
        <ImageComparisonViewer
          result={{ alt: "结果图", src: result.url }}
          source={{ alt: "源图", src: source.url }}
        />
      ) : (
        <img
          alt="源图"
          className="max-h-[70vh] w-full rounded-md border border-border bg-muted object-contain"
          src={source.url}
        />
      )}
      {manifest.regions.length > 1 && (
        <label className="flex items-center gap-2 text-sm">
          <span>编辑区域</span>
          <select
            aria-label="选择编辑区域"
            className="h-11 rounded-md border border-input bg-background px-2"
            onChange={(event) => setRegionIndex(Number(event.target.value))}
            value={regionIndex}
          >
            {manifest.regions.map((_region, index) => (
              <option key={index} value={index}>区域 {index + 1}</option>
            ))}
          </select>
        </label>
      )}
      {!canEdit && (
        <div className="text-xs text-muted-foreground" role="status">
          当前会话没有图片编辑动作权限
        </div>
      )}
      {actionPending && (
        <div className="text-xs text-muted-foreground" role="status">
          正在提交图片编辑
        </div>
      )}
      {actionError && (
        <div className="text-xs text-destructive" role="alert">
          {actionError}
        </div>
      )}
      <ImageEditComposer
        actionSchemaVersion={item.schemaVersion}
        artifactId={item.artifactId}
        baseRevision={item.revision}
        disabled={!canEdit || actionPending}
        normalizedRegion={normalizedRegion}
        onSubmit={(command) => {
          setActionError(null);
          setActionPending(true);
          void runtime
            .executeArtifactAction!(sessionId, command)
            .catch((cause: unknown) => setActionError(artifactActionError(cause)))
            .finally(() => setActionPending(false));
        }}
        representationId={manifest.sourceRepresentationId}
        sourceDimensions={manifest.sourceDimensions}
      />
    </section>
  );
}
function selectResultRepresentationId(
  item: AgentContentRendererProps["item"],
  manifest: AgentImageEditManifest,
): string | undefined {
  if (manifest.resultRepresentationId) return manifest.resultRepresentationId;
  const representationIds = new Set(
    item.representations.map((representation) => representation.representationId),
  );
  return manifest.candidateRepresentationIds.find((id) =>
    representationIds.has(id),
  );
}
function imageEditError(
  filename: string,
  presentation: AgentContentRendererProps["presentation"],
  text: ReturnType<typeof useAgentWorkspaceText>,
  developerMessage: string,
) {
  return (
    <ArtifactViewerError
      filename={filename}
      message={
        presentation === "user" ? text.artifact.loadFailed : developerMessage
      }
    />
  );
}

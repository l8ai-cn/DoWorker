import { useState } from "react";

import { artifactActionError } from "../../artifactActionError";
import { artifactActionAllowed } from "../../artifactGrantActions";
import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";
import type { AgentPresentationManifest } from "../../contracts";
import type { AgentContentRendererProps } from "../../react/contentRendererTypes";
import { useArtifactRepresentationUrls } from "../../useArtifactRepresentationUrls";
import {
  ArtifactViewerError,
  ArtifactViewerLoading,
} from "../ArtifactViewerStatus";
import { PresentationArtifactViewer } from "./PresentationArtifactViewer";
import {
  presentationManifestSlides,
  presentationRepresentationIds,
} from "./presentationManifestSlides";
import {
  PRESENTATION_GRANTS,
  type PresentationGrant,
  type PresentationSlide,
} from "./presentationContracts";

export function PresentationManifestArtifactViewer({
  filename,
  item,
  presentation = "developer",
  runtime,
  sessionId,
}: AgentContentRendererProps) {
  const text = useAgentWorkspaceText();
  const manifest = item.manifest;
  if (manifest?.kind !== "presentation") {
    return (
      <ArtifactViewerError
        filename={filename}
        message={
          presentation === "user"
            ? text.artifact.loadFailed
            : "presentation_manifest_missing"
        }
      />
    );
  }
  return (
    <PresentationManifestArtifactContent
      filename={filename}
      item={item}
      manifest={manifest}
      presentation={presentation}
      runtime={runtime}
      sessionId={sessionId}
    />
  );
}

function PresentationManifestArtifactContent({
  filename,
  item,
  manifest,
  presentation = "developer",
  runtime,
  sessionId,
}: AgentContentRendererProps & { manifest: AgentPresentationManifest }) {
  const text = useAgentWorkspaceText();
  const renderableSlides = presentationManifestSlides(
    manifest,
    item.representations,
  );
  const representationIds = presentationRepresentationIds(renderableSlides);
  const resources = useArtifactRepresentationUrls(
    item,
    runtime,
    sessionId,
    representationIds,
  );
  const [actionError, setActionError] = useState<string | null>(null);
  const [actionPending, setActionPending] = useState(false);
  const error = representationIds
    .map((id) => resources[id])
    .find((state) => state?.status === "error");
  if (error?.status === "error") {
    return (
      <ArtifactViewerError
        filename={filename}
        message={
          presentation === "user" ? text.artifact.loadFailed : error.message
        }
      />
    );
  }
  if (representationIds.some((id) => resources[id]?.status !== "ready")) {
    return <ArtifactViewerLoading filename={filename} />;
  }
  const slides = renderableSlides.flatMap<PresentationSlide>((slide) => {
    const page = resources[slide.pageRepresentationId];
    const thumbnail = slide.thumbnailRepresentationId
      ? resources[slide.thumbnailRepresentationId]
      : undefined;
    if (page?.status !== "ready") return [];
    return [{
      imageSrc: page.url,
      notes: slide.notes,
      position: slide.position,
      representationId: slide.pageRepresentationId,
      slideId: slide.slideId,
      thumbnailSrc: thumbnail?.status === "ready" ? thumbnail.url : undefined,
      title: slide.title,
    }];
  });
  const canExecute = runtime.executeArtifactAction !== undefined;
  const actionRepresentationIds = [
    ...(item.selectedRepresentationId ? [item.selectedRepresentationId] : []),
    ...slides.flatMap((slide) =>
      slide.representationId ? [slide.representationId] : [],
    ),
  ];
  const grants = canExecute && !actionPending
    ? item.actions.filter(
        (action): action is PresentationGrant =>
          isPresentationGrant(action) &&
          actionRepresentationIds.some((representationId) =>
            artifactActionAllowed(item, action, representationId),
          ),
      )
    : [];
  const selectedVersionId =
    manifest.selectedVersionId ?? manifest.versions[0]?.id ?? "";
  const canSelectVersion =
    canExecute &&
    !actionPending &&
    artifactActionAllowed(
      item,
      PRESENTATION_GRANTS.selectVersion,
      item.selectedRepresentationId ?? undefined,
    );
  const execute = (
    command: Parameters<NonNullable<typeof runtime.executeArtifactAction>>[1],
  ) => {
    setActionError(null);
    setActionPending(true);
    void runtime
      .executeArtifactAction!(sessionId, command)
      .catch((cause: unknown) => setActionError(artifactActionError(cause)))
      .finally(() => setActionPending(false));
  };

  return (
    <section className="space-y-2">
      {actionPending && (
        <div className="text-xs text-muted-foreground" role="status">
          正在提交演示文稿操作
        </div>
      )}
      {actionError && (
        <div className="text-xs text-destructive" role="alert">
          {actionError}
        </div>
      )}
      <PresentationArtifactViewer
        actionSchemaVersion={item.schemaVersion}
        artifactId={item.artifactId}
        baseRevision={item.revision}
        grants={grants}
        onAction={execute}
        onSelectVersion={
          canSelectVersion
            ? (versionId) =>
                execute({
                  actionSchemaVersion: item.schemaVersion,
                  actionType: PRESENTATION_GRANTS.selectVersion,
                  artifactId: item.artifactId,
                  baseRevision: item.revision,
                  commandId: crypto.randomUUID(),
                  payload: { versionId },
                  ...(item.selectedRepresentationId
                    ? { representationId: item.selectedRepresentationId }
                    : {}),
                })
            : undefined
        }
        representationId={item.selectedRepresentationId ?? undefined}
        selectedVersionId={selectedVersionId}
        slides={slides}
        versions={manifest.versions}
      />
    </section>
  );
}

function isPresentationGrant(action: string): action is PresentationGrant {
  return Object.values(PRESENTATION_GRANTS).includes(
    action as PresentationGrant,
  );
}

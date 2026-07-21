import {
  VideoStage,
  type ArtifactManifest,
} from "@agent-cloud/proto/agent_workbench/v2/artifact_pb";
import {
  UnsupportedReason,
  type StructuredPayload,
} from "@agent-cloud/proto/agent_workbench/v2/content_pb";

import type {
  AgentArtifactManifest,
  AgentStructuredPayload,
} from "../agentArtifactContracts";

export function projectManifest(
  manifest: ArtifactManifest | undefined,
): AgentArtifactManifest | null {
  const value = manifest?.manifest;
  if (!value?.case) return null;
  if (value.case === "imageEdit") {
    return {
      kind: "image_edit",
      sourceRepresentationId: value.value.sourceRepresentationId,
      ...(value.value.resultRepresentationId
        ? { resultRepresentationId: value.value.resultRepresentationId }
        : {}),
      candidateRepresentationIds: [...value.value.candidateRepresentationIds],
      ...(value.value.maskRepresentationId
        ? { maskRepresentationId: value.value.maskRepresentationId }
        : {}),
      sourceDimensions: {
        width: value.value.sourceWidth,
        height: value.value.sourceHeight,
      },
      ...(value.value.exifOrientation
        ? { exifOrientation: value.value.exifOrientation }
        : {}),
      regions: value.value.regions.map((region) => ({ ...region })),
      annotations: value.value.annotations.map((annotation) => ({
        annotationId: annotation.annotationId,
        path: annotation.path.map((point) => ({ ...point })),
        ...(annotation.label ? { label: annotation.label } : {}),
        ...(annotation.style
          ? { style: structuredPayload(annotation.style) }
          : {}),
      })),
    };
  }
  if (value.case === "video") {
    return {
      kind: "video",
      stage: videoStage(value.value.stage),
      ...(value.value.progressFraction !== undefined
        ? { progressFraction: value.value.progressFraction }
        : {}),
      ...(value.value.durationMillis !== undefined
        ? { durationMillis: value.value.durationMillis }
        : {}),
      ...(value.value.dimensions
        ? { dimensions: { ...value.value.dimensions } }
        : {}),
      ...(value.value.originalRepresentationId
        ? { originalRepresentationId: value.value.originalRepresentationId }
        : {}),
      ...(value.value.playableRepresentationId
        ? { playableRepresentationId: value.value.playableRepresentationId }
        : {}),
      ...(value.value.posterRepresentationId
        ? { posterRepresentationId: value.value.posterRepresentationId }
        : {}),
      thumbnailRepresentationIds: [...value.value.thumbnailRepresentationIds],
      derivativeRepresentationIds: [...value.value.derivativeRepresentationIds],
    };
  }
  if (value.case === "presentation") {
    return {
      kind: "presentation",
      deckRevision: value.value.deckRevision,
      slides: value.value.slides.map((slide) => ({ ...slide })),
      versions: value.value.versions.map((version, index) => ({
        id: version.versionId,
        revision: version.revision,
        label: version.label || `Version ${index + 1}`,
      })),
      ...(value.value.selectedVersionId
        ? { selectedVersionId: value.value.selectedVersionId }
        : {}),
    };
  }
  if (value.case === "extension") {
    return {
      kind: "extension",
      namespace: value.value.namespace,
      semanticType: value.value.semanticType,
      schemaVersion: value.value.schemaVersion,
      ...(value.value.payload
        ? { payload: structuredPayload(value.value.payload) }
        : {}),
    };
  }
  return {
    kind: "unsupported",
    reason: unsupportedReason(value.value.reason),
    ...(value.value.identity
      ? { identity: { ...value.value.identity } }
      : {}),
    ...(value.value.payload
      ? { payload: structuredPayload(value.value.payload) }
      : {}),
  };
}

function structuredPayload(
  payload: StructuredPayload,
): AgentStructuredPayload {
  return {
    mediaType: payload.mediaType,
    data: new Uint8Array(payload.data),
  };
}

function unsupportedReason(reason: UnsupportedReason) {
  if (reason === UnsupportedReason.UNKNOWN) return "unknown" as const;
  if (reason === UnsupportedReason.UNSUPPORTED) return "unsupported" as const;
  if (reason === UnsupportedReason.INVALID) return "invalid" as const;
  return "unspecified" as const;
}

function videoStage(stage: VideoStage) {
  if (stage === VideoStage.QUEUED) return "queued" as const;
  if (stage === VideoStage.RENDERING) return "rendering" as const;
  if (stage === VideoStage.TRANSCODING) return "transcoding" as const;
  if (stage === VideoStage.READY) return "ready" as const;
  if (stage === VideoStage.FAILED) return "failed" as const;
  return "unknown" as const;
}

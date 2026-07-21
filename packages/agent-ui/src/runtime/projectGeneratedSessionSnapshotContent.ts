import type {
  ContentBlock,
  MediaContent,
} from "@agent-cloud/proto/agent_workbench/v2/content_pb";

import type {
  AgentActivityItem,
  AgentAttachmentItem,
  AgentArtifactItem,
  AgentToolResult,
} from "../contracts";
import {
  projectArtifactReference,
  type ArtifactCatalog,
  type ArtifactProjectionReference,
} from "./projectGeneratedSessionSnapshotArtifacts";
import {
  contentBlockData,
  contentBlockText,
} from "./projectGeneratedSessionSnapshotContentText";
import {
  toolArtifactMetadata,
  unsupportedContentEvidence,
} from "./projectGeneratedSessionSnapshotContentEvidence";
import { formatUnsupported } from "./projectGeneratedSessionSnapshotPayload";

export interface TimelineContentProjection {
  artifacts: AgentArtifactItem[];
  attachments: AgentAttachmentItem[];
  evidence: AgentActivityItem[];
  text: string[];
}

export function projectTimelineContent(
  blocks: readonly ContentBlock[],
  parentId: string,
  catalog: ArtifactCatalog,
): TimelineContentProjection {
  const projection: TimelineContentProjection = {
    artifacts: [],
    attachments: [],
    evidence: [],
    text: [],
  };
  blocks.forEach((block, index) => {
    const id = `${parentId}:${block.contentId || index}`;
    const text = contentBlockText(block);
    if (text !== undefined) projection.text.push(text);
    projectBlockArtifacts(block, id, catalog, projection);
    const evidence = unsupportedContentEvidence(block, id);
    if (evidence) projection.evidence.push(evidence);
  });
  return projection;
}

export function projectToolResultBlocks(
  blocks: readonly ContentBlock[],
  resultId: string,
  catalog: ArtifactCatalog,
): AgentToolResult[] {
  return blocks.flatMap<AgentToolResult>(
    (block, index): AgentToolResult[] => {
      const id = `${resultId}:${block.contentId || index}`;
      const artifactReferences = blockArtifactReferences(block);
      if (artifactReferences.length > 0) {
        const results = artifactReferences.map<AgentToolResult>(
          (reference, referenceIndex) => {
            const projected = projectArtifactReference(
              reference,
              referenceIndex === 0 ? id : `${id}:${referenceIndex}`,
              catalog,
              { schemaVersion: block.identity?.schemaVersion },
            );
            if (projected.kind !== "artifact") {
              return {
                id: projected.id,
                kind: "data",
                value: projected.detail,
              };
            }
            return {
              id: projected.id,
              kind: "artifact",
              artifactId: projected.artifactId,
              mediaType: projected.mimeType,
              representationId: projected.selectedRepresentationId,
              revision: projected.revision,
              role: projected.role,
              schemaVersion: projected.schemaVersion,
            };
          },
        );
        const metadata = toolArtifactMetadata(block);
        if (metadata !== undefined) {
          results.push({ id: `${id}:metadata`, kind: "data", value: metadata });
        }
        return results;
      }
      if (block.content.case === "unsupported") {
        return [
          {
            id,
            kind: "data",
            value: formatUnsupported(block.content.value),
          },
        ];
      }
      if (
        block.content.case === "json" ||
        block.content.case === "table" ||
        block.content.case === "progress" ||
        block.content.case === "livePreview" ||
        block.content.case === "restrictedIframe" ||
        block.content.case === "imageComparison" ||
        block.content.case === "annotation"
      ) {
        return [{ id, kind: "data", value: contentBlockData(block) }];
      }
      const text = contentBlockText(block);
      return text === undefined
        ? [{ id, kind: "data", value: "missing content payload" }]
        : [{ id, kind: "text", text }];
    },
  );
}

function projectBlockArtifacts(
  block: ContentBlock,
  id: string,
  catalog: ArtifactCatalog,
  projection: TimelineContentProjection,
): void {
  blockArtifactReferences(block).forEach((reference, index) => {
    if (reference.role === "input" && !catalog.has(reference.artifactId)) {
      projection.attachments.push({
        attachmentId: reference.artifactId,
        filename: reference.filename || reference.artifactId,
        id: index === 0 ? id : `${id}:${index}`,
        kind: "attachment",
        mimeType: reference.mediaType || null,
      });
      return;
    }
    const projected = projectArtifactReference(
      reference,
      index === 0 ? id : `${id}:${index}`,
      catalog,
      { schemaVersion: block.identity?.schemaVersion },
    );
    if (projected.kind === "artifact") projection.artifacts.push(projected);
    else projection.evidence.push(projected);
  });
}

function blockArtifactReferences(
  block: ContentBlock,
): ArtifactProjectionReference[] {
  const content = block.content;
  if (
    content.case === "image" ||
    content.case === "video" ||
    content.case === "audio" ||
    content.case === "pdf" ||
    content.case === "presentation" ||
    content.case === "spreadsheet" ||
    content.case === "file"
  ) {
    return [mediaReference(content.value)];
  }
  if (content.case === "imageComparison") {
    return [content.value.source, content.value.result]
      .filter((value): value is MediaContent => value !== undefined)
      .map(mediaReference);
  }
  if (content.case === "annotation" && content.value.source) {
    return [mediaReference(content.value.source)];
  }
  if (content.case === "html" && content.value.payload.case === "artifact") {
    return [mediaReference(content.value.payload.value)];
  }
  if (content.case === "artifactRef") return [content.value];
  return [];
}

function mediaReference(media: MediaContent): ArtifactProjectionReference {
  return {
    artifactId: media.artifactId,
    filename: media.filename,
    mediaType: media.mediaType,
    representationId: media.representationId,
    revision: media.revision,
    role: media.role,
  };
}

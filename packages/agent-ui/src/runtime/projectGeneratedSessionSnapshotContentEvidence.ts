import type {
  ContentBlock,
  MediaContent,
} from "@do-worker/proto/agent_workbench/v2/content_pb";

import type { AgentActivityItem } from "../contracts";
import {
  decodeStructuredPayload,
  formatUnsupported,
} from "./projectGeneratedSessionSnapshotPayload";

export function unsupportedContentEvidence(
  block: ContentBlock,
  id: string,
): AgentActivityItem | undefined {
  const detail = unsupportedContentDetail(block);
  return detail
    ? {
        id,
        kind: "system",
        title: "Unsupported content",
        detail,
        status: "failed",
      }
    : undefined;
}

export function toolArtifactMetadata(block: ContentBlock): unknown | undefined {
  const content = block.content;
  if (content.case === "imageComparison") {
    return {
      kind: content.case,
      defaultMode: content.value.defaultMode,
      source: mediaMetadata(content.value.source),
      result: mediaMetadata(content.value.result),
    };
  }
  if (content.case === "annotation") {
    return {
      kind: content.case,
      source: mediaMetadata(content.value.source),
      annotations: decodeStructuredPayload(content.value.annotations)?.value,
    };
  }
  if (content.case === "html" && content.value.payload.case === "artifact") {
    return {
      kind: content.case,
      securityProfile: content.value.securityProfile,
      artifact: mediaMetadata(content.value.payload.value),
    };
  }
  if (
    content.case === "image" ||
    content.case === "video" ||
    content.case === "audio" ||
    content.case === "pdf" ||
    content.case === "presentation" ||
    content.case === "spreadsheet" ||
    content.case === "file"
  ) {
    return mediaMetadata(content.value);
  }
  return undefined;
}

function unsupportedContentDetail(block: ContentBlock): string | undefined {
  const content = block.content;
  if (content.case === "unsupported") return formatUnsupported(content.value);
  if (content.case === "livePreview") {
    return [
      "kind=livePreview",
      `resourceId=${content.value.resourceId}`,
      `securityProfile=${content.value.securityProfile}`,
      content.value.sessionUrl ? `sessionUrl=${content.value.sessionUrl}` : undefined,
    ]
      .filter(Boolean)
      .join("; ");
  }
  if (content.case === "restrictedIframe") {
    return [
      "kind=restrictedIframe",
      `rendererId=${content.value.rendererId}`,
      `protocolVersion=${content.value.protocolVersion}`,
      `url=${content.value.url}`,
      `origin=${content.value.origin}`,
      `sandboxTokens=${content.value.sandboxTokens.join(",")}`,
      `permissions=${content.value.permissions.join(",")}`,
      `maxInboundBytes=${content.value.maxInboundBytes.toString()}`,
      `maxOutboundBytes=${content.value.maxOutboundBytes.toString()}`,
    ].join("; ");
  }
  if (content.case === "imageComparison") {
    return `kind=imageComparison; defaultMode=${content.value.defaultMode ?? "unspecified"}`;
  }
  if (content.case === "annotation") {
    return `kind=annotation; annotations=${decodeStructuredPayload(content.value.annotations)?.text ?? "missing"}`;
  }
  if (content.case === "html" && content.value.payload.case === "artifact") {
    return `kind=html; securityProfile=${content.value.securityProfile}`;
  }
  if (content.case === undefined) return "missing content payload";
  return undefined;
}

function mediaMetadata(media: MediaContent | undefined): unknown {
  if (!media) return undefined;
  return {
    artifactId: media.artifactId,
    representationId: media.representationId,
    revision: media.revision.toString(),
    mediaType: media.mediaType,
    role: media.role,
    filename: media.filename,
    altText: media.altText,
  };
}

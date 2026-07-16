import type {
  ArtifactDescriptor,
  ArtifactRepresentation,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";

import type { AgentArtifactLoadRequest } from "./AgentSessionRuntimeV2";

export interface AgentArtifactTransportResources {
  getArtifacts(sessionId: string): readonly ArtifactDescriptor[];
  loadDownload(
    url: string,
    context: AgentArtifactTransportContext,
  ): Promise<Blob>;
  loadResource(
    resourceId: string,
    context: AgentArtifactTransportContext,
  ): Promise<Blob>;
}

export interface AgentArtifactTransportContext
  extends AgentArtifactLoadRequest {
  representation: ArtifactRepresentation;
}

export function createAgentArtifactLoader(
  resources: AgentArtifactTransportResources,
): (request: AgentArtifactLoadRequest) => Promise<Blob> {
  return async ({ artifactId, representationId, sessionId }) => {
    if (!representationId) {
      throw new Error("artifact_representation_missing");
    }
    const descriptor = resources
      .getArtifacts(sessionId)
      .find((artifact) => artifact.artifactId === artifactId);
    if (!descriptor) throw new Error("artifact_descriptor_missing");
    const representation = descriptor.representations.find(
      (candidate) => candidate.representationId === representationId,
    );
    if (!representation) {
      throw new Error("artifact_representation_missing");
    }
    return loadArtifactRepresentation(resources, {
      artifactId,
      representation,
      representationId,
      sessionId,
    });
  };
}

async function loadArtifactRepresentation(
  resources: AgentArtifactTransportResources,
  context: AgentArtifactTransportContext,
): Promise<Blob> {
  const { representation } = context;
  const transport = representation.transport?.transport;
  if (!transport?.case) throw new Error("artifact_transport_missing");
  if (transport.case === "inlineBytes") {
    return new Blob([transport.value.slice()], {
      type: representation.mediaType,
    });
  }
  if (transport.case === "inlineText") {
    return new Blob([transport.value], { type: representation.mediaType });
  }
  if (transport.case === "resourceId") {
    return resources.loadResource(transport.value, context);
  }
  return resources.loadDownload(transport.value, context);
}

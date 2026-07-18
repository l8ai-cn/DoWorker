import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "./contracts";
import type { AgentContentRendererRegistration } from "./react/contentRendererTypes";
import type { ContentRendererRegistry } from "./registry/ContentRendererRegistry";
import {
  artifactRendererKey,
  artifactUnsupportedReason,
} from "./registry/artifactRendererKey";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import { ArtifactError, GenericArtifactCard } from "./GenericArtifactCard";

export interface ArtifactCardProps {
  contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
  item: AgentArtifactItem;
  presentation?: "developer" | "user";
  runtime: AgentSessionRuntime;
  sessionId: string;
}

export function ArtifactCard({
  contentRenderers,
  item,
  presentation = "developer",
  runtime,
  sessionId,
}: ArtifactCardProps) {
  const text = useAgentWorkspaceText();
  const filename = item.filename.trim() || text.artifact.generatedArtifact;
  const unsupportedReason = artifactUnsupportedReason(item);
  if (unsupportedReason) {
    return (
      <ArtifactError
        filename={filename}
        message={
          presentation === "user"
            ? text.artifact.loadFailed
            : unsupportedReason
        }
      />
    );
  }

  const key = artifactRendererKey(item);
  const RegisteredViewer = key ? contentRenderers?.lookup(key)?.viewer : undefined;
  if (RegisteredViewer) {
    return (
      <RegisteredViewer
        filename={filename}
        item={item}
        presentation={presentation}
        runtime={runtime}
        sessionId={sessionId}
      />
    );
  }
  return (
    <GenericArtifactCard
      filename={filename}
      item={item}
      runtime={runtime}
      sessionId={sessionId}
    />
  );
}

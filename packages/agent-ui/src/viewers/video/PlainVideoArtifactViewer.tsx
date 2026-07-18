import type { AgentContentRendererProps } from "../../react/contentRendererTypes";
import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";
import { useArtifactBlobUrl } from "../../useArtifactBlobUrl";
import {
  ArtifactViewerError,
  ArtifactViewerLoading,
} from "../ArtifactViewerStatus";
import { VideoArtifactViewer } from "./VideoArtifactViewer";

export function PlainVideoArtifactViewer({
  filename,
  item,
  presentation = "developer",
  runtime,
  sessionId,
}: AgentContentRendererProps) {
  const text = useAgentWorkspaceText().artifact;
  const state = useArtifactBlobUrl(item, runtime, sessionId);
  if (state.status === "idle" || state.status === "loading") {
    return <ArtifactViewerLoading filename={filename} />;
  }
  if (state.status === "error") {
    return (
      <ArtifactViewerError
        filename={filename}
        message={presentation === "user" ? text.loadFailed : state.message}
      />
    );
  }
  if (!state.mimeType) {
    return (
      <ArtifactViewerError
        filename={filename}
        message={
          presentation === "user"
            ? text.loadFailed
            : "video_artifact_media_type_missing"
        }
      />
    );
  }
  return (
    <VideoArtifactViewer
      filename={filename}
      mimeType={state.mimeType}
      onDownload={() => download(state.url, filename)}
      src={state.url}
      status="ready"
      technicalMetadata={presentation === "developer"}
    />
  );
}

function download(url: string, filename: string) {
  const link = document.createElement("a");
  link.download = filename;
  link.href = url;
  link.click();
}

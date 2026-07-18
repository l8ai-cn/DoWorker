import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";

export function ArtifactAudioPreview({
  filename,
  src,
}: {
  filename: string;
  src: string;
}) {
  const text = useAgentWorkspaceText().artifact;
  return (
    <div className="border-b border-border bg-muted/30 p-4">
      <audio
        aria-label={text.audioPreview(filename)}
        className="w-full"
        controls
        preload="metadata"
        src={src}
      />
    </div>
  );
}

import { MarkdownMessage } from "../../MarkdownMessage";
import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";
import { MAX_TEXT_PREVIEW_BYTES } from "../../useArtifactBlobUrl";

export function ArtifactMarkdownPreview({
  text,
  truncated,
}: {
  text: string;
  truncated: boolean;
}) {
  const labels = useAgentWorkspaceText().artifact;
  return (
    <div className="max-h-[32rem] overflow-auto border-b border-border p-4">
      <MarkdownMessage text={text} />
      {truncated && (
        <div className="mt-4 border-t border-border pt-3 text-xs text-muted-foreground">
          {labels.previewLimited(MAX_TEXT_PREVIEW_BYTES)}
        </div>
      )}
    </div>
  );
}

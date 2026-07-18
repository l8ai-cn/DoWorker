import { FileText, LoaderCircle, Paperclip, X } from "lucide-react";
import { useRef, useState, type ChangeEvent } from "react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type {
  AgentAttachmentReference,
  AgentSessionRuntime,
} from "./contracts";

export function ComposerAttachments({
  attachments,
  disabled,
  onChange,
  runtime,
  sessionId,
}: {
  attachments: AgentAttachmentReference[];
  disabled: boolean;
  onChange: (attachments: AgentAttachmentReference[]) => void;
  runtime: AgentSessionRuntime;
  sessionId: string;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const text = useAgentWorkspaceText();

  const handleFile = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || !runtime.uploadAttachment || uploading) return;
    setUploadError(null);
    setUploading(true);
    try {
      const attachment = await runtime.uploadAttachment(sessionId, file);
      onChange([...attachments, attachment]);
    } catch (cause) {
      setUploadError(attachmentUploadError(cause, text));
    } finally {
      setUploading(false);
    }
  };

  if (!runtime.uploadAttachment) return null;

  return (
    <div className="flex min-w-0 flex-1 flex-wrap items-center gap-1.5">
      <button
        aria-label={text.addAttachment}
        className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground outline-none hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-40"
        disabled={disabled || uploading}
        onClick={() => inputRef.current?.click()}
        title={text.addAttachment}
        type="button"
      >
        {uploading ? (
          <LoaderCircle className="size-4 animate-spin" />
        ) : (
          <Paperclip className="size-4" />
        )}
      </button>
      <input
        className="hidden"
        data-testid="agent-attachment-input"
        onChange={handleFile}
        ref={inputRef}
        type="file"
      />
      {attachments.map((attachment) => (
        <span
          className="flex h-8 max-w-56 items-center gap-1.5 rounded-md border border-border bg-muted/50 px-2 text-xs"
          key={attachment.id}
        >
          <FileText className="size-3.5 shrink-0 text-muted-foreground" />
          <span className="truncate">{attachment.name}</span>
          <button
            aria-label={`${text.removeAttachment}: ${attachment.name}`}
            className="shrink-0 text-muted-foreground hover:text-foreground"
            onClick={() =>
              onChange(attachments.filter((item) => item.id !== attachment.id))
            }
            type="button"
          >
            <X className="size-3.5" />
          </button>
        </span>
      ))}
      {uploadError && (
        <span className="basis-full text-xs text-destructive" role="alert">
          {uploadError}
        </span>
      )}
    </div>
  );
}

function attachmentUploadError(
  cause: unknown,
  text: ReturnType<typeof useAgentWorkspaceText>,
) {
  return cause instanceof Error &&
    cause.message === "agent_attachment_upload_unsupported"
    ? text.attachmentUploadUnsupported
    : text.attachmentUploadFailed;
}

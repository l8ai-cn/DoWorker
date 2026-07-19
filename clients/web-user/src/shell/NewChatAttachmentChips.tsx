import { FileTextIcon, FolderIcon, ImageIcon, XIcon } from "lucide-react";
import type { MentionItem } from "@/lib/composerMentions";
import { mentionItemPath } from "@/lib/composerMentions";

export function NewChatAttachmentChips({
  mentionedItems,
  files,
  onRemoveMention,
  onRemoveFile,
}: {
  mentionedItems: MentionItem[];
  files: File[];
  onRemoveMention: (index: number) => void;
  onRemoveFile: (index: number) => void;
}) {
  return (
    <>
      {mentionedItems.length > 0 && (
        <div className="flex flex-wrap gap-1.5 px-4 pb-2">
          {mentionedItems.map((item, index) => (
            <span
              key={mentionItemPath(item)}
              className="flex items-center gap-1 rounded-full border border-border bg-muted px-2 py-0.5 text-xs text-muted-foreground"
            >
              {item.isDir ? (
                <FolderIcon className="size-3 shrink-0" />
              ) : (
                <FileTextIcon className="size-3 shrink-0" />
              )}
              <span className="max-w-[200px] truncate" title={mentionItemPath(item)}>
                @{item.path}
                {item.isDir ? "/" : ""}
              </span>
              <button
                type="button"
                onClick={() => onRemoveMention(index)}
                className="ml-0.5 rounded-full hover:text-foreground"
                aria-label={`Remove ${item.path}`}
              >
                <XIcon className="size-3" />
              </button>
            </span>
          ))}
        </div>
      )}
      {files.length > 0 && (
        <div className="flex flex-wrap gap-1.5 px-4 pb-2">
          {files.map((file, index) => (
            <span
              key={index}
              className="flex items-center gap-1 rounded-full border border-border bg-muted px-2 py-0.5 text-xs text-muted-foreground"
            >
              {file.type.startsWith("image/") ? (
                <ImageIcon className="size-3 shrink-0" />
              ) : (
                <FileTextIcon className="size-3 shrink-0" />
              )}
              <span className="max-w-[140px] truncate">{file.name || "image.png"}</span>
              <button
                type="button"
                onClick={() => onRemoveFile(index)}
                className="ml-0.5 rounded-full hover:text-foreground"
                aria-label={`Remove ${file.name || "image.png"}`}
              >
                <XIcon className="size-3" />
              </button>
            </span>
          ))}
        </div>
      )}
    </>
  );
}

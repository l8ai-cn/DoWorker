import { type Dispatch, type RefObject, type SetStateAction, useMemo, useState } from "react";
import { useHostFilesystem } from "@/hooks/useHostFilesystem";
import { useMentionBrowser } from "@/hooks/useMentionBrowser";
import type { WorkspaceFile } from "@/hooks/useWorkspaceChangedFiles";
import { type MentionState, parseMentionToken, rankMentionEntries } from "@/lib/composerMentions";

export function useNewChatMentionState({
  isNativeTerminalAgent,
  sandboxSelected,
  selectedHostId,
  workspaceValid,
  workspaceTrimmed,
  message,
  setMessage,
  textareaRef,
}: {
  isNativeTerminalAgent: boolean;
  sandboxSelected: boolean;
  selectedHostId: string | null;
  workspaceValid: boolean;
  workspaceTrimmed: string;
  message: string;
  setMessage: Dispatch<SetStateAction<string>>;
  textareaRef: RefObject<HTMLTextAreaElement | null>;
}) {
  const [mention, setMention] = useState<MentionState | null>(null);
  const mentionEnabled =
    isNativeTerminalAgent && !sandboxSelected && selectedHostId !== null && workspaceValid;
  const { dir: mentionDir, filter: mentionFilter } = parseMentionToken(mention?.query ?? "");
  const workspaceRoot = workspaceTrimmed.replace(/\/+$/, "");
  const mentionAbsDir =
    mentionEnabled && mention
      ? mentionDir
        ? `${workspaceRoot}/${mentionDir}`
        : workspaceRoot
      : null;
  const mentionFsQuery = useHostFilesystem(
    mentionEnabled && mention ? selectedHostId : null,
    mentionAbsDir,
  );
  const mentionEntries: WorkspaceFile[] = useMemo(() => {
    if (!mentionEnabled || !mention || mentionFsQuery.isPlaceholderData) return [];
    const rows = (mentionFsQuery.data?.entries ?? [])
      .filter((entry) => entry.type === "directory" || entry.type === "file")
      .map(
        (entry): WorkspaceFile => ({
          path: entry.path.startsWith(workspaceRoot)
            ? entry.path.slice(workspaceRoot.length).replace(/^\/+/, "")
            : entry.name,
          name: entry.name,
          type: entry.type === "directory" ? "directory" : "file",
          bytes: entry.bytes,
          modified_at: entry.modified_at,
        }),
      );
    return rankMentionEntries(rows, mentionFilter);
  }, [
    mentionEnabled,
    mention,
    mentionFsQuery.data,
    mentionFsQuery.isPlaceholderData,
    mentionFilter,
    workspaceRoot,
  ]);
  const mentionListingPending =
    mentionEnabled &&
    mention !== null &&
    (mentionFsQuery.isLoading || mentionFsQuery.isPlaceholderData);
  const mentionBrowser = useMentionBrowser({
    mention,
    setMention,
    mentionEntries,
    text: message,
    setText: setMessage,
    textareaRef,
  });

  return {
    mentionEnabled,
    mentionDir,
    mentionEntries,
    mentionOpen: mentionEntries.length > 0,
    mentionListingPending,
    setMention,
    ...mentionBrowser,
  };
}

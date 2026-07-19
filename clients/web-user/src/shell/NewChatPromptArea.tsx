import type { Dispatch, KeyboardEvent, MutableRefObject, RefObject, SetStateAction } from "react";
import { FileMentionMenu } from "@/components/FileMentionMenu";
import { SkillPills } from "@/components/SkillPills";
import { SlashCommandMenu } from "@/components/SlashCommandMenu";
import { isImeCompositionKeyEvent } from "@/lib/ime";
import { detectMentionAt, type MentionState } from "@/lib/composerMentions";
import type { WorkspaceFile } from "@/hooks/useWorkspaceChangedFiles";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";

export function NewChatPromptArea({
  textareaRef,
  message,
  setMessage,
  mentionEnabled,
  setMention,
  dismissMention,
  isComposingRef,
  handleMentionKeyDown,
  mentionOpen,
  mentionListingPending,
  mentionDir,
  mentionIndex,
  mentionEntries,
  openMentionDir,
  attachMention,
  slashMenuOpen,
  slashMenuQuery,
  slashMenuIndex,
  slashMenuMatches,
  setSlashMenuIndex,
  applySlashSelection,
  skillCommands,
  pillSkills,
  applySkillPill,
  onSubmit,
  onFiles,
  placeholder,
  placeholderSkills,
}: {
  textareaRef: RefObject<HTMLTextAreaElement | null>;
  message: string;
  setMessage: Dispatch<SetStateAction<string>>;
  mentionEnabled: boolean;
  setMention: (next: MentionState | null) => void;
  dismissMention: () => void;
  isComposingRef: MutableRefObject<boolean>;
  handleMentionKeyDown: (event: KeyboardEvent<HTMLTextAreaElement>) => boolean;
  mentionOpen: boolean;
  mentionListingPending: boolean;
  mentionDir: string;
  mentionIndex: number;
  mentionEntries: WorkspaceFile[];
  openMentionDir: (path: string) => void;
  attachMention: (path: string, isDir: boolean) => void;
  slashMenuOpen: boolean;
  slashMenuQuery: string;
  slashMenuIndex: number;
  slashMenuMatches: string[];
  setSlashMenuIndex: Dispatch<SetStateAction<number>>;
  applySlashSelection: (cmd: string) => void;
  skillCommands: Record<string, string>;
  pillSkills: AvailableAgent["skills"];
  applySkillPill: (name: string) => void;
  onSubmit: () => void;
  onFiles: (files: File[]) => void;
  placeholder: string;
  placeholderSkills: string;
}) {
  const onKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (isImeCompositionKeyEvent(event, isComposingRef.current)) return;
    if (handleMentionKeyDown(event)) return;
    if (slashMenuOpen && slashMenuMatches.length > 0) {
      if (event.key === "ArrowDown") {
        event.preventDefault();
        setSlashMenuIndex((index) => (index + 1) % slashMenuMatches.length);
        return;
      }
      if (event.key === "ArrowUp") {
        event.preventDefault();
        setSlashMenuIndex((index) => (index <= 0 ? slashMenuMatches.length - 1 : index - 1));
        return;
      }
      if ((event.key === "Tab" || (event.key === "Enter" && !event.shiftKey)) && slashMenuIndex >= 0) {
        event.preventDefault();
        applySlashSelection(slashMenuMatches[slashMenuIndex]!);
        return;
      }
      if (event.key === "Escape") {
        event.preventDefault();
        setMessage("");
        setSlashMenuIndex(-1);
        return;
      }
    }
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      if (!mentionListingPending) onSubmit();
    }
  };

  return (
    <>
      {slashMenuOpen && (
        <SlashCommandMenu
          query={slashMenuQuery}
          activeIndex={slashMenuIndex}
          onSelect={applySlashSelection}
          commands={skillCommands}
        />
      )}
      {(mentionOpen || mentionListingPending) && (
        <FileMentionMenu
          currentDir={mentionDir}
          activeIndex={mentionIndex}
          entries={mentionEntries}
          loading={mentionListingPending}
          onOpenDir={openMentionDir}
          onAttach={attachMention}
        />
      )}
      <textarea
        ref={textareaRef}
        value={message}
        onChange={(event) => {
          setMessage(event.target.value);
          setMention(
            mentionEnabled
              ? detectMentionAt(event.target.value, event.target.selectionStart ?? event.target.value.length)
              : null,
          );
        }}
        onBlur={dismissMention}
        onCompositionStart={() => {
          isComposingRef.current = true;
        }}
        onCompositionEnd={() => {
          isComposingRef.current = false;
        }}
        onKeyDown={onKeyDown}
        onPaste={(event) => {
          const pasted = Array.from(event.clipboardData.items)
            .filter((item) => item.kind === "file")
            .map((item) => item.getAsFile())
            .filter((file): file is File => file !== null);
          if (pasted.length > 0) {
            event.preventDefault();
            onFiles(pasted);
          }
        }}
        placeholder={pillSkills.length > 0 ? "" : placeholder}
        aria-label={placeholder}
        rows={1}
        autoFocus
        data-testid="new-chat-landing-input"
        className="max-h-[200px] min-h-[60px] w-full resize-none overflow-y-auto bg-transparent px-4 pt-4 pb-1 font-['SF_Pro_Text',-apple-system,BlinkMacSystemFont,system-ui,sans-serif] text-sm leading-5 text-foreground outline-none placeholder:text-muted-foreground md:select-text"
      />
      {pillSkills.length > 0 && message.length === 0 && (
        <div className="pointer-events-none absolute inset-x-4 top-4 flex flex-wrap items-center gap-2">
          <span className="font-['SF_Pro_Text',-apple-system,BlinkMacSystemFont,system-ui,sans-serif] text-sm leading-5 text-muted-foreground">
            {placeholderSkills}
          </span>
          <SkillPills skills={pillSkills} onPick={applySkillPill} />
        </div>
      )}
    </>
  );
}

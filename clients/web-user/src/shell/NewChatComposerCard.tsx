import { cn } from "@/lib/utils";
import { NewChatAttachmentChips } from "./NewChatAttachmentChips";
import { NewChatComposerFooter } from "./NewChatComposerFooter";
import { NewChatPromptArea } from "./NewChatPromptArea";
import type { NewChatLandingController } from "./useNewChatLandingController";

export function NewChatComposerCard({ state }: { state: NewChatLandingController }) {
  const { agent, files, mention, message, placeholder, placeholderSkills, setMessage, slash, submit } = state;
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        void submit.handleCreate();
      }}
      onDrop={files.handleDrop}
      onDragOver={files.handleDragOver}
      onDragEnter={files.handleDragEnter}
      onDragLeave={files.handleDragLeave}
      className={cn(
        "relative z-10 flex w-full flex-col rounded-2xl border border-border bg-card dark:bg-card-solid shadow-[0_12px_20px_-20px_rgba(15,118,110,0.12),0_20px_28px_-28px_rgba(15,118,110,0.08)] transition-[border-color,box-shadow] duration-150 has-[textarea:focus]:border-primary",
        files.isDragActive && "ring-2 ring-ring ring-inset",
      )}
      data-testid="new-chat-landing-composer"
    >
      {files.isDragActive && (
        <div className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-card/80">
          <span className="text-sm font-medium text-ring">Drop files here</span>
        </div>
      )}
      <NewChatPromptArea
        textareaRef={state.textareaRef}
        message={message}
        setMessage={setMessage}
        mentionEnabled={mention.mentionEnabled}
        setMention={mention.setMention}
        dismissMention={mention.dismiss}
        isComposingRef={state.isComposingRef}
        handleMentionKeyDown={mention.handleKeyDown}
        mentionOpen={mention.mentionOpen}
        mentionListingPending={mention.mentionListingPending}
        mentionDir={mention.mentionDir}
        mentionIndex={mention.mentionIndex}
        mentionEntries={mention.mentionEntries}
        openMentionDir={mention.openMentionDir}
        attachMention={mention.attachMention}
        slashMenuOpen={slash.slashMenuOpen}
        slashMenuQuery={slash.slashMenuQuery}
        slashMenuIndex={slash.slashMenuIndex}
        slashMenuMatches={slash.slashMenuMatches}
        setSlashMenuIndex={slash.setSlashMenuIndex}
        applySlashSelection={slash.applySlashSelection}
        skillCommands={slash.skillCommands}
        pillSkills={slash.pillSkills}
        applySkillPill={slash.applySkillPill}
        onSubmit={() => void submit.handleCreate()}
        onFiles={files.addFiles}
        placeholder={placeholder}
        placeholderSkills={placeholderSkills}
      />
      <input
        ref={files.fileInputRef}
        type="file"
        multiple
        accept="image/*,application/pdf,text/*,application/json,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        className="hidden"
        data-testid="new-chat-landing-file-input"
        onChange={(event) => {
          if (event.target.files) {
            files.addFiles(Array.from(event.target.files));
            event.target.value = "";
          }
        }}
      />
      <NewChatAttachmentChips
        mentionedItems={mention.mentionedItems}
        files={files.files}
        onRemoveMention={mention.removeMentionedItem}
        onRemoveFile={files.removeFile}
      />
      {files.attachmentError && (
        <p className="px-4 pb-2 text-xs text-destructive" data-testid="new-chat-landing-attachment-error">
          {files.attachmentError}
        </p>
      )}
      <NewChatComposerFooter
        fileInputRef={files.fileInputRef}
        creating={submit.creating}
        setMessage={setMessage}
        agentEntries={agent.agentEntries}
        harnessEntries={agent.harnessEntries}
        effectiveAgentId={agent.effectiveAgentId}
        agentLabel={state.labels.agent}
        hasAgents={agent.agentList.length > 0}
        host={state.location.sandboxSelected ? undefined : state.location.selectedHost}
        onSelectAgent={agent.selectAgent}
        showWorkerModelPicker={submit.showWorkerModelPicker}
        modelConfigId={submit.modelConfigId}
        setModelConfigId={submit.setModelConfigId}
        workerTokenBudget={submit.workerTokenBudget}
        setWorkerTokenBudget={submit.setWorkerTokenBudget}
        canSubmit={submit.canSubmit}
        submitDisabledReason={submit.submitDisabledReason}
      />
    </form>
  );
}

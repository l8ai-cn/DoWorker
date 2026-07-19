import { ArrowUpIcon, PaperclipIcon } from "lucide-react";
import type { Dispatch, RefObject, SetStateAction } from "react";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { ComposerMicButton } from "@/components/ComposerMicButton";
import { ModelConfigPicker } from "@/shell/ModelConfigPicker";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import type { Host } from "@/hooks/useHosts";
import { NewChatAgentPicker } from "./NewChatAgentPicker";

export function NewChatComposerFooter({
  fileInputRef,
  creating,
  setMessage,
  agentEntries,
  harnessEntries,
  effectiveAgentId,
  agentLabel,
  hasAgents,
  host,
  onSelectAgent,
  showWorkerModelPicker,
  modelConfigId,
  setModelConfigId,
  workerTokenBudget,
  setWorkerTokenBudget,
  canSubmit,
  submitDisabledReason,
}: {
  fileInputRef: RefObject<HTMLInputElement | null>;
  creating: boolean;
  setMessage: Dispatch<SetStateAction<string>>;
  agentEntries: AvailableAgent[];
  harnessEntries: AvailableAgent[];
  effectiveAgentId: string | null;
  agentLabel: string;
  hasAgents: boolean;
  host: Host | undefined | null;
  onSelectAgent: (agent: AvailableAgent) => void;
  showWorkerModelPicker: boolean;
  modelConfigId: number | null;
  setModelConfigId: (id: number | null) => void;
  workerTokenBudget: number | null;
  setWorkerTokenBudget: (value: number | null) => void;
  canSubmit: boolean;
  submitDisabledReason: string | null;
}) {
  return (
    <div className="flex items-center justify-between pt-1 pr-4 pb-3 pl-2">
      <div className="flex items-center gap-0.5">
        <Button
          type="button"
          size="icon"
          variant="ghost"
          className="size-9 md:size-8"
          disabled={creating}
          onClick={() => fileInputRef.current?.click()}
          title="Attach files"
          data-testid="new-chat-landing-attach"
        >
          <PaperclipIcon className="size-4" />
          <span className="sr-only">Attach files</span>
        </Button>
        <ComposerMicButton
          disabled={creating}
          onTranscript={(text) => setMessage((prev) => (prev ? `${prev} ${text}` : text))}
        />
      </div>
      <div className="flex items-center gap-0.5">
        <NewChatAgentPicker
          agentEntries={agentEntries}
          harnessEntries={harnessEntries}
          effectiveAgentId={effectiveAgentId}
          agentLabel={agentLabel}
          hasAgents={hasAgents}
          host={host}
          onSelectAgent={onSelectAgent}
        />
        {showWorkerModelPicker && (
          <ModelConfigPicker
            selectedId={modelConfigId}
            onSelect={setModelConfigId}
            tokenBudget={workerTokenBudget}
            onTokenBudgetChange={setWorkerTokenBudget}
            disabled={creating}
          />
        )}
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex">
                <Button
                  type="submit"
                  size="icon"
                  disabled={!canSubmit}
                  aria-label="Start session"
                  data-testid="new-chat-landing-submit"
                  className="size-8 rounded-full bg-primary text-primary-foreground transition-opacity hover:bg-primary/90 disabled:opacity-50"
                >
                  <ArrowUpIcon className="size-4" />
                </Button>
              </span>
            </TooltipTrigger>
            {submitDisabledReason != null && <TooltipContent>{submitDisabledReason}</TooltipContent>}
          </Tooltip>
        </TooltipProvider>
      </div>
    </div>
  );
}

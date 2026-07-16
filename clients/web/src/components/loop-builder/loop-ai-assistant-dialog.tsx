"use client";

import { AlertMessage } from "@/components/ui/alert-message";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
} from "@/components/ui/dialog";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import type {
  LoopAIMode,
  LoopAIProposal,
  LoopAIRepairTarget,
} from "./loop-ai-assistant-types";
import { LoopAIModeTabs } from "./loop-ai-mode-tabs";
import { LoopAIProposalPreview } from "./loop-ai-proposal-preview";
import { LoopAIRepairForm } from "./loop-ai-repair-form";
import { LoopAIRequestForm, type LoopAIRequestFormProps } from "./loop-ai-request-form";
import { LoopProgramExplanation } from "./loop-program-explanation";

interface LoopAIAssistantDialogProps extends LoopAIRequestFormProps {
  open: boolean;
  mode: LoopAIMode;
  parseStatus: string;
  program?: LoopProgram;
  proposal?: LoopAIProposal;
  repairTarget?: LoopAIRepairTarget;
  onBack: () => void;
  onConfirm: () => void;
}

export function LoopAIAssistantDialog(props: LoopAIAssistantDialogProps) {
  const {
    mode,
    open,
    proposal,
    repairTarget,
    messages,
    onOpenChange,
  } = props;
  const title = repairTarget
    ? messages.repair.title
    : mode === "generate"
      ? messages.generateTitle
      : messages.explainTitle;
  const description = repairTarget
    ? messages.repair.description
    : mode === "generate"
      ? messages.generateDescription
      : messages.explainDescription;
  const resultClassName = proposal
    ? "max-w-5xl"
    : mode === "explain"
      ? "max-w-3xl"
      : "max-w-lg";

  return (
    <Dialog open={open} onOpenChange={onOpenChange} overlayClassName="z-[100001]">
      <DialogContent
        className={resultClassName}
        title={title}
        description={description}
      >
        {proposal ? (
          <ProposalReview {...props} proposal={proposal} />
        ) : repairTarget ? (
          <LoopAIRepairForm {...props} target={repairTarget} />
        ) : mode === "explain" ? (
          <ExplainProjection {...props} />
        ) : (
          <LoopAIRequestForm {...props} />
        )}
      </DialogContent>
    </Dialog>
  );
}

function ExplainProjection({
  busy,
  messages,
  parseStatus,
  program,
  onModeChange,
  onOpenChange,
}: LoopAIAssistantDialogProps) {
  return (
    <>
      <DialogBody className="space-y-4">
        <LoopAIModeTabs
          busy={busy}
          messages={messages}
          mode="explain"
          onModeChange={onModeChange}
        />
        <LoopProgramExplanation
          messages={messages.projection}
          program={program}
          valid={parseStatus === "valid"}
        />
      </DialogBody>
      <DialogFooter>
        <Button onClick={() => onOpenChange(false)}>{messages.close}</Button>
      </DialogFooter>
    </>
  );
}

function ProposalReview({
  proposal,
  busy,
  requestError,
  messages,
  onBack,
  onConfirm,
}: LoopAIAssistantDialogProps & { proposal: LoopAIProposal }) {
  return (
    <>
      <DialogBody>
        {proposal.repair && (
          <div className="mb-4 rounded-md border border-border bg-surface-muted/40 p-3">
            <p className="text-xs font-semibold text-muted-foreground">
              {messages.repair.patch}
            </p>
            <p className="mt-1 text-sm font-medium">
              {messages.repair.patchValue(
                proposal.repair.oldValue.toString(),
                proposal.repair.newValue.toString(),
              )}
            </p>
            <code className="mt-1 block break-all text-xs text-muted-foreground">
              {proposal.repair.fieldPath}
            </code>
          </div>
        )}
        <LoopAIProposalPreview
          currentSource={proposal.currentSource}
          messages={messages}
          proposedSource={proposal.proposedSource}
        />
        {requestError && (
          <AlertMessage className="mt-4" message={requestError} type="error" />
        )}
      </DialogBody>
      <DialogFooter>
        <Button disabled={busy} onClick={onBack} variant="outline">
          {messages.back}
        </Button>
        <Button disabled={busy} loading={busy} onClick={onConfirm}>
          {messages.confirm}
        </Button>
      </DialogFooter>
    </>
  );
}

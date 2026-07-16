"use client";

import { AlertMessage } from "@/components/ui/alert-message";
import { Button } from "@/components/ui/button";
import { DialogBody, DialogFooter } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import type {
  LoopAIRepairTarget,
  LoopAIResource,
} from "./loop-ai-assistant-types";
import type { LoopAIMessages } from "./loop-workbench-messages";

interface LoopAIRepairFormProps {
  target: LoopAIRepairTarget;
  prompt: string;
  selectedResourceId: string;
  resources: LoopAIResource[];
  resourcesLoading: boolean;
  resourceError?: string;
  busy: boolean;
  requestError?: string;
  messages: LoopAIMessages;
  onOpenChange: (open: boolean) => void;
  onPromptChange: (prompt: string) => void;
  onResourceChange: (resourceId: string) => void;
  onRetryResources: () => void;
  onSubmit: () => void;
}

export function LoopAIRepairForm({
  target,
  prompt,
  selectedResourceId,
  resources,
  resourcesLoading,
  resourceError,
  busy,
  requestError,
  messages,
  onOpenChange,
  onPromptChange,
  onResourceChange,
  onRetryResources,
  onSubmit,
}: LoopAIRepairFormProps) {
  const canSubmit =
    resources.some((resource) => resource.id === selectedResourceId) &&
    !resourcesLoading &&
    !resourceError &&
    !busy;

  return (
    <>
      <DialogBody className="space-y-4">
        <dl className="grid gap-3 rounded-md border border-border bg-surface-muted/40 p-3 text-sm">
          <TargetRow label={messages.repair.diagnostic} value={target.diagnosticLabel} />
          <TargetRow code label={messages.repair.field} value={target.fieldPath} />
        </dl>
        <div className="space-y-2">
          <Label>{messages.resource}</Label>
          <Select
            disabled={resourcesLoading || resources.length === 0 || busy}
            value={selectedResourceId}
            onValueChange={onResourceChange}
          >
            <SelectTrigger>
              <SelectValue placeholder={messages.resourcePlaceholder} />
            </SelectTrigger>
            <SelectContent>
              {resources.map((resource) => (
                <SelectItem key={resource.id} value={resource.id}>
                  {resource.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {resourcesLoading && <StateText>{messages.loadingResources}</StateText>}
          {!resourcesLoading && resourceError && (
            <div className="flex items-center justify-between gap-3">
              <StateText destructive>{resourceError}</StateText>
              <Button onClick={onRetryResources} size="sm" variant="outline">
                {messages.retry}
              </Button>
            </div>
          )}
          {!resourcesLoading && !resourceError && resources.length === 0 && (
            <StateText destructive>{messages.noResources}</StateText>
          )}
        </div>
        <div className="space-y-2">
          <Label htmlFor="loop-ai-repair-prompt">{messages.repair.prompt}</Label>
          <Textarea
            id="loop-ai-repair-prompt"
            disabled={busy}
            rows={3}
            placeholder={messages.repair.promptPlaceholder}
            value={prompt}
            onChange={(event) => onPromptChange(event.target.value)}
          />
        </div>
        {requestError && <AlertMessage message={requestError} type="error" />}
      </DialogBody>
      <DialogFooter>
        <Button disabled={busy} onClick={() => onOpenChange(false)} variant="outline">
          {messages.cancel}
        </Button>
        <Button disabled={!canSubmit} loading={busy} onClick={onSubmit}>
          {busy ? messages.repair.repairing : messages.repair.repair}
        </Button>
      </DialogFooter>
    </>
  );
}

function TargetRow({
  label,
  value,
  code = false,
}: {
  label: string;
  value: string;
  code?: boolean;
}) {
  return (
    <div className="grid gap-1 sm:grid-cols-[7rem_minmax(0,1fr)]">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className={code ? "break-all font-mono text-xs" : "text-foreground"}>
        {value}
      </dd>
    </div>
  );
}

function StateText({
  children,
  destructive = false,
}: {
  children: React.ReactNode;
  destructive?: boolean;
}) {
  return (
    <p className={destructive ? "text-sm text-destructive" : "text-sm text-muted-foreground"}>
      {children}
    </p>
  );
}

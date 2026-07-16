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
import type { LoopAIMode, LoopAIResource } from "./loop-ai-assistant-types";
import { LoopAIModeTabs } from "./loop-ai-mode-tabs";
import type { LoopAIMessages } from "./loop-workbench-messages";

export interface LoopAIRequestFormProps {
  prompt: string;
  selectedResourceId: string;
  resources: LoopAIResource[];
  resourcesLoading: boolean;
  resourceError?: string;
  busy: boolean;
  requestError?: string;
  messages: LoopAIMessages;
  onOpenChange: (open: boolean) => void;
  onModeChange: (mode: LoopAIMode) => void;
  onPromptChange: (prompt: string) => void;
  onResourceChange: (resourceId: string) => void;
  onRetryResources: () => void;
  onSubmit: () => void;
}

export function LoopAIRequestForm({
  prompt,
  selectedResourceId,
  resources,
  resourcesLoading,
  resourceError,
  busy,
  requestError,
  messages,
  onOpenChange,
  onModeChange,
  onPromptChange,
  onResourceChange,
  onRetryResources,
  onSubmit,
}: LoopAIRequestFormProps) {
  const canSubmit =
    resources.some((resource) => resource.id === selectedResourceId) &&
    prompt.trim().length > 0 &&
    !resourcesLoading &&
    !resourceError &&
    !busy;

  return (
    <>
      <DialogBody className="space-y-4">
        <LoopAIModeTabs
          busy={busy}
          messages={messages}
          mode="generate"
          onModeChange={onModeChange}
        />
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
          <Label htmlFor="loop-ai-prompt">{messages.prompt}</Label>
          <Textarea
            id="loop-ai-prompt"
            disabled={busy}
            rows={5}
            placeholder={messages.promptPlaceholder}
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
          {busy ? messages.generating : messages.generate}
        </Button>
      </DialogFooter>
    </>
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

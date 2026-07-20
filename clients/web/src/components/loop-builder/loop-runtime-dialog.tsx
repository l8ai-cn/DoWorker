"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { LoopRuntimeTemplate } from "@/lib/viewModels/loop-program";
import type { LoopRuntimeMessages } from "./loop-workbench-messages";

interface LoopRuntimeDialogProps {
  error?: string;
  loading: boolean;
  open: boolean;
  running: boolean;
  templates: LoopRuntimeTemplate[];
  messages: LoopRuntimeMessages;
  onOpenChange: (open: boolean) => void;
  onRetry: () => void;
  onRun: (templateName: string) => void;
}

function runtimeLabel(
  template: LoopRuntimeTemplate,
  messages: LoopRuntimeMessages,
): string {
  return messages.templateLabel(
    template.alias || messages.unnamed,
    template.workerType,
    template.id,
  );
}

export function LoopRuntimeDialog({
  error,
  loading,
  open,
  running,
  templates,
  messages,
  onOpenChange,
  onRetry,
  onRun,
}: LoopRuntimeDialogProps) {
  const [selectedId, setSelectedId] = useState("");
  const selected = templates.find(({ id }) => id === selectedId);

  function changeOpen(nextOpen: boolean) {
    if (!nextOpen) setSelectedId("");
    onOpenChange(nextOpen);
  }

  return (
    <Dialog open={open} onOpenChange={changeOpen} overlayClassName="z-[100001]">
      <DialogContent
        className="max-w-md"
        title={messages.title}
        description={messages.description}
      >
        <DialogBody className="space-y-3">
          <Label>{messages.field}</Label>
          <Select
            disabled={loading || templates.length === 0 || running}
            value={selectedId}
            onValueChange={setSelectedId}
          >
            <SelectTrigger>
              {selected ? (
                runtimeLabel(selected, messages)
              ) : (
                <SelectValue placeholder={messages.placeholder} />
              )}
            </SelectTrigger>
            <SelectContent>
              {templates.map((template) => (
                <SelectItem key={template.id} value={template.id}>
                  {runtimeLabel(template, messages)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {loading && (
            <p className="text-sm text-muted-foreground">{messages.loading}</p>
          )}
          {!loading && error && (
            <div className="flex items-center justify-between gap-3">
              <p className="text-sm text-destructive">{error}</p>
              <Button onClick={onRetry} variant="outline">
                {messages.retry}
              </Button>
            </div>
          )}
          {!loading && !error && templates.length === 0 && (
            <p className="text-sm text-destructive">{messages.empty}</p>
          )}
        </DialogBody>
        <DialogFooter>
          <Button disabled={running} onClick={() => changeOpen(false)} variant="outline">
            {messages.cancel}
          </Button>
          <Button
            disabled={loading || !selected || running}
            loading={running}
            onClick={() => {
              if (!selected) return;
              setSelectedId("");
              onRun(selected.id);
            }}
          >
            {messages.start}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

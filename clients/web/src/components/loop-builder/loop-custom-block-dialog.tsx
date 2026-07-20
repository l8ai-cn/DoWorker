"use client";

import { useId, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  buildCustomBlockDefinition,
  type LoopCustomBlockDefinition,
  type LoopCustomBlockDraft,
} from "./loop-custom-block-types";
import type { LoopCustomBlockMessages } from "./loop-workbench-messages";

interface LoopCustomBlockDialogProps {
  definitions: readonly LoopCustomBlockDefinition[];
  messages: LoopCustomBlockMessages;
  open: boolean;
  onCreate: (definition: LoopCustomBlockDefinition) => Promise<void>;
  onOpenChange: (open: boolean) => void;
}

const EMPTY_DRAFT: LoopCustomBlockDraft = {
  slug: "ppt-step",
  label: "",
  promptTemplate: "制作 {{topic}} 的专业 PPT",
  commandTemplate: "test -f {{file}}",
  acceptTemplate: "{{file}} 存在且可打开",
};

export function LoopCustomBlockDialog({
  definitions,
  messages,
  open,
  onCreate,
  onOpenChange,
}: LoopCustomBlockDialogProps) {
  const id = useId();
  const [draft, setDraft] = useState(EMPTY_DRAFT);
  const [submitted, setSubmitted] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string>();
  const result = useMemo(
    () => buildCustomBlockDefinition(draft, definitions),
    [definitions, draft],
  );

  function error(field: keyof LoopCustomBlockDraft): string | undefined {
    if (!submitted) return undefined;
    const issue = result.issues.find((item) => item.field === field);
    if (!issue) return undefined;
    if (issue.code === "identifier") return messages.identifier;
    return messages.required;
  }

  function update(field: keyof LoopCustomBlockDraft, value: string) {
    setSaveError(undefined);
    setDraft((current) => ({ ...current, [field]: value }));
  }

  function changeOpen(nextOpen: boolean) {
    if (!nextOpen) {
      setDraft(EMPTY_DRAFT);
      setSubmitted(false);
      setSaveError(undefined);
    }
    onOpenChange(nextOpen);
  }

  async function create() {
    setSubmitted(true);
    if (!result.definition) return;
    setSaving(true);
    setSaveError(undefined);
    try {
      await onCreate(result.definition);
      changeOpen(false);
    } catch (cause) {
      setSaveError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={changeOpen} overlayClassName="z-[100001]">
      <DialogContent
        className="max-w-xl"
        title={messages.title}
        description={messages.description}
      >
        <DialogBody className="space-y-4">
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="space-y-1">
              <Label htmlFor={`${id}-label`}>{messages.label}</Label>
              <Input
                autoFocus
                error={error("label")}
                id={`${id}-label`}
                value={draft.label}
                onChange={(event) => update("label", event.target.value)}
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor={`${id}-slug`}>{messages.slug}</Label>
              <Input
                error={error("slug")}
                id={`${id}-slug`}
                value={draft.slug}
                onChange={(event) => update("slug", event.target.value)}
              />
            </div>
          </div>
          <div className="space-y-1">
            <Label htmlFor={`${id}-prompt-template`}>{messages.promptTemplate}</Label>
            <Textarea
              error={error("promptTemplate")}
              id={`${id}-prompt-template`}
              rows={3}
              value={draft.promptTemplate}
              onChange={(event) => update("promptTemplate", event.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor={`${id}-command-template`}>{messages.commandTemplate}</Label>
            <Input
              error={error("commandTemplate")}
              id={`${id}-command-template`}
              value={draft.commandTemplate}
              onChange={(event) => update("commandTemplate", event.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor={`${id}-accept-template`}>{messages.acceptTemplate}</Label>
            <Input
              error={error("acceptTemplate")}
              id={`${id}-accept-template`}
              value={draft.acceptTemplate}
              onChange={(event) => update("acceptTemplate", event.target.value)}
            />
          </div>
          {saveError && <p className="text-sm text-destructive" role="alert">{saveError}</p>}
        </DialogBody>
        <DialogFooter>
          <Button disabled={saving} onClick={() => changeOpen(false)} variant="outline">
            {messages.cancel}
          </Button>
          <Button disabled={saving} onClick={() => void create()}>
            {messages.create}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

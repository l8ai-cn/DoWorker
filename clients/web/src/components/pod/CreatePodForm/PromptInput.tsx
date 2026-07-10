"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { WorkerSlashDropdown } from "@/components/shared/WorkerSlashDropdown";
import { useWorkerSlashComposer } from "@/hooks/useWorkerSlashComposer";
import type { WorkerSlashCommandDef } from "@/lib/workerSlashCommands";

interface PromptInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  t: (key: string) => string;
  enableSlashCommands?: boolean;
  extraSlashCommands?: WorkerSlashCommandDef[];
}

export function PromptInput({
  value,
  onChange,
  placeholder,
  t,
  enableSlashCommands = true,
  extraSlashCommands = [],
}: PromptInputProps) {
  const tRoot = useTranslations();
  const slash = useWorkerSlashComposer(tRoot, extraSlashCommands);
  const matches = useMemo(
    () => (enableSlashCommands ? slash.matchesFor(value, value.length) : []),
    [enableSlashCommands, slash, value],
  );

  return (
    <div>
      <label
        htmlFor="prompt-input"
        className="block text-sm font-medium mb-2"
      >
        {t("ide.createPod.prompt")}
      </label>
      <div className="relative">
        {enableSlashCommands && (
          <WorkerSlashDropdown
            commands={matches}
            activeIndex={slash.active}
            visible={slash.visible}
            onSelect={(command) => {
              onChange(slash.applySelection(command, value));
              slash.setVisible(false);
            }}
          />
        )}
        <textarea
          id="prompt-input"
          className="w-full px-3 py-2 border border-border rounded-md bg-background resize-none"
          rows={3}
          placeholder={placeholder || t("ide.createPod.promptPlaceholder")}
          value={value}
          onChange={(e) => {
            onChange(e.target.value);
            if (enableSlashCommands) {
              slash.syncMenu(e.target.value, e.target.selectionStart ?? e.target.value.length);
            }
          }}
          onKeyDown={(e) => {
            if (!enableSlashCommands) return;
            slash.handleKeyDown(
              e,
              value,
              matches,
              onChange,
              () => {
                // Create flow keeps the draft in the textarea; Enter only picks slash items.
              },
            );
          }}
        />
        {enableSlashCommands && (
          <p className="mt-1.5 text-[11px] text-muted-foreground">
            {tRoot("workerSlash.createHint")}
          </p>
        )}
      </div>
    </div>
  );
}

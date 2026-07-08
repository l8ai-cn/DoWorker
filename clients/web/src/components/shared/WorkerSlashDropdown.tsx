"use client";

import { CommandIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import type { WorkerSlashCommand } from "@/lib/workerSlashCommands";

interface WorkerSlashDropdownProps {
  commands: WorkerSlashCommand[];
  activeIndex: number;
  visible: boolean;
  onSelect: (command: WorkerSlashCommand) => void;
}

export function WorkerSlashDropdown({
  commands,
  activeIndex,
  visible,
  onSelect,
}: WorkerSlashDropdownProps) {
  if (!visible || commands.length === 0) return null;

  const safeActive = Math.min(activeIndex, Math.max(commands.length - 1, 0));

  return (
    <div
      data-testid="worker-slash-dropdown"
      className="absolute bottom-full left-0 right-0 z-20 mb-2 overflow-hidden rounded-lg border bg-popover shadow-md"
    >
      <div className="border-b px-3 py-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
        Commands
      </div>
      <div className="max-h-48 overflow-y-auto p-1">
        {commands.map((command, index) => (
          <button
            key={command.id}
            type="button"
            data-testid={`worker-slash-item-${command.id}`}
            data-active={index === safeActive ? "true" : undefined}
            className={cn(
              "flex w-full items-start gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-accent",
              index === safeActive && "bg-accent",
            )}
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => onSelect(command)}
          >
            <CommandIcon className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <span className="min-w-0 flex-1">
              <span className="block font-medium">{command.label}</span>
              <span className="block truncate text-xs text-muted-foreground">{command.hint}</span>
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}

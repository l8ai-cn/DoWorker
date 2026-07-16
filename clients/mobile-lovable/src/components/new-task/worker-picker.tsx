import { Check, ChevronDown } from "lucide-react";
import type { AgentPickerOption } from "@/lib/agent-display";
import { cn } from "@/lib/utils";

interface WorkerPickerProps {
  disabled: boolean;
  engines: AgentPickerOption[];
  current: AgentPickerOption | null;
  message: string | null;
  open: boolean;
  selectedID: string;
  onOpenChange: (open: boolean) => void;
  onSelect: (id: string) => void;
}

export function WorkerPicker({
  disabled,
  engines,
  current,
  message,
  open,
  selectedID,
  onOpenChange,
  onSelect,
}: WorkerPickerProps) {
  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between px-1">
        <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
          Agent 工具
        </p>
        <span className="text-[10px] text-muted-foreground/70">必选</span>
      </div>
      <button
        type="button"
        onClick={() => onOpenChange(!open)}
        disabled={disabled || !current}
        className={cn(
          "flex w-full items-center gap-3 rounded-2xl bg-card p-3 ring-1 ring-border/50 transition hover:ring-primary/40",
          disabled && "opacity-70",
        )}
      >
        {current ? (
          <>
            <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/15 text-lg ring-1 ring-white/5">
              {current.avatar}
            </span>
            <div className="min-w-0 flex-1 text-left">
              <p className="truncate text-[13.5px] font-semibold">{current.name}</p>
              <p className="truncate text-[10.5px] text-muted-foreground">
                {current.vendor} · {current.desc}
              </p>
            </div>
            <ChevronDown
              className={cn(
                "h-4 w-4 text-muted-foreground transition-transform",
                open && "rotate-180",
              )}
            />
          </>
        ) : (
          <p className="min-w-0 flex-1 text-left text-[12px] text-muted-foreground">{message}</p>
        )}
      </button>
      {open && !disabled && (
        <div className="mt-1 grid grid-cols-2 gap-1.5 rounded-2xl bg-card p-1.5 ring-1 ring-border/50 stream-in">
          {engines.map((engine) => (
            <button
              key={engine.id}
              type="button"
              onClick={() => {
                onSelect(engine.id);
                onOpenChange(false);
              }}
              className={cn(
                "flex items-center gap-2 rounded-xl px-2 py-2 text-left transition",
                engine.id === selectedID
                  ? "bg-primary/10 ring-1 ring-primary/40"
                  : "hover:bg-surface",
              )}
            >
              <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-surface text-sm">
                {engine.avatar}
              </span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-[12px] font-medium">{engine.name}</p>
                <p className="truncate text-[10px] text-muted-foreground">{engine.vendor}</p>
              </div>
              {engine.id === selectedID && <Check className="h-3 w-3 shrink-0 text-primary" />}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

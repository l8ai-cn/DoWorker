import { Check, ChevronDown } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { localProjectMeta } from "@/lib/projects-local";
import { cn } from "@/lib/utils";

interface ProjectPickerProps {
  names: string[];
  open: boolean;
  selectedName: string;
  onOpenChange: (open: boolean) => void;
  onSelect: (name: string) => void;
}

export function ProjectPicker({
  names,
  open,
  selectedName,
  onOpenChange,
  onSelect,
}: ProjectPickerProps) {
  const meta = localProjectMeta(selectedName);
  return (
    <div>
      <button
        type="button"
        onClick={() => onOpenChange(!open)}
        className="flex w-full items-center gap-3 rounded-2xl bg-card p-3 ring-1 ring-border/50 transition hover:ring-primary/40"
      >
        <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/15 ring-1 ring-white/5">
          <span className="h-2 w-2 rounded-full bg-success" />
        </span>
        <div className="min-w-0 flex-1 text-left">
          <p className="text-[10.5px] uppercase tracking-wider text-muted-foreground">目标项目</p>
          <p className="truncate text-[13.5px] font-semibold">{selectedName}</p>
          {meta?.repo && (
            <p className="truncate font-mono text-[10.5px] text-muted-foreground">{meta.repo}</p>
          )}
        </div>
        <ChevronDown
          className={cn("h-4 w-4 text-muted-foreground transition-transform", open && "rotate-180")}
        />
      </button>
      {open && (
        <div className="mt-1 space-y-1 rounded-2xl bg-card p-1.5 ring-1 ring-border/50 stream-in">
          {names.map((name) => (
            <button
              key={name}
              type="button"
              onClick={() => {
                onSelect(name);
                onOpenChange(false);
              }}
              className={cn(
                "flex w-full items-center gap-3 rounded-xl px-2.5 py-2 text-left transition",
                name === selectedName ? "bg-primary/10" : "hover:bg-surface",
              )}
            >
              <span className="h-2 w-2 rounded-full bg-success" />
              <div className="min-w-0 flex-1">
                <p className="truncate text-[13px] font-medium">{name}</p>
                {localProjectMeta(name)?.repo && (
                  <p className="truncate font-mono text-[10.5px] text-muted-foreground">
                    {localProjectMeta(name)?.repo}
                  </p>
                )}
              </div>
              {name === selectedName && <Check className="h-3.5 w-3.5 text-primary" />}
            </button>
          ))}
          <Link
            to="/projects/new"
            className="block rounded-xl px-2.5 py-2 text-center text-[11px] text-primary hover:bg-surface"
          >
            + 新建项目
          </Link>
        </div>
      )}
    </div>
  );
}

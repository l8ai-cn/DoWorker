import { Check, ChevronDown } from "lucide-react";
import { Link } from "@tanstack/react-router";
import type { LiveExpert } from "@/lib/experts-api";
import { cn } from "@/lib/utils";

interface ExpertPickerProps {
  authenticated: boolean;
  disabled: boolean;
  experts: LiveExpert[];
  current: LiveExpert | undefined;
  open: boolean;
  selectedSlug: string | undefined;
  onOpenChange: (open: boolean) => void;
  onSelect: (slug: string | undefined) => void;
}

export function ExpertPicker({
  authenticated,
  disabled,
  experts,
  current,
  open,
  selectedSlug,
  onOpenChange,
  onSelect,
}: ExpertPickerProps) {
  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between px-1">
        <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
          执行专家
        </p>
        <span className="text-[10px] text-muted-foreground/70">可选</span>
      </div>
      <button
        type="button"
        onClick={() => onOpenChange(!open)}
        disabled={disabled}
        className="flex w-full items-center gap-3 rounded-2xl bg-card p-3 ring-1 ring-border/50 transition hover:ring-primary/40"
      >
        <span
          className={cn(
            "flex h-9 w-9 items-center justify-center rounded-xl text-lg ring-1 ring-white/5",
            current ? "bg-primary/15" : "bg-surface",
          )}
        >
          🤖
        </span>
        <div className="min-w-0 flex-1 text-left">
          <p className="truncate text-[13.5px] font-semibold">{current?.name ?? "不使用专家"}</p>
          <p className="truncate text-[10.5px] text-muted-foreground">
            {current?.description ?? "叠加一位专家来注入领域能力"}
          </p>
        </div>
        <ChevronDown
          className={cn("h-4 w-4 text-muted-foreground transition-transform", open && "rotate-180")}
        />
      </button>
      {open && !disabled && (
        <div className="mt-1 max-h-64 space-y-1 overflow-y-auto rounded-2xl bg-card p-1.5 ring-1 ring-border/50 stream-in">
          <button
            type="button"
            onClick={() => {
              onSelect(undefined);
              onOpenChange(false);
            }}
            className={cn(
              "flex w-full items-center gap-3 rounded-xl px-2.5 py-2 text-left transition",
              !selectedSlug ? "bg-primary/10" : "hover:bg-surface",
            )}
          >
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-surface text-base">
              ✨
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-[13px] font-medium">通用 Agent</p>
              <p className="truncate text-[10.5px] text-muted-foreground">不指定专家</p>
            </div>
            {!selectedSlug && <Check className="h-3.5 w-3.5 text-primary" />}
          </button>
          {experts.map((expert) => (
            <button
              key={expert.slug}
              type="button"
              onClick={() => {
                onSelect(expert.slug);
                onOpenChange(false);
              }}
              className={cn(
                "flex w-full items-center gap-3 rounded-xl px-2.5 py-2 text-left transition",
                expert.slug === selectedSlug ? "bg-primary/10" : "hover:bg-surface",
              )}
            >
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/15 text-base">
                🤖
              </span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-[13px] font-medium">{expert.name}</p>
                <p className="truncate text-[10.5px] text-muted-foreground">
                  {expert.description ?? expert.agent_slug}
                </p>
              </div>
              {expert.slug === selectedSlug && <Check className="h-3.5 w-3.5 text-primary" />}
            </button>
          ))}
          {authenticated && experts.length === 0 && (
            <p className="px-2 py-3 text-center text-[11px] text-muted-foreground">
              组织暂无专家 ·{" "}
              <Link to="/experts" className="text-primary">
                专家库
              </Link>
            </p>
          )}
        </div>
      )}
    </div>
  );
}

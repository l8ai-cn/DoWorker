import type { ComponentType, ReactNode } from "react";

export function MetaChip({
  children,
  icon: Icon,
}: {
  children: ReactNode;
  icon?: ComponentType<{ className?: string }>;
}) {
  return (
    <span className="inline-flex shrink-0 items-center gap-1 rounded-md bg-surface px-2 py-0.5 font-mono text-[10px] text-muted-foreground ring-1 ring-border/40">
      {Icon && <Icon className="h-3 w-3" />}
      {children}
    </span>
  );
}

export function Metric({
  icon: Icon,
  label,
  value,
}: {
  icon: ComponentType<{ className?: string }>;
  label: string;
  value: string;
}) {
  return (
    <div className="flex flex-col items-center gap-0.5 bg-background/60 py-2">
      <div className="flex items-center gap-1 text-muted-foreground">
        <Icon className="h-3 w-3" />
        <span className="text-[10px] uppercase tracking-wider">{label}</span>
      </div>
      <span className="font-mono text-[12px] font-semibold tabular-nums">{value}</span>
    </div>
  );
}

export function formatTokens(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}

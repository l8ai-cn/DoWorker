import { Sparkles, TerminalSquare } from "lucide-react";

export function MarketplaceSignalPanel({ total, visible }: { total: number; visible: number }) {
  return (
    <div className="surface-card relative overflow-hidden p-6">
      <div className="absolute right-0 top-0 h-40 w-40 rounded-full bg-[var(--azure-cyan)]/10 blur-3xl" />
      <div className="relative space-y-5">
        <div className="flex items-center gap-3 text-[var(--azure-cyan)]">
          <Sparkles className="h-5 w-5" />
          <span className="font-headline text-xs font-bold uppercase tracking-[0.2em]">
            Marketplace index
          </span>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <Metric label="Published" value={String(total)} />
          <Metric label="Matching" value={String(visible)} />
        </div>
        <div className="rounded-2xl border border-border/70 bg-background/40 p-4">
          <div className="flex items-center gap-3">
            <TerminalSquare className="h-5 w-5 text-[var(--azure-cyan)]" />
            <p className="text-sm text-[var(--azure-text-muted)]">
              Skills install into repositories first, then workers load them through
              the existing SKILLS AgentFile directive.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-border/70 bg-background/35 p-4">
      <div className="text-3xl font-black tracking-tight text-foreground">{value}</div>
      <div className="mt-1 text-[10px] font-semibold uppercase tracking-[0.18em] text-[var(--azure-text-muted)]">
        {label}
      </div>
    </div>
  );
}

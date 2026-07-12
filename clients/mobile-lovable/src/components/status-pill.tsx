import { statusMeta, type SessionStatus } from "@/lib/session-types";
import { cn } from "@/lib/utils";

export function StatusPill({ status, className }: { status: SessionStatus; className?: string }) {
  const m = statusMeta[status];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full bg-surface-2/80 px-2 py-0.5 text-[10px] font-medium ring-1",
        m.ring,
        m.textClass,
        className,
      )}
    >
      <span className={cn("h-1.5 w-1.5 rounded-full", m.dotClass)} />
      {m.label}
    </span>
  );
}

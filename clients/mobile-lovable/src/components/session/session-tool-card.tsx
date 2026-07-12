import {
  Check,
  ChevronDown,
  CircleAlert,
  FileEdit,
  FilePlus,
  FileText,
  Globe,
  Hammer,
  Loader2,
  Search,
  ShieldAlert,
  Terminal as TerminalIcon,
  X,
} from "lucide-react";
import type { ComponentType } from "react";
import { useState } from "react";
import type { AgentEvent, DiffHunk, ToolKind } from "@/lib/session-types";
import { useDecision } from "@/lib/session-approval-decisions";
import { cn } from "@/lib/utils";

export const TOOL_KIND_META: Record<ToolKind, { Icon: ComponentType<{ className?: string }>; color: string; ring: string; label: string }> = {
  read:   { Icon: FileText,      color: "text-event-tool",     ring: "ring-event-tool/30", label: "读文件" },
  write:  { Icon: FilePlus,      color: "text-warning",        ring: "ring-warning/40",    label: "写文件" },
  edit:   { Icon: FileEdit,      color: "text-warning",        ring: "ring-warning/40",    label: "改文件" },
  shell:  { Icon: TerminalIcon,  color: "text-event-tool",     ring: "ring-event-tool/30", label: "命令" },
  search: { Icon: Search,        color: "text-info",           ring: "ring-info/30",       label: "搜索" },
  fetch:  { Icon: Globe,         color: "text-info",           ring: "ring-info/30",       label: "网络" },
  other:  { Icon: Hammer,        color: "text-muted-foreground", ring: "ring-border",      label: "工具" },
};

export function ToolCard({ event }: { event: AgentEvent }) {
  const kind = event.toolKind ?? "other";
  const meta = TOOL_KIND_META[kind];
  const isApproval = event.type === "permission_request";
  const isRunning = event.status === "in_progress";
  const isFailed = event.status === "failed";
  const decision = useDecision(event.id);
  const [open, setOpen] = useState(isApproval || isRunning);

  const hasBody = Boolean(
    event.command || event.output || event.diff?.length || event.results?.length || event.detail,
  );

  return (
    <div className={cn(
      "stream-in overflow-hidden rounded-xl border bg-card/60",
      isApproval ? "border-warning/50 ring-1 ring-warning/20" : "border-border/50",
      isFailed && "border-destructive/50",
    )}>
      <button
        type="button"
        onClick={() => hasBody && setOpen((o) => !o)}
        className={cn(
          "flex w-full items-center gap-2 px-3 py-2 text-left",
          hasBody && "hover:bg-surface/50",
        )}
      >
        <span className={cn(
          "flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-surface ring-1",
          meta.ring,
        )}>
          <meta.Icon className={cn("h-3.5 w-3.5", meta.color)} />
        </span>
        <span className="rounded bg-surface-2 px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
          {event.tool ?? meta.label}
        </span>
        <span className="min-w-0 flex-1 truncate text-[12.5px] font-medium">
          {event.filePath ?? event.command ?? event.query ?? event.title}
        </span>
        <StatusDot status={event.status} isApproval={isApproval} />
        <span className="font-mono text-[9.5px] text-muted-foreground">{event.duration ?? event.ts}</span>
        {hasBody && <ChevronDown className={cn("h-3 w-3 text-muted-foreground transition-transform", open && "rotate-180")} />}
      </button>

      {open && hasBody && (
        <div className="border-t border-border/40 bg-background/40">
          {(kind === "write" || kind === "edit") && event.filePath && (
            <div className="flex items-center justify-between border-b border-border/40 px-3 py-1.5 font-mono text-[10.5px] text-muted-foreground">
              <span className="truncate">{event.filePath}</span>
              <span className="shrink-0 space-x-2">
                {event.additions != null && <span className="text-success">+{event.additions}</span>}
                {event.deletions != null && <span className="text-destructive">−{event.deletions}</span>}
              </span>
            </div>
          )}
          {event.detail && (
            <p className="px-3 pt-2 text-[12px] text-muted-foreground">{event.detail}</p>
          )}

          {kind === "shell" && event.command && (
            <TerminalView command={event.command} cwd={event.cwd} output={event.output} exitCode={event.exitCode} running={isRunning} />
          )}

          {event.diff?.length ? <DiffView hunks={event.diff} /> : null}

          {event.results?.length ? (
            <ul className="divide-y divide-border/40 px-3 py-2">
              {event.results.map((r, i) => (
                <li key={i} className="py-1.5">
                  <p className="truncate font-mono text-[11px] text-primary">{r.title}</p>
                  {r.snippet && <p className="truncate font-mono text-[10.5px] text-muted-foreground">{r.snippet}</p>}
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      )}

      {isApproval && (
        <div className="border-t border-warning/20 bg-warning/5 px-3 py-1.5">
          {decision ? (
            <div className="flex items-center gap-1.5 text-[11px]">
              {decision === "approved" ? (
                <><Check className="h-3 w-3 text-success" /><span className="text-success">已批准</span></>
              ) : (
                <><X className="h-3 w-3 text-muted-foreground" /><span className="text-muted-foreground">已拒绝</span></>
              )}
              <span className="ml-auto font-mono text-[10px] text-muted-foreground">{event.tool ?? meta.label}</span>
            </div>
          ) : (
            <div className="flex items-center gap-1.5 text-[11px] text-warning">
              <ShieldAlert className="h-3 w-3 shrink-0" />
              <span className="min-w-0 flex-1 truncate">等待你在下方确认</span>
              <span className="font-mono text-[10px] text-muted-foreground">{event.tool ?? meta.label}</span>
            </div>
          )}
        </div>
      )}

    </div>
  );
}

export function StatusDot({ status, isApproval }: { status?: AgentEvent["status"]; isApproval?: boolean }) {
  if (isApproval) return <span className="rounded-full bg-warning/20 px-1.5 py-0.5 text-[9.5px] font-semibold text-warning">待审批</span>;
  if (status === "in_progress") return <Loader2 className="h-3 w-3 animate-spin text-primary" />;
  if (status === "failed")      return <CircleAlert className="h-3 w-3 text-destructive" />;
  if (status === "completed")   return <Check className="h-3 w-3 text-success" />;
  return null;
}

export function TerminalView({ command, cwd, output, exitCode, running }: {
  command: string; cwd?: string; output?: string; exitCode?: number; running?: boolean;
}) {
  return (
    <div className="m-3 overflow-hidden rounded-lg bg-[#0a0d12] ring-1 ring-border/60">
      <div className="flex items-center gap-1.5 border-b border-white/5 px-2.5 py-1">
        <span className="h-2 w-2 rounded-full bg-destructive/70" />
        <span className="h-2 w-2 rounded-full bg-warning/70" />
        <span className="h-2 w-2 rounded-full bg-success/70" />
        <span className="ml-2 font-mono text-[10px] text-muted-foreground">{cwd ?? "shell"}</span>
        <span className="ml-auto font-mono text-[10px] text-muted-foreground">
          {running ? "running..." : exitCode === 0 ? "exit 0" : exitCode != null ? `exit ${exitCode}` : ""}
        </span>
      </div>
      <div className="max-h-72 overflow-auto p-2.5 font-mono text-[11px] leading-relaxed">
        <div className="flex gap-2 text-success">
          <span className="select-none">$</span>
          <span className="whitespace-pre-wrap break-all text-foreground/95">{command}</span>
        </div>
        {output && (
          <pre className="mt-1 whitespace-pre-wrap text-foreground/75">{output}{running && <span className="terminal-caret" />}</pre>
        )}
      </div>
    </div>
  );
}

export function DiffView({ hunks }: { hunks: DiffHunk[] }) {
  return (
    <div className="m-3 overflow-hidden rounded-lg bg-[#0a0d12] font-mono text-[11px] leading-relaxed ring-1 ring-border/60">
      {hunks.map((h, i) => (
        <div key={i}>
          {h.header && (
            <div className="border-b border-white/5 bg-info/5 px-2.5 py-1 text-[10px] text-info/80">{h.header}</div>
          )}
          <div className="max-h-80 overflow-auto py-1">
            {h.lines.map((l, j) => (
              <div key={j} className={cn(
                "flex gap-2 px-2.5",
                l.kind === "add" && "bg-success/10 text-success",
                l.kind === "del" && "bg-destructive/10 text-destructive",
                l.kind === "ctx" && "text-foreground/60",
              )}>
                <span className="w-3 select-none text-center opacity-60">
                  {l.kind === "add" ? "+" : l.kind === "del" ? "−" : " "}
                </span>
                <span className="whitespace-pre">{l.text || " "}</span>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

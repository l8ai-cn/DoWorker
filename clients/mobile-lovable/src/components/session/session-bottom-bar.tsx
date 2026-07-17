import { ShieldAlert, Square } from "lucide-react";
import type { AgentEvent } from "@/lib/session-types";
import { useSessionActions } from "@/lib/session-action-state";
import { setApprovalDecision } from "@/lib/session-approval-decisions";
import { SessionComposer } from "@/components/session/session-composer";
import { TOOL_KIND_META } from "@/components/session/session-tool-card";

export function SessionBottomBar({ pendingApproval }: { pendingApproval?: AgentEvent }) {
  return (
    <div className="safe-bottom fixed bottom-0 left-0 right-0 z-30 md:absolute">
      {pendingApproval && <ApprovalBar event={pendingApproval} />}
      <SessionComposer />
    </div>
  );
}

function ApprovalBar({ event }: { event: AgentEvent }) {
  const actions = useSessionActions();
  const meta = TOOL_KIND_META[event.toolKind ?? "other"];
  const label = event.tool ?? meta.label;
  const decide = (accept: boolean) => {
    setApprovalDecision(event.id, accept ? "approved" : "rejected");
    if (event.elicitationId && actions.onApprove) {
      void actions.onApprove(event.elicitationId, accept);
    }
  };
  return (
    <div className="border-t border-warning/40 bg-warning/10 px-3 py-2 backdrop-blur-xl stream-in">
      <div className="flex items-center gap-2">
        <ShieldAlert className="h-4 w-4 shrink-0 text-warning" />
        <div className="min-w-0 flex-1">
          <p className="truncate text-[12px] font-semibold text-warning leading-tight">
            需要你的批准
          </p>
          <p className="truncate text-[11px] text-muted-foreground leading-tight">
            <span className="font-mono">{label}</span>
            {event.filePath && (
              <>
                {" "}
                · <span className="font-mono">{event.filePath}</span>
              </>
            )}
          </p>
        </div>
        <button
          onClick={() => decide(false)}
          className="shrink-0 whitespace-nowrap rounded-full border border-border bg-surface px-3 py-1.5 text-[12px] font-medium hover:bg-surface-2"
        >
          拒绝
        </button>
        <button
          onClick={() => decide(true)}
          className="shrink-0 whitespace-nowrap rounded-full bg-primary px-3 py-1.5 text-[12px] font-semibold text-primary-foreground glow-primary"
        >
          批准
        </button>
      </div>
      <button
        onClick={() => decide(true)}
        className="mt-1 flex w-full items-center justify-center gap-1 text-[10.5px] text-muted-foreground hover:text-foreground"
      >
        <Square className="h-2.5 w-2.5" /> 本会话内始终允许{" "}
        <span className="font-mono">{label}</span>
      </button>
    </div>
  );
}

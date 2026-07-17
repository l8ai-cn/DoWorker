import { Bot, CircleAlert, Wifi, WifiOff } from "lucide-react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type { AgentSessionSnapshot } from "./contracts";

export function WorkspaceHeader({
  snapshot,
}: {
  snapshot: AgentSessionSnapshot;
}) {
  const connected = snapshot.connection === "connected";
  const text = useAgentWorkspaceText();
  return (
    <header className="flex min-h-12 items-center gap-3 border-b border-border px-3">
      <div className="flex size-8 shrink-0 items-center justify-center rounded-md bg-muted">
        <Bot className="size-4" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium">{snapshot.title}</div>
        <div className="flex min-w-0 items-center gap-1.5 overflow-hidden text-xs text-muted-foreground">
          <span className="shrink-0">{snapshot.agentLabel}</span>
          {(snapshot.metadata ?? []).map((item) => (
            <span className="contents" key={item.id}>
              <span aria-hidden="true">·</span>
              <span className="truncate" title={`${item.label}: ${item.value}`}>
                {item.value}
              </span>
            </span>
          ))}
        </div>
      </div>
      <div className="flex shrink-0 items-center gap-1.5 text-xs text-muted-foreground">
        {snapshot.status === "failed" ? (
          <CircleAlert className="size-3.5 text-destructive" />
        ) : connected ? (
          <Wifi className="size-3.5 text-emerald-600" />
        ) : (
          <WifiOff className="size-3.5" />
        )}
        <span>{text.sessionStatus(snapshot.status, snapshot.connection)}</span>
      </div>
    </header>
  );
}

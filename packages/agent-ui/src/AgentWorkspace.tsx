import { History, Loader2, MessageSquare, Terminal } from "lucide-react";
import { useState } from "react";

import { ActivityTimeline } from "./ActivityTimeline";
import { AgentWorkspaceLocaleProvider } from "./AgentWorkspaceLocaleContext";
import { ApprovalDock } from "./ApprovalDock";
import { ConversationComposer } from "./ConversationComposer";
import { ConversationEmptyState } from "./ConversationEmptyState";
import { PlanStrip } from "./PlanStrip";
import { TerminalSurface } from "./TerminalSurface";
import type { AgentSessionRuntime, TerminalRuntime } from "./contracts";
import { useAgentSessionSnapshot } from "./useAgentSessionSnapshot";
import { WorkspaceHeader } from "./WorkspaceHeader";
import {
  agentWorkspaceText,
  type AgentWorkspaceLocale,
} from "./agentWorkspaceText";

export interface AgentWorkspaceProps {
  runtime: AgentSessionRuntime;
  terminalRuntime?: TerminalRuntime;
  sessionId: string;
  clientLabel?: string;
  className?: string;
  locale?: AgentWorkspaceLocale;
}

export function AgentWorkspace({
  runtime,
  terminalRuntime,
  sessionId,
  clientLabel = "agent-workspace",
  className = "",
  locale = "en-US",
}: AgentWorkspaceProps) {
  const snapshot = useAgentSessionSnapshot(runtime, sessionId);
  const text = agentWorkspaceText(locale);
  const [view, setView] = useState<"conversation" | "terminal">("conversation");
  const [loadingOlder, setLoadingOlder] = useState(false);
  const [surfaceError, setSurfaceError] = useState<string | null>(null);
  const terminal = snapshot.terminals[0];
  const terminalEnabled =
    snapshot.capabilities.terminal && terminalRuntime !== undefined && terminal !== undefined;

  return (
    <AgentWorkspaceLocaleProvider locale={locale}>
      <div
        className={`flex h-full min-h-0 flex-col overflow-hidden bg-background text-foreground ${className}`}
        data-agent-workspace={sessionId}
      >
      <WorkspaceHeader snapshot={snapshot} />
      <nav className="flex h-10 items-center gap-1 border-b border-border px-2" role="tablist">
        <ViewTab
          active={view === "conversation"}
          icon={<MessageSquare className="size-3.5" />}
          label={text.conversation}
          onClick={() => setView("conversation")}
        />
        {terminalEnabled && (
          <ViewTab
            active={view === "terminal"}
            icon={<Terminal className="size-3.5" />}
            label={text.terminal}
            onClick={() => setView("terminal")}
          />
        )}
      </nav>
      {(snapshot.error || surfaceError) && (
        <div className="border-b border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {snapshot.error || surfaceError}
        </div>
      )}
      {view === "terminal" && terminalEnabled ? (
        <TerminalSurface
          clientLabel={clientLabel}
          resource={terminal}
          runtime={terminalRuntime}
        />
      ) : (
        <>
          <PlanStrip steps={snapshot.plan} />
          {snapshot.items.length === 0 &&
          snapshot.permissions.length === 0 ? (
            <main className="flex min-h-0 flex-1 flex-col justify-center gap-5 overflow-y-auto py-6">
              <ConversationEmptyState agentLabel={snapshot.agentLabel} />
              <ConversationComposer
                onError={(cause) =>
                  setSurfaceError(
                    cause instanceof Error ? cause.message : String(cause),
                  )
                }
                runtime={runtime}
                snapshot={snapshot}
              />
            </main>
          ) : (
            <>
              <main className="min-h-0 flex-1 overflow-y-auto">
                {snapshot.hasOlderItems && (
                  <div className="flex justify-center px-3 pt-3">
                    <button
                      className="flex h-8 items-center gap-1.5 rounded-md border border-border px-3 text-xs text-muted-foreground"
                      disabled={loadingOlder}
                      onClick={() => {
                        setLoadingOlder(true);
                        setSurfaceError(null);
                        void runtime
                          .loadOlder(sessionId)
                          .catch((cause) =>
                            setSurfaceError(
                              cause instanceof Error
                                ? cause.message
                                : String(cause),
                            ),
                          )
                          .finally(() => setLoadingOlder(false));
                      }}
                      type="button"
                    >
                      {loadingOlder ? (
                        <Loader2 className="size-3.5 animate-spin" />
                      ) : (
                        <History className="size-3.5" />
                      )}
                      {text.loadEarlierActivity}
                    </button>
                  </div>
                )}
                <ActivityTimeline
                  items={snapshot.items}
                  runtime={runtime}
                  sessionId={sessionId}
                />
              </main>
              <ApprovalDock
                onError={(cause) =>
                  setSurfaceError(
                    cause instanceof Error ? cause.message : String(cause),
                  )
                }
                permissions={snapshot.permissions}
                runtime={runtime}
                sessionId={sessionId}
              />
              <ConversationComposer
                onError={(cause) =>
                  setSurfaceError(
                    cause instanceof Error ? cause.message : String(cause),
                  )
                }
                runtime={runtime}
                snapshot={snapshot}
              />
            </>
          )}
        </>
      )}
      </div>
    </AgentWorkspaceLocaleProvider>
  );
}

function ViewTab({
  active,
  icon,
  label,
  onClick,
}: {
  active: boolean;
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      aria-selected={active}
      className={`flex h-8 items-center gap-1.5 rounded-md px-2 text-xs ${
        active ? "bg-muted font-medium" : "text-muted-foreground hover:bg-muted/60"
      }`}
      onClick={onClick}
      role="tab"
      type="button"
    >
      {icon}
      {label}
    </button>
  );
}

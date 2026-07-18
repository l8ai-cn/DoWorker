import { MessageSquare, Terminal } from "lucide-react";
import { useId, useMemo, useState } from "react";

import { AgentConversationSurface } from "./AgentConversationSurface";
import { AgentWorkspaceLocaleProvider } from "./AgentWorkspaceLocaleContext";
import { PlanStrip } from "./PlanStrip";
import { TerminalSurface } from "./TerminalSurface";
import type { AgentSessionRuntime, TerminalRuntime } from "./contracts";
import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import type { AgentContentRendererRegistration } from "./react/contentRendererTypes";
import type { ContentRendererRegistry } from "./registry/ContentRendererRegistry";
import type { ToolRendererRegistry } from "./registry/ToolRendererRegistry";
import { ResultWorkbench } from "./react/ResultWorkbench";
import { useWorkbenchContainerMode } from "./react/useWorkbenchContainerMode";
import { useAgentSessionSnapshot } from "./useAgentSessionSnapshot";
import { WorkspaceHeader } from "./WorkspaceHeader";
import { ReadOnlyAgentSessionRuntime } from "./runtime/ReadOnlyAgentSessionRuntime";
import {
  agentWorkspaceText,
  type AgentWorkspaceLocale,
} from "./agentWorkspaceText";
import { focusAdjacentTab } from "./react/tabKeyboardNavigation";

export interface AgentWorkspaceProps {
  runtime: AgentSessionRuntime;
  terminalRuntime?: TerminalRuntime;
  sessionId: string;
  clientLabel?: string;
  className?: string;
  contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
  locale?: AgentWorkspaceLocale;
  readOnly?: boolean;
  toolRenderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
}

export function AgentWorkspace({
  runtime,
  terminalRuntime,
  sessionId,
  clientLabel = "agent-workspace",
  className = "",
  contentRenderers,
  locale = "en-US",
  readOnly = false,
  toolRenderers,
}: AgentWorkspaceProps) {
  const activeRuntime = useMemo(
    () => (readOnly ? new ReadOnlyAgentSessionRuntime(runtime) : runtime),
    [readOnly, runtime],
  );
  const snapshot = useAgentSessionSnapshot(activeRuntime, sessionId, runtime);
  const text = agentWorkspaceText(locale);
  const [view, setView] = useState<"conversation" | "terminal">("conversation");
  const tabId = useId();
  const conversationTabId = `${tabId}-conversation-tab`;
  const conversationPanelId = `${tabId}-conversation-panel`;
  const terminalTabId = `${tabId}-terminal-tab`;
  const terminalPanelId = `${tabId}-terminal-panel`;
  const [surfaceError, setSurfaceError] = useState<string | null>(null);
  const { containerRef, mode } = useWorkbenchContainerMode();
  const terminal = snapshot.terminals[0];
  const artifacts = snapshot.items.filter((item) => item.kind === "artifact");
  const conversationItems = snapshot.items.filter(
    (item) => item.kind !== "artifact",
  );
  const terminalEnabled =
    snapshot.capabilities.terminal && terminalRuntime !== undefined && terminal !== undefined;

  return (
    <AgentWorkspaceLocaleProvider locale={locale}>
      <div
        className={`flex h-full min-h-0 flex-col overflow-hidden bg-background text-foreground ${className}`}
        data-agent-workspace={sessionId}
        ref={containerRef}
      >
      <WorkspaceHeader snapshot={snapshot} />
      <nav
        className="flex h-12 items-center gap-1 border-b border-border px-2"
        onKeyDown={focusAdjacentTab}
        role="tablist"
      >
        <ViewTab
          active={view === "conversation"}
          icon={<MessageSquare className="size-3.5" />}
          id={conversationTabId}
          label={text.conversation}
          onClick={() => setView("conversation")}
          panelId={conversationPanelId}
        />
        {terminalEnabled && (
          <ViewTab
            active={view === "terminal"}
            icon={<Terminal className="size-3.5" />}
            id={terminalTabId}
            label={text.terminal}
            onClick={() => setView("terminal")}
            panelId={terminalPanelId}
          />
        )}
      </nav>
      {(snapshot.error || surfaceError) && (
        <div className="border-b border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {snapshot.error || surfaceError}
        </div>
      )}
      {view === "terminal" && terminalEnabled ? (
        <section
          aria-labelledby={terminalTabId}
          className="flex min-h-0 flex-1"
          id={terminalPanelId}
          role="tabpanel"
        >
          <TerminalSurface
            clientLabel={clientLabel}
            resource={terminal}
            runtime={terminalRuntime}
          />
        </section>
      ) : (
        <section
          aria-labelledby={conversationTabId}
          className="flex min-h-0 flex-1 flex-col"
          id={conversationPanelId}
          role="tabpanel"
        >
          <PlanStrip steps={snapshot.plan} />
          <ResultWorkbench
            artifacts={artifacts}
            contentRenderers={contentRenderers}
            conversation={
              <AgentConversationSurface
                contentRenderers={contentRenderers}
                items={conversationItems}
                onError={(cause) =>
                  setSurfaceError(
                    cause instanceof Error ? cause.message : String(cause),
                  )
                }
                runtime={activeRuntime}
                snapshot={snapshot}
                toolRenderers={toolRenderers}
              />
            }
            mode={mode}
            runtime={activeRuntime}
            sessionId={sessionId}
            toolRenderers={toolRenderers}
            tools={conversationItems.filter((item) => item.kind === "tool")}
          />
        </section>
      )}
      </div>
    </AgentWorkspaceLocaleProvider>
  );
}

function ViewTab({
  active,
  icon,
  id,
  label,
  onClick,
  panelId,
}: {
  active: boolean;
  icon: React.ReactNode;
  id: string;
  label: string;
  onClick: () => void;
  panelId: string;
}) {
  return (
    <button
      aria-controls={panelId}
      aria-selected={active}
      className={`flex h-11 items-center gap-1.5 rounded-md px-3 text-xs ${
        active ? "bg-muted font-medium" : "text-muted-foreground hover:bg-muted/60"
      }`}
      id={id}
      onClick={onClick}
      role="tab"
      tabIndex={active ? 0 : -1}
      type="button"
    >
      {icon}
      {label}
    </button>
  );
}

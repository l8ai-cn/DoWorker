import { useId, useState, type ReactNode } from "react";

import type { AgentArtifactItem } from "../agentArtifactContracts";
import type { AgentToolActivityItem } from "../agentToolContracts";
import { useAgentWorkspaceText } from "../AgentWorkspaceLocaleContext";
import type { AgentSessionRuntime } from "../contracts";
import type { ContentRendererRegistry } from "../registry/ContentRendererRegistry";
import type { ToolRendererRegistry } from "../registry/ToolRendererRegistry";
import type { AgentContentRendererRegistration } from "./contentRendererTypes";
import type { AgentToolRendererRegistration } from "./rendererTypes";
import type { WorkbenchContainerMode } from "./useWorkbenchContainerMode";
import {
  collectWorkbenchResults,
  type WorkbenchResult,
} from "./workbenchResults";
import { WorkbenchResultsPane } from "./WorkbenchResultsPane";
import { focusAdjacentTab } from "./tabKeyboardNavigation";

export interface ResultWorkbenchProps {
  artifacts: AgentArtifactItem[];
  contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
  conversation: ReactNode;
  mode: WorkbenchContainerMode;
  runtime: AgentSessionRuntime;
  sessionId: string;
  toolRenderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
  tools?: AgentToolActivityItem[];
}

export function ResultWorkbench({
  artifacts,
  contentRenderers,
  conversation,
  mode,
  runtime,
  sessionId,
  toolRenderers,
  tools = [],
}: ResultWorkbenchProps) {
  const text = useAgentWorkspaceText();
  const [narrowPane, setNarrowPane] =
    useState<"conversation" | "results">("conversation");
  const [resultsActivated, setResultsActivated] = useState(false);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const tabId = useId();
  const conversationTabId = `${tabId}-conversation-tab`;
  const conversationPanelId = `${tabId}-conversation-panel`;
  const resultsTabId = `${tabId}-results-tab`;
  const resultsPanelId = `${tabId}-results-panel`;
  const workbenchResults = collectWorkbenchResults(
    artifacts,
    tools,
    toolRenderers,
  );
  const selected = selectedResult(workbenchResults, selectedId);

  if (workbenchResults.length === 0) return conversation;
  const results = (
    <WorkbenchResultsPane
      contentRenderers={contentRenderers}
      onSelect={setSelectedId}
      results={workbenchResults}
      runtime={runtime}
      selected={selected}
      sessionId={sessionId}
    />
  );

  if (mode !== "narrow") {
    return (
      <div
        className={`grid min-h-0 flex-1 ${
          mode === "wide"
            ? "grid-cols-[minmax(0,3fr)_minmax(360px,2fr)]"
            : "grid-cols-[minmax(0,1fr)_minmax(320px,44%)]"
        }`}
      >
        <section className="min-h-0 min-w-0">{conversation}</section>
        <aside
          aria-label={text.results}
          className="min-h-0 min-w-0 border-l border-border"
        >
          {results}
        </aside>
      </div>
    );
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <nav
        aria-label={text.workspaceViews}
        className="flex h-12 shrink-0 items-center gap-1 border-b border-border px-2"
        onKeyDown={focusAdjacentTab}
        role="tablist"
      >
        <PaneTab
          active={narrowPane === "conversation"}
          id={conversationTabId}
          label={text.conversation}
          onClick={() => setNarrowPane("conversation")}
          panelId={conversationPanelId}
        />
        <PaneTab
          active={narrowPane === "results"}
          id={resultsTabId}
          label={text.results}
          onClick={() => {
            setResultsActivated(true);
            setNarrowPane("results");
          }}
          panelId={resultsPanelId}
        />
      </nav>
      <section
        aria-labelledby={conversationTabId}
        aria-hidden={narrowPane !== "conversation"}
        className={`min-h-0 flex-1 ${narrowPane === "conversation" ? "" : "hidden"}`}
        id={conversationPanelId}
        inert={narrowPane !== "conversation"}
        role="tabpanel"
      >
        {conversation}
      </section>
      <section
        aria-labelledby={resultsTabId}
        aria-hidden={narrowPane !== "results"}
        className={`min-h-0 flex-1 ${narrowPane === "results" ? "" : "hidden"}`}
        id={resultsPanelId}
        inert={narrowPane !== "results"}
        role="tabpanel"
      >
        {resultsActivated ? results : null}
      </section>
    </div>
  );
}

function selectedResult(
  results: readonly WorkbenchResult[],
  selectedId: string | null,
): WorkbenchResult | undefined {
  return results.find((result) => result.id === selectedId) ?? results.at(-1);
}

function PaneTab({
  active,
  id,
  label,
  onClick,
  panelId,
}: {
  active: boolean;
  id: string;
  label: string;
  onClick: () => void;
  panelId: string;
}) {
  return (
    <button
      aria-controls={panelId}
      aria-selected={active}
      className={`h-11 rounded-md px-3 text-xs ${
        active ? "bg-muted font-medium" : "text-muted-foreground"
      }`}
      id={id}
      onClick={onClick}
      role="tab"
      tabIndex={active ? 0 : -1}
      type="button"
    >
      {label}
    </button>
  );
}

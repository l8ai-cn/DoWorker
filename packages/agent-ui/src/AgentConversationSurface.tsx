import { History, Loader2 } from "lucide-react";
import { type ReactNode, useState } from "react";

import { ActivityTimeline } from "./ActivityTimeline";
import { ApprovalDock } from "./ApprovalDock";
import { ConversationComposer } from "./ConversationComposer";
import { ConversationEmptyState } from "./ConversationEmptyState";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type {
  AgentSessionRuntime,
  AgentSessionSnapshot,
  AgentTimelineItem,
} from "./contracts";
import type { ContentRendererRegistry } from "./registry/ContentRendererRegistry";
import type { ToolRendererRegistry } from "./registry/ToolRendererRegistry";
import type { AgentContentRendererRegistration } from "./react/contentRendererTypes";
import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import type { AgentWorkspacePresentation } from "./userWorkspacePresentation";

export interface AgentConversationSurfaceProps {
  contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
  executionTrace?: ReactNode;
  items: AgentTimelineItem[];
  onError: (cause: unknown) => void;
  presentation: AgentWorkspacePresentation;
  runtime: AgentSessionRuntime;
  snapshot: AgentSessionSnapshot;
  toolRenderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
}

export function AgentConversationSurface({
  contentRenderers,
  executionTrace,
  items,
  onError,
  presentation,
  runtime,
  snapshot,
  toolRenderers,
}: AgentConversationSurfaceProps) {
  const text = useAgentWorkspaceText();
  const [loadingOlder, setLoadingOlder] = useState(false);
  const isEmpty =
    !executionTrace && items.length === 0 && snapshot.permissions.length === 0;

  if (isEmpty) {
    return (
      <main className="flex h-full min-h-0 flex-col justify-center gap-5 overflow-y-auto py-6">
        <ConversationEmptyState agentLabel={snapshot.agentLabel} />
        <ConversationComposer
          onError={onError}
          presentation={presentation}
          runtime={runtime}
          snapshot={snapshot}
        />
      </main>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      {executionTrace}
      <main className="min-h-0 flex-1 overflow-y-auto">
        {snapshot.hasOlderItems && (
          <div className="flex justify-center px-3 pt-3">
            <button
              className="flex h-8 items-center gap-1.5 rounded-md border border-border px-3 text-xs text-muted-foreground"
              disabled={loadingOlder}
              onClick={() => {
                setLoadingOlder(true);
                void runtime
                  .loadOlder(snapshot.sessionId)
                  .catch(onError)
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
          contentRenderers={contentRenderers}
          items={items}
          runtime={runtime}
          sessionId={snapshot.sessionId}
          toolRenderers={toolRenderers}
        />
      </main>
      <ApprovalDock
        disabled={!snapshot.capabilities.resolvePermission}
        onError={onError}
        permissions={snapshot.permissions}
        runtime={runtime}
        sessionId={snapshot.sessionId}
      />
      <ConversationComposer
        onError={onError}
        presentation={presentation}
        runtime={runtime}
        snapshot={snapshot}
      />
    </div>
  );
}

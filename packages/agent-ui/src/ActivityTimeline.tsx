import { AlertTriangle, CheckCircle2, Loader2 } from "lucide-react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import { ArtifactCard } from "./ArtifactCard";
import type {
  AgentSessionRuntime,
  AgentTimelineItem,
} from "./contracts";
import { MarkdownMessage } from "./MarkdownMessage";
import { ToolActivityGroup } from "./ToolActivityGroup";
import { groupToolActivity } from "./toolActivityGrouping";
import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import type { AgentContentRendererRegistration } from "./react/contentRendererTypes";
import type { ContentRendererRegistry } from "./registry/ContentRendererRegistry";
import type { ToolRendererRegistry } from "./registry/ToolRendererRegistry";

export function ActivityTimeline({
  items,
  runtime,
  sessionId,
  contentRenderers,
  toolRenderers,
}: {
  contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
  items: AgentTimelineItem[];
  runtime: AgentSessionRuntime;
  sessionId: string;
  toolRenderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
}) {
  const text = useAgentWorkspaceText();
  if (items.length === 0) {
    return (
      <div className="flex min-h-48 flex-col items-center justify-center gap-1 px-6 text-center">
        <div className="text-sm font-medium">{text.readyForTask}</div>
        <div className="text-xs text-muted-foreground">
          {text.startSession}
        </div>
      </div>
    );
  }
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-3 px-4 py-4">
      {groupToolActivity(items).map((item) =>
        item.kind === "tool-run" ? (
          <ToolActivityGroup
            key={item.id}
            renderers={toolRenderers}
            tools={item.tools}
          />
        ) : item.kind === "message" ? (
          <MessageRow item={item} key={item.id} />
        ) : item.kind === "artifact" ? (
          <ArtifactCard
            item={item}
            key={item.id}
            runtime={runtime}
            sessionId={sessionId}
            contentRenderers={contentRenderers}
          />
        ) : (
          <ActivityRow item={item} key={item.id} />
        ),
      )}
    </div>
  );
}

function MessageRow({
  item,
}: {
  item: Extract<AgentTimelineItem, { kind: "message" }>;
}) {
  const user = item.role === "user";
  const text = useAgentWorkspaceText();
  return (
    <article className={user ? "ml-auto max-w-[85%]" : "max-w-full"}>
      <div className="mb-1 text-xs text-muted-foreground">
        {user ? text.you : item.role === "assistant" ? text.agent : text.system}
      </div>
      <div
        className={
          user
            ? "rounded-md bg-muted px-3 py-2 text-sm whitespace-pre-wrap break-words"
            : ""
        }
      >
        {user ? item.text : <MarkdownMessage text={item.text} />}
        {item.status === "streaming" && (
          <span className="ml-1 inline-block h-4 w-1.5 animate-pulse bg-current align-text-bottom" />
        )}
      </div>
    </article>
  );
}

function ActivityRow({
  item,
}: {
  item: Exclude<AgentTimelineItem, { kind: "artifact" | "message" }>;
}) {
  const text = useAgentWorkspaceText();
  const Icon =
    item.status === "running"
      ? Loader2
      : item.status === "failed" || item.kind === "error"
        ? AlertTriangle
          : CheckCircle2;
  return (
    <article className="border-l-2 border-border px-3 py-2">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Icon className={item.status === "running" ? "size-4 animate-spin" : "size-4"} />
        <span>{item.title}</span>
        <span className="ml-auto text-xs font-normal text-muted-foreground">
          {text.activityStatus(item.status)}
        </span>
      </div>
      {item.detail && (
        <pre className="mt-2 overflow-x-auto whitespace-pre-wrap text-xs text-muted-foreground">
          {item.detail}
        </pre>
      )}
    </article>
  );
}

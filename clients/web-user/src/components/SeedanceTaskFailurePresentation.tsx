import {
  type AgentSessionSnapshot,
  UserVideoTaskPresentation,
  userVideoFailureState,
} from "@do-worker/agent-ui";

import type {
  Bubble,
  RenderItem,
} from "@/lib/renderItems";

const SEEDANCE_AGENT_ID = "seedance-expert";

export interface SeedanceTaskFailureProjection {
  bubbles: Bubble[];
  snapshot: AgentSessionSnapshot;
}

interface SeedanceTaskFailureInput {
  agentId: string | null | undefined;
  agentLabel: string | null | undefined;
  bubbles: Bubble[];
  sessionId: string | null;
}

export function projectSeedanceTaskFailure(
  input: SeedanceTaskFailureInput,
): SeedanceTaskFailureProjection | null {
  if (input.agentId !== SEEDANCE_AGENT_ID || input.sessionId === null) return null;
  const userIndex = input.bubbles.findLastIndex((bubble) => bubble.kind === "user");
  const failure = currentTaskFailure(input.bubbles, userIndex);
  if (failure === null) return null;

  const bubbles = input.bubbles.map((bubble, index) =>
    index === failure.bubbleIndex ? sanitizeFailureBubble(bubble) : bubble,
  );
  const user = userIndex >= 0 ? input.bubbles[userIndex] : undefined;
  return {
    bubbles,
    snapshot: {
      agentLabel: input.agentLabel ?? "Seedance Expert",
      capabilities: {
        interrupt: false,
        resolvePermission: false,
        sendMessage: false,
        terminal: false,
        updateConfiguration: false,
      },
      connection: "connected",
      error: failure.message,
      hasOlderItems: false,
      interactionMode: "acp",
      items: [],
      latestUserCommandId: user?.kind === "user" ? user.itemId : undefined,
      permissions: [],
      plan: [],
      sessionId: input.sessionId,
      status: "failed",
      terminals: [],
      title: input.agentLabel ?? "Seedance Expert",
    },
  };
}

export function SeedanceTaskFailurePresentation({
  projection,
}: {
  projection: SeedanceTaskFailureProjection | null;
}) {
  if (projection === null) return null;
  return (
    <UserVideoTaskPresentation
      artifacts={[]}
      locale="zh-CN"
      snapshot={projection.snapshot}
    />
  );
}

function currentTaskFailure(
  bubbles: readonly Bubble[],
  afterIndex: number,
): { bubbleIndex: number; message: string } | null {
  for (let index = bubbles.length - 1; index > afterIndex; index -= 1) {
    const bubble = bubbles[index];
    if (bubble?.kind !== "assistant") continue;
    if (isVideoFailure(bubble.error)) return { bubbleIndex: index, message: bubble.error };
    for (const item of bubble.items) {
      const message = failureMessage(item);
      if (message !== null) return { bubbleIndex: index, message };
    }
  }
  return null;
}

function sanitizeFailureBubble(bubble: Bubble): Bubble {
  if (bubble.kind !== "assistant") return bubble;
  const failure = isVideoFailure(bubble.error) || bubble.items.some(isVideoFailureItem);
  if (!failure) return bubble;
  return {
    ...bubble,
    error: null,
    items: bubble.items
      .filter((item) => !isVideoFailureItem(item))
      .map((item) =>
        item.kind === "tool" && isVideoFailure(item.output)
          ? { ...item, output: null, state: "no-output" }
          : item,
      ),
    lifecycle: bubble.lifecycle === "failed" ? "completed" : bubble.lifecycle,
  };
}

function isVideoFailureItem(item: RenderItem): boolean {
  return failureMessage(item) !== null;
}

function failureMessage(item: RenderItem): string | null {
  if (item.kind === "text" && isVideoFailure(item.text)) return item.text;
  if (item.kind === "error" && isVideoFailure(item.message)) return item.message;
  if (item.kind === "tool" && isVideoFailure(item.output)) return item.output;
  return null;
}

function isVideoFailure(value: string | null): value is string {
  return value !== null && userVideoFailureState(value) !== null;
}

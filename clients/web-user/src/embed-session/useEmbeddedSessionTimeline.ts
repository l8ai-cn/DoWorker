import { useCallback, useEffect, useState, type Dispatch, type SetStateAction } from "react";

import type { ActiveResponse } from "@/store/types";
import type { AnyBlock, UserMessageBlock } from "@/lib/blocks";
import { BlockStream } from "@/lib/blockStream";
import { itemsToBlocks } from "@/lib/itemsToBlocks";
import { parseSseStream } from "@/lib/sse";
import type { StreamEvent } from "@/lib/events";
import type { SessionStatus } from "@/lib/types";
import type { EmbedSessionClient, EmbeddedSession } from "@/embed-session-api";

interface EmbeddedTimelineState {
  blocks: AnyBlock[];
  error: Error | null;
  isLoading: boolean;
  isSending: boolean;
  session: EmbeddedSession | null;
  status: SessionStatus;
  activeResponse: ActiveResponse | null;
}

const initialState: EmbeddedTimelineState = {
  blocks: [],
  error: null,
  isLoading: true,
  isSending: false,
  session: null,
  status: "idle",
  activeResponse: null,
};

export function useEmbeddedSessionTimeline(client: EmbedSessionClient): {
  state: EmbeddedTimelineState;
  sendMessage: (text: string) => Promise<void>;
} {
  const [state, setState] = useState(initialState);

  useEffect(() => {
    const controller = new AbortController();
    setState(initialState);
    void hydrate(client, controller, setState);
    return () => controller.abort();
  }, [client]);

  const sendMessage = useCallback(
    async (rawText: string) => {
      const text = rawText.trim();
      if (!text || !client.sendMessage) return;
      const tempId = `embedded-pending-${crypto.randomUUID()}`;
      const optimistic = userBlock(tempId, text);
      setState((current) => ({
        ...current,
        blocks: [...current.blocks, optimistic],
        isSending: true,
        error: null,
      }));
      try {
        const result = await client.sendMessage(text);
        setState((current) => ({
          ...current,
          blocks: current.blocks.map((block) =>
            block.ctx.itemId === tempId && result.itemId
              ? { ...block, ctx: { ...block.ctx, itemId: result.itemId } }
              : block,
          ),
          isSending: false,
        }));
      } catch (error) {
        setState((current) => ({
          ...current,
          blocks: current.blocks.filter((block) => block.ctx.itemId !== tempId),
          error: asError(error),
          isSending: false,
        }));
      }
    },
    [client],
  );

  return { state, sendMessage };
}

async function hydrate(
  client: EmbedSessionClient,
  controller: AbortController,
  setState: Dispatch<SetStateAction<EmbeddedTimelineState>>,
): Promise<void> {
  void consumeStream(client, controller, setState);
  try {
    const [session, page] = await Promise.all([client.getSession(), client.getItems()]);
    if (controller.signal.aborted) return;
    setState((current) => ({
      ...current,
      blocks: mergeBlocks(itemsToBlocks(page.items), current.blocks),
      isLoading: false,
      session,
      status: session.status,
    }));
  } catch (error) {
    if (!controller.signal.aborted) {
      setState((current) => ({ ...current, error: asError(error), isLoading: false }));
    }
  }
}

async function consumeStream(
  client: EmbedSessionClient,
  controller: AbortController,
  setState: Dispatch<SetStateAction<EmbeddedTimelineState>>,
): Promise<void> {
  while (!controller.signal.aborted) {
    try {
      const response = await client.openStream(controller.signal);
      if (!response.ok || !response.body) {
        throw new Error(`Embedded session stream failed (${response.status})`);
      }
      const reducer = new BlockStream();
      for await (const block of reducer.reduce(trackSessionEvents(parseSseStream(response.body), setState))) {
        if (controller.signal.aborted) return;
        setState((current) => applyBlock(current, block));
      }
    } catch (error) {
      if (controller.signal.aborted) return;
      setState((current) => ({ ...current, error: asError(error) }));
    }
    await pause(500, controller.signal);
  }
}

async function* trackSessionEvents(
  events: AsyncIterable<StreamEvent>,
  setState: Dispatch<SetStateAction<EmbeddedTimelineState>>,
): AsyncIterable<StreamEvent> {
  for await (const event of events) {
    if (event.type === "session_status") {
      setState((current) => ({ ...current, status: event.status }));
    }
    yield event;
  }
}

function applyBlock(state: EmbeddedTimelineState, block: AnyBlock): EmbeddedTimelineState {
  const blocks = mergeBlocks(state.blocks, [block]);
  if (block.type === "response_start") {
    return {
      ...state,
      blocks,
      status: "running",
      activeResponse: { responseId: block.responseId, state: "streaming", error: null },
    };
  }
  if (block.type === "response_end") {
    return { ...state, blocks, status: "idle", activeResponse: null };
  }
  return { ...state, blocks };
}

function mergeBlocks(first: AnyBlock[], second: AnyBlock[]): AnyBlock[] {
  const seen = new Set<string>();
  return [...first, ...second].filter((block) => {
    const id = block.ctx.itemId;
    if (!id || id.startsWith("embedded-pending-")) return true;
    if (seen.has(id)) return false;
    seen.add(id);
    return true;
  });
}

function userBlock(itemId: string, text: string): UserMessageBlock {
  return {
    type: "user_message",
    ctx: { agent: null, depth: 0, turn: 0, timestamp: 0, responseId: "", itemId },
    content: [{ type: "input_text", text }],
  };
}

function pause(ms: number, signal: AbortSignal): Promise<void> {
  return new Promise((resolve) => {
    const timer = window.setTimeout(resolve, ms);
    signal.addEventListener(
      "abort",
      () => {
        window.clearTimeout(timer);
        resolve();
      },
      { once: true },
    );
  });
}

function asError(value: unknown): Error {
  return value instanceof Error ? value : new Error(String(value));
}

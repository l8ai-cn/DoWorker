import type { EmbedSessionClient } from "@/embed-session-api";
import {
  optimisticUserBlock,
  type EmbeddedRuntimeState,
} from "./embeddedRuntimeState";

export async function sendEmbeddedMessage(
  client: EmbedSessionClient,
  state: EmbeddedRuntimeState,
  commandId: string,
  text: string,
  update: (
    transform: (state: EmbeddedRuntimeState) => EmbeddedRuntimeState,
  ) => void,
): Promise<void> {
  if (!client.sendMessage) throw new Error("Embedded session is read-only");
  const pending = beginEmbeddedMessage(state, commandId, text);
  update(() => pending.state);
  try {
    const result = await client.sendMessage(text);
    update((current) =>
      confirmEmbeddedMessage(current, pending.pendingItemId, result.itemId),
    );
  } catch (cause) {
    update((current) => removeEmbeddedMessage(current, pending.pendingItemId));
    throw cause;
  }
}

function beginEmbeddedMessage(
  state: EmbeddedRuntimeState,
  commandId: string,
  text: string,
) {
  const block = optimisticUserBlock(commandId, text);
  return {
    pendingItemId: block.ctx.itemId,
    state: {
      ...state,
      blocks: [...state.blocks, block],
      error: null,
    },
  };
}

function confirmEmbeddedMessage(
  state: EmbeddedRuntimeState,
  pendingItemId: string | null,
  itemId: string | null,
): EmbeddedRuntimeState {
  if (!itemId) return state;
  return {
    ...state,
    blocks: state.blocks.map((block) =>
      block.ctx.itemId === pendingItemId
        ? { ...block, ctx: { ...block.ctx, itemId } }
        : block,
    ),
  };
}

function removeEmbeddedMessage(
  state: EmbeddedRuntimeState,
  pendingItemId: string | null,
): EmbeddedRuntimeState {
  return {
    ...state,
    blocks: state.blocks.filter((block) => block.ctx.itemId !== pendingItemId),
  };
}

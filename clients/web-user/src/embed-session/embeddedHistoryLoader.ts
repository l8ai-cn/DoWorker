import type { EmbedSessionClient } from "@/embed-session-api";
import { itemsToBlocks } from "@/lib/itemsToBlocks";
import {
  mergeBlocks,
  type EmbeddedRuntimeState,
} from "./embeddedRuntimeState";

export async function loadEmbeddedOlderItems(
  client: EmbedSessionClient,
  state: EmbeddedRuntimeState,
  beforeItemId?: string,
): Promise<EmbeddedRuntimeState> {
  const cursor = beforeItemId ?? oldestItemId(state);
  const page = await client.getItems(cursor);
  return {
    ...state,
    blocks: mergeBlocks(itemsToBlocks(page.items), state.blocks),
    hasOlderItems: page.hasMore,
  };
}

function oldestItemId(state: EmbeddedRuntimeState): string | undefined {
  return state.blocks.find((block) => block.ctx.itemId)?.ctx.itemId ?? undefined;
}

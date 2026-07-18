import type {
  EmbeddedItemsPage,
  EmbeddedSession,
  EmbedSessionClient,
} from "@/embed-session-api";
import { itemsToBlocks } from "@/lib/itemsToBlocks";
import type { AgentArtifactItem, TerminalResource } from "@do-worker/agent-ui";
import {
  mergeBlocks,
  type EmbeddedRuntimeState,
} from "./embeddedRuntimeState";

export interface EmbeddedRuntimeHydration {
  page: EmbeddedItemsPage;
  resourceError: string | null;
  session: EmbeddedSession;
  terminals: TerminalResource[];
  workspaceArtifacts: AgentArtifactItem[];
}

export async function loadEmbeddedRuntimeHydration(
  client: EmbedSessionClient,
): Promise<EmbeddedRuntimeHydration> {
  const [session, page, terminals, workspaceArtifacts] = await Promise.all([
    client.getSession(),
    client.getItems(),
    loadRuntimeResource(client.getTerminals, "终端暂不可用"),
    loadRuntimeResource(
      client.listWorkspaceArtifacts,
      "工作区成果暂不可用",
    ),
  ]);
  return {
    page,
    resourceError: [terminals.error, workspaceArtifacts.error]
      .filter((error): error is string => error !== null)
      .join(" · ") || null,
    session,
    terminals: terminals.items,
    workspaceArtifacts: workspaceArtifacts.items,
  };
}

export function applyEmbeddedRuntimeHydration(
  state: EmbeddedRuntimeState,
  hydration: EmbeddedRuntimeHydration,
  preserveStatus = false,
): EmbeddedRuntimeState {
  const durableBlocks = itemsToBlocks(hydration.page.items);
  return {
    ...state,
    blocks: mergeBlocks(
      durableBlocks,
      removeReconciledStreamedText(state.blocks, durableBlocks),
    ),
    error: hydration.resourceError,
    hasOlderItems: hydration.page.hasMore,
    session: hydration.session,
    status: preserveStatus ? state.status : hydration.session.status,
    terminals: hydration.terminals,
    workspaceArtifacts: hydration.workspaceArtifacts,
  };
}

async function loadRuntimeResource<T>(
  load: (() => Promise<T[]>) | undefined,
  errorLabel: string,
): Promise<{ error: string | null; items: T[] }> {
  if (!load) return { error: null, items: [] };
  try {
    return { error: null, items: await load() };
  } catch (cause) {
    const message = cause instanceof Error ? cause.message : String(cause);
    return { error: `${errorLabel}：${message}`, items: [] };
  }
}

function removeReconciledStreamedText(
  streamedBlocks: EmbeddedRuntimeState["blocks"],
  durableBlocks: EmbeddedRuntimeState["blocks"],
): EmbeddedRuntimeState["blocks"] {
  const durableItemIds = new Set(
    durableBlocks
      .map((block) => block.ctx.itemId)
      .filter((itemId): itemId is string => Boolean(itemId)),
  );
  const durableText = new Set(
    durableBlocks
      .filter((block) => block.type === "text_done")
      .map((block) => textIdentity(block.ctx.responseId, block.fullText)),
  );
  const reconciled: EmbeddedRuntimeState["blocks"] = [];
  let index = 0;
  while (index < streamedBlocks.length) {
    if (!isTextBlock(streamedBlocks[index]!)) {
      reconciled.push(streamedBlocks[index]!);
      index += 1;
      continue;
    }
    const start = index;
    while (index < streamedBlocks.length && isTextBlock(streamedBlocks[index]!)) {
      index += 1;
    }
    reconciled.push(
      ...unreconciledTextRun(
        streamedBlocks.slice(start, index),
        durableItemIds,
        durableText,
      ),
    );
  }
  return reconciled;
}

function unreconciledTextRun(
  run: EmbeddedRuntimeState["blocks"],
  durableItemIds: ReadonlySet<string>,
  durableText: ReadonlySet<string>,
): EmbeddedRuntimeState["blocks"] {
  const kept: EmbeddedRuntimeState["blocks"] = [];
  let segmentStart = 0;
  run.forEach((block, index) => {
    if (block.type !== "text_done") return;
    const matched =
      (block.ctx.itemId !== null && durableItemIds.has(block.ctx.itemId)) ||
      durableText.has(textIdentity(block.ctx.responseId, block.fullText));
    if (!matched) kept.push(...run.slice(segmentStart, index + 1));
    segmentStart = index + 1;
  });
  const trailing = run.slice(segmentStart);
  if (trailing.length === 0) return kept;
  const text = trailing
    .filter((block) => block.type === "text_chunk")
    .map((block) => block.text)
    .join("");
  if (!durableText.has(textIdentity(trailing[0]!.ctx.responseId, text))) {
    kept.push(...trailing);
  }
  return kept;
}

function isTextBlock(
  block: EmbeddedRuntimeState["blocks"][number],
): boolean {
  return block.type === "text_chunk" || block.type === "text_done";
}

function textIdentity(responseId: string, text: string): string {
  return `${responseId}\u0000${text}`;
}

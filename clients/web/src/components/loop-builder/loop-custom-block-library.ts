import { createBlockOp } from "@/lib/blockstore/opBuilder";
import { randomUUID } from "@/lib/blockstore/uuid";
import { blockstoreApi } from "@/lib/api/facade/blockstoreApi";
import {
  BLOCK_TYPE_TYPEDEF,
  type ApplyOpsRequest,
  type Block,
  type Workspace,
} from "@/lib/viewModels/blockstore";
import {
  customBlockDefinitionFromRecord,
  customBlockDefinitionRecord,
  LOOP_CUSTOM_BLOCK_RECORD_KEY,
} from "./loop-custom-block-library-record";
import {
  nextCustomBlockVersion,
  type LoopCustomBlockDefinition,
} from "./loop-custom-block-types";

export interface LoopCustomBlockLibrary {
  ensureDefaultWorkspace(): Promise<Workspace>;
  listTypeDefs(workspaceID: string): Promise<{ blocks: Block[] }>;
  applyOps(request: ApplyOpsRequest): Promise<unknown>;
}

export interface LoadedLoopCustomBlocks {
  definitions: LoopCustomBlockDefinition[];
  workspace: Workspace;
}

export async function loadLoopCustomBlocks(
  library: LoopCustomBlockLibrary = blockstoreApi,
): Promise<LoadedLoopCustomBlocks> {
  const workspace = await library.ensureDefaultWorkspace();
  const { blocks } = await library.listTypeDefs(workspace.id);
  return { definitions: definitionsFromBlocks(blocks), workspace };
}

export async function createLoopCustomBlock(
  definition: LoopCustomBlockDefinition,
  library: LoopCustomBlockLibrary = blockstoreApi,
): Promise<LoadedLoopCustomBlocks> {
  const current = await loadLoopCustomBlocks(library);
  const expectedVersion = nextCustomBlockVersion(current.definitions, definition.slug);
  if (definition.version !== expectedVersion) {
    throw new Error(`custom block version is stale: expected ${expectedVersion}`);
  }
  await library.applyOps({
    workspace_id: current.workspace.id,
    idempotency_key: randomUUID(),
    ops: [createBlockOp({
      type: BLOCK_TYPE_TYPEDEF,
      data: customBlockDefinitionRecord(definition),
    })],
  });
  return loadLoopCustomBlocks(library);
}

export function definitionsFromBlocks(
  blocks: readonly Block[],
): LoopCustomBlockDefinition[] {
  const definitions = new Map<string, LoopCustomBlockDefinition>();
  for (const block of blocks) {
    if (block.type !== BLOCK_TYPE_TYPEDEF) continue;
    const definition = customBlockDefinitionFromRecord(block.data);
    if (!definition) {
      if (Object.hasOwn(block.data, LOOP_CUSTOM_BLOCK_RECORD_KEY)) {
        throw new Error(`invalid custom block definition: ${block.id}`);
      }
      continue;
    }
    const key = `${definition.slug}@${definition.version}`;
    if (definitions.has(key)) throw new Error(`duplicate custom block version: ${key}`);
    definitions.set(key, definition);
  }
  return [...definitions.values()].sort(
    (left, right) => left.slug.localeCompare(right.slug) || left.version - right.version,
  );
}

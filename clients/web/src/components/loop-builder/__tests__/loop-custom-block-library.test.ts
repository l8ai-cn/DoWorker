import { describe, expect, it, vi } from "vitest";
import {
  createLoopCustomBlock,
  definitionsFromBlocks,
  loadLoopCustomBlocks,
  type LoopCustomBlockLibrary,
} from "../loop-custom-block-library";
import { customBlockDefinitionRecord } from "../loop-custom-block-library-record";
import type { LoopCustomBlockDefinition } from "../loop-custom-block-types";
import { BLOCK_TYPE_TYPEDEF, type Block, type Workspace } from "@/lib/viewModels/blockstore";

const workspace: Workspace = {
  id: "workspace-1",
  organization_id: 1,
  slug: "default",
  name: "Default Workspace",
  root_block_id: "root-1",
  created_at: "2026-07-20T00:00:00Z",
};

function definition(version = 1): LoopCustomBlockDefinition {
  return {
    slug: "ppt-step",
    version,
    label: "专业 PPT",
    parameters: ["topic", "file"],
    expansion: {
      agentLocalId: "ppt-step-task",
      verifierLocalId: "ppt-step-check",
      promptTemplate: "制作 {{topic}} 的专业 PPT",
      commandTemplate: "test -f {{file}}",
      acceptTemplate: "{{file}} 存在且可打开",
    },
  };
}

function block(record: LoopCustomBlockDefinition): Block {
  return {
    id: `${record.slug}-${record.version}`,
    workspace_id: workspace.id,
    type: BLOCK_TYPE_TYPEDEF,
    data: customBlockDefinitionRecord(record),
    meta: {},
    created_by: 1,
    created_at: workspace.created_at,
    updated_at: workspace.created_at,
  };
}

function library(blocks: Block[]): LoopCustomBlockLibrary {
  return {
    ensureDefaultWorkspace: vi.fn(async () => workspace),
    listTypeDefs: vi.fn(async () => ({ blocks })),
    applyOps: vi.fn(async (request) => {
      const payload = request.ops[0].payload;
      blocks.push({
        id: String(payload.id),
        workspace_id: workspace.id,
        type: String(payload.type),
        data: payload.data as Block["data"],
        meta: {},
        created_by: 1,
        created_at: workspace.created_at,
        updated_at: workspace.created_at,
      });
    }),
  };
}

describe("Loop custom block library", () => {
  it("loads only schema-checked custom definitions from the organization workspace", async () => {
    const invalid = {
      ...block(definition()),
      data: { type_key: "loop_custom_ppt_step", revision: 1 },
    };
    const loaded = await loadLoopCustomBlocks(library([block(definition()), invalid]));

    expect(loaded.workspace).toEqual(workspace);
    expect(loaded.definitions).toEqual([expect.objectContaining({
      ...definition(),
      definitionId: "ppt-step-1",
      definitionDigest: expect.stringMatching(/^[a-f0-9]{64}$/),
    })]);
  });

  it("appends the next version as an audited Blockstore type definition", async () => {
    const blocks = [block(definition())];
    const store = library(blocks);
    const created = await createLoopCustomBlock(definition(2), store);

    expect(store.applyOps).toHaveBeenCalledWith(expect.objectContaining({
      workspace_id: workspace.id,
      idempotency_key: expect.any(String),
      ops: [expect.objectContaining({
        op: "createBlock",
        payload: expect.objectContaining({
          type: BLOCK_TYPE_TYPEDEF,
          data: expect.objectContaining({
            type_key: "loop_custom_ppt_step",
            revision: 2,
          }),
        }),
      })],
    }));
    expect(created.definitions).toEqual([
      expect.objectContaining({ ...definition(), definitionId: "ppt-step-1" }),
      expect.objectContaining({ ...definition(2), definitionDigest: expect.stringMatching(/^[a-f0-9]{64}$/) }),
    ]);
  });

  it("fails closed when a version is duplicated or stale", async () => {
    await expect(definitionsFromBlocks([block(definition()), block(definition())]))
      .rejects.toThrow("duplicate custom block version: ppt-step@1");
    await expect(definitionsFromBlocks([{
      ...block(definition()),
      data: { loop_custom_definition: { schema_version: 1 } },
    }])).rejects.toThrow("invalid custom block definition: ppt-step-1");
    await expect(createLoopCustomBlock(definition(1), library([block(definition())])))
      .rejects.toThrow("custom block version is stale: expected 2");
  });
});

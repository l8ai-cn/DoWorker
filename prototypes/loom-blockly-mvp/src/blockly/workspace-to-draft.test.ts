import * as Blockly from "blockly";
import { beforeAll, describe, expect, it } from "vitest";

import { registerLoomBlocks } from "./block-catalog";
import { workspaceToDraft } from "./workspace-to-draft";
import {
  createCustomBlockDefinition,
  customBlockType,
  registerCustomBlock,
} from "../custom-blocks/custom-block-definition";

function connectValue(parent: Blockly.Block, input: string, child: Blockly.Block) {
  parent.getInput(input)!.connection!.connect(child.outputConnection!);
}

function connectStatement(parent: Blockly.Block, input: string, child: Blockly.Block) {
  parent.getInput(input)!.connection!.connect(child.previousConnection!);
}

function createCompleteWorkspace(): Blockly.Workspace {
  const workspace = new Blockly.Workspace();
  const root = workspace.newBlock("loom_goal_loop");
  const worker = workspace.newBlock("loom_worker");
  const task = workspace.newBlock("loom_instruction");
  const acceptance = workspace.newBlock("loom_acceptance");
  const verifier = workspace.newBlock("loom_verifier");
  const limits = workspace.newBlock("loom_limits");
  const escalation = workspace.newBlock("loom_escalation");

  root.setFieldValue("发布购物车修复", "NAME");
  worker.setFieldValue("42", "SNAPSHOT_ID");
  worker.setFieldValue("Codex", "LABEL");
  task.setFieldValue("修复税额计算", "TEXT");
  acceptance.setFieldValue("购物车测试通过", "TEXT");
  verifier.setFieldValue("pnpm test", "COMMAND");
  connectValue(root, "WORKER", worker);
  connectStatement(root, "INSTRUCTIONS", task);
  connectStatement(root, "ACCEPTANCE", acceptance);
  connectValue(root, "VERIFIER", verifier);
  connectValue(root, "LIMITS", limits);
  connectValue(root, "ESCALATION", escalation);
  return workspace;
}

beforeAll(registerLoomBlocks);

describe("workspaceToDraft", () => {
  it("maps a complete connected program", () => {
    const workspace = createCompleteWorkspace();
    const draft = workspaceToDraft(workspace);

    expect(draft).toMatchObject({
      name: "发布购物车修复",
      worker: { value: { snapshotId: 42, label: "Codex" } },
      instructions: [{ value: "修复税额计算" }],
      acceptanceCriteria: [{ value: "购物车测试通过" }],
      verification: { value: "pnpm test" },
      limits: {
        value: {
          maxIterations: 10,
          tokenBudget: 80000,
          timeoutMinutes: 60,
          noProgressLimit: 3,
          sameErrorLimit: 2,
        },
      },
      escalationPolicy: { value: "pause" },
      looseBlockIds: [],
      unknownBlockTypes: [],
      adapterDiagnostics: [],
    });
  });

  it("reports every disconnected block", () => {
    const workspace = createCompleteWorkspace();
    const loose = workspace.newBlock("loom_instruction");

    expect(workspaceToDraft(workspace).looseBlockIds).toContain(loose.id);
  });

  it("reports unknown block types instead of ignoring them", () => {
    Blockly.common.defineBlocksWithJsonArray([{
      type: "loom_unknown_test",
      message0: "unknown",
      previousStatement: "LoomInstruction",
      nextStatement: "LoomInstruction",
    }]);
    const workspace = createCompleteWorkspace();
    const unknown = workspace.newBlock("loom_unknown_test");

    expect(workspaceToDraft(workspace).unknownBlockTypes).toContainEqual({
      blockId: unknown.id,
      type: "loom_unknown_test",
    });
  });

  it("expands a connected custom macro into an instruction", () => {
    const definition = createCustomBlockDefinition({
      id: "fix-file",
      name: "修复文件",
      template: "修复 {{file-path}} 并运行 {{test-command}}",
    }).definition!;
    registerCustomBlock(definition);
    const workspace = createCompleteWorkspace();
    const root = workspace.getBlocksByType("loom_goal_loop", false)[0];
    const original = root.getInputTargetBlock("INSTRUCTIONS")!;
    original.dispose(false);
    const custom = workspace.newBlock(customBlockType(definition.id));
    custom.setFieldValue("src/cart.ts", "file-path");
    custom.setFieldValue("pnpm test", "test-command");
    connectStatement(root, "INSTRUCTIONS", custom);

    expect(workspaceToDraft(workspace, [definition]).instructions).toEqual([
      {
        blockId: custom.id,
        value: "修复 src/cart.ts 并运行 pnpm test",
      },
    ]);
  });

  it("reports missing custom macro parameters", () => {
    const definition = createCustomBlockDefinition({
      id: "fix-file-required",
      name: "修复文件",
      template: "修复 {{file-path}}",
    }).definition!;
    registerCustomBlock(definition);
    const workspace = createCompleteWorkspace();
    const root = workspace.getBlocksByType("loom_goal_loop", false)[0];
    root.getInputTargetBlock("INSTRUCTIONS")!.dispose(false);
    const custom = workspace.newBlock(customBlockType(definition.id));
    custom.setFieldValue(" ", "file-path");
    connectStatement(root, "INSTRUCTIONS", custom);

    expect(workspaceToDraft(workspace, [definition]).adapterDiagnostics).toEqual([
      {
        code: "missing-custom-parameter",
        message: "自定义积木参数 file-path 不能为空。",
        blockId: custom.id,
        slot: "file-path",
      },
    ]);
  });
});

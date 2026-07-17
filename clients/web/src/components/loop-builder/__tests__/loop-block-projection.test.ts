import { create } from "@bufbuild/protobuf";
import * as Blockly from "blockly";
import { describe, expect, it } from "vitest";
import {
  LoopProgramSchema,
  type LoopProgram,
} from "@proto/goalloop/v1/goalloop_pb";
import enMessages from "@/messages/en/app.json";
import zhMessages from "@/messages/zh/app.json";
import {
  createLoopBlockCatalog,
  LOOP_BLOCK_TYPES,
  registerLoopBlocks,
} from "../loop-block-catalog";
import {
  findBlockByNodeId,
  projectProgramToWorkspace,
  workspaceToLoopSource,
} from "../loop-block-projection";

function program(): LoopProgram {
  return create(LoopProgramSchema, {
    schemaVersion: 1,
    loop: { nodeId: "n-checkout-fix", localId: "checkout-fix" },
    limits: {
      iterations: 5n,
      tokens: 80000n,
      timeoutMinutes: 60n,
      noProgress: 3n,
      sameError: 2n,
    },
    repeat: {
      identity: { nodeId: "n-fix-cycle", localId: "fix-cycle" },
      max: 5n,
      until: { localId: "tests", field: "passed" },
      agent: {
        identity: { nodeId: "n-fix-tax", localId: "fix-tax" },
        prompt: "修复结算页税额计算，并补充边界测试。",
      },
      verifier: {
        identity: { nodeId: "n-tests", localId: "tests" },
        command: "pnpm test --filter billing",
        accept: "完整测试集通过",
      },
    },
    failurePolicy: "pause",
  });
}

describe("Loop Blockly projection", () => {
  it("round-trips a typed AST to canonical LoopScript", () => {
    registerLoopBlocks(zhMessages.loopWorkbench.blockly);
    const workspace = new Blockly.Workspace();

    projectProgramToWorkspace(workspace, program());
    const repeat = findBlockByNodeId(workspace, "n-fix-cycle");

    expect(workspaceToLoopSource(workspace).source).toBe(`@id(n-checkout-fix)
loop checkout-fix {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax { prompt """修复结算页税额计算，并补充边界测试。""" }
    @id(n-tests)
    verify tests { command "pnpm test --filter billing" accept "完整测试集通过" }
  }
  on_failure pause
}`);
    expect(repeat?.nextConnection).toBeNull();
  });

  it("keeps labels separate from block types, source, and node ids", () => {
    const project = (
      messages: typeof enMessages.loopWorkbench.blockly,
    ) => {
      registerLoopBlocks(messages);
      const workspace = new Blockly.Workspace();
      projectProgramToWorkspace(workspace, program());
      const root = workspace.getBlocksByType(LOOP_BLOCK_TYPES.loop, false)[0];
      const result = workspaceToLoopSource(workspace);
      return {
        label: root.toString(),
        nodeIds: [...result.nodeIndex.keys()].sort(),
        source: result.source,
        types: workspace.getAllBlocks(false).map(({ type }) => type).sort(),
      };
    };

    const english = project(enMessages.loopWorkbench.blockly);
    const chinese = project(zhMessages.loopWorkbench.blockly);

    expect(english.label).toContain("Loop");
    expect(chinese.label).toContain("循环");
    expect(english.source).toBe(chinese.source);
    expect(english.types).toEqual(chinese.types);
    expect(english.nodeIds).toEqual(chinese.nodeIds);
  });

  it.each([
    ["English", enMessages.loopWorkbench.blockly],
    ["Chinese", zhMessages.loopWorkbench.blockly],
  ])("does not expose Worker in the %s toolbox", (_, messages) => {
    const { toolbox } = createLoopBlockCatalog(messages);
    const serialized = JSON.stringify(toolbox).toLowerCase();

    expect(serialized).not.toContain("worker");
    expect(serialized).not.toContain("loop_worker");
  });

  it("uses localized starter text for newly inserted semantic blocks", () => {
    registerLoopBlocks(enMessages.loopWorkbench.blockly);
    const workspace = new Blockly.Workspace();
    const agent = workspace.newBlock(LOOP_BLOCK_TYPES.agent);
    const verifier = workspace.newBlock(LOOP_BLOCK_TYPES.verifier);

    expect(agent.getFieldValue("PROMPT")).toBe("Describe the task to complete");
    expect(verifier.getFieldValue("ACCEPT")).toBe("Verification passes");
  });

  it("preserves semantic node ids after block edits", () => {
    registerLoopBlocks(zhMessages.loopWorkbench.blockly);
    const workspace = new Blockly.Workspace();
    projectProgramToWorkspace(workspace, program());

    const agent = findBlockByNodeId(workspace, "n-fix-tax");
    agent?.setFieldValue("只修复税额舍入", "PROMPT");

    const result = workspaceToLoopSource(workspace);
    expect(result.source).toContain('prompt """只修复税额舍入"""');
    expect(result.nodeIndex.get("n-fix-tax")).toBe(agent?.id);
  });

  it("does not silently accept an incomplete block tree", () => {
    registerLoopBlocks(zhMessages.loopWorkbench.blockly);
    const workspace = new Blockly.Workspace();
    projectProgramToWorkspace(workspace, program());
    findBlockByNodeId(workspace, "n-tests")?.dispose(false);

    const result = workspaceToLoopSource(workspace);

    expect(result.complete).toBe(false);
    expect(result.source).not.toContain("verify tests");
    expect(result.source).toContain("invalid-block-structure");
  });

  it("does not silently discard extra connected steps", () => {
    registerLoopBlocks(zhMessages.loopWorkbench.blockly);
    const workspace = new Blockly.Workspace();
    projectProgramToWorkspace(workspace, program());
    const agent = findBlockByNodeId(workspace, "n-fix-tax");
    const verifier = findBlockByNodeId(workspace, "n-tests");
    const extra = workspace.newBlock("loop_agent");
    verifier?.previousConnection?.disconnect();
    agent?.nextConnection?.setCheck(null);
    extra.previousConnection?.setCheck(null);
    extra.nextConnection?.setCheck(null);
    verifier?.previousConnection?.setCheck(null);
    if (agent?.nextConnection && extra.previousConnection) {
      agent.nextConnection.connect(extra.previousConnection);
    }
    if (extra.nextConnection && verifier?.previousConnection) {
      extra.nextConnection.connect(verifier.previousConnection);
    }

    const result = workspaceToLoopSource(workspace);

    expect(result.complete).toBe(false);
    expect(result.issues).toContain(
      "重复执行必须且只能按顺序包含一个智能体任务和一个验证步骤",
    );
    expect(result.source).toContain("invalid-block-structure");
  });
});

import { create } from "@bufbuild/protobuf";
import * as Blockly from "blockly";
import { describe, expect, it } from "vitest";
import {
  LoopProgramSchema,
  type LoopProgram,
} from "@proto/goalloop/v1/goalloop_pb";
import { registerLoopBlocks } from "../loop-block-catalog";
import {
  findBlockByNodeId,
  projectProgramToWorkspace,
  workspaceToLoopSource,
} from "../loop-block-projection";
import { loopToolbox } from "../loop-block-catalog";

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
    registerLoopBlocks();
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

  it("does not expose Worker as a programmable block", () => {
    expect(JSON.stringify(loopToolbox)).not.toContain("Worker");
    expect(JSON.stringify(loopToolbox)).not.toContain("loop_worker");
  });

  it("preserves semantic node ids after block edits", () => {
    registerLoopBlocks();
    const workspace = new Blockly.Workspace();
    projectProgramToWorkspace(workspace, program());

    const agent = findBlockByNodeId(workspace, "n-fix-tax");
    agent?.setFieldValue("只修复税额舍入", "PROMPT");

    const result = workspaceToLoopSource(workspace);
    expect(result.source).toContain('prompt """只修复税额舍入"""');
    expect(result.nodeIndex.get("n-fix-tax")).toBe(agent?.id);
  });

  it("does not silently accept an incomplete block tree", () => {
    registerLoopBlocks();
    const workspace = new Blockly.Workspace();
    projectProgramToWorkspace(workspace, program());
    findBlockByNodeId(workspace, "n-tests")?.dispose(false);

    const result = workspaceToLoopSource(workspace);

    expect(result.complete).toBe(false);
    expect(result.source).not.toContain("verify tests");
    expect(result.source).toContain("invalid-block-structure");
  });

  it("does not silently discard extra connected steps", () => {
    registerLoopBlocks();
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

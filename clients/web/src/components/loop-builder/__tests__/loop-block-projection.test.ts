import { create } from "@bufbuild/protobuf";
import * as Blockly from "blockly";
import { describe, expect, it } from "vitest";
import {
  LoopProgramSchema,
  type LoopProgram,
} from "@proto/goalloop/v1/goalloop_pb";
import deMessages from "@/messages/de/app.json";
import enMessages from "@/messages/en/app.json";
import esMessages from "@/messages/es/app.json";
import frMessages from "@/messages/fr/app.json";
import jaMessages from "@/messages/ja/app.json";
import koMessages from "@/messages/ko/app.json";
import ptMessages from "@/messages/pt/app.json";
import zhMessages from "@/messages/zh/app.json";
import {
  createLoopBlockCatalog,
  LOOP_BLOCK_TYPES,
  registerLoopBlocks,
} from "../loop-block-catalog";
import { insertLoopBlock } from "../loop-block-insertion";
import { loopBlockProgrammingHostAdapter } from "../loop-block-programming-host-adapter";
import {
  customBlockType,
  type LoopResolvedCustomBlockDefinition,
} from "../loop-custom-block-types";
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

const CUSTOM_BLOCK_DIGEST = "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0";

function pptCustomBlock(): LoopResolvedCustomBlockDefinition {
  return {
    definitionId: "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4",
    definitionDigest: CUSTOM_BLOCK_DIGEST,
    slug: "ppt-step",
    version: 1,
    label: "制作 PPT",
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

function customProgram(): LoopProgram {
  return create(LoopProgramSchema, {
    ...program(),
    repeat: {
      identity: { nodeId: "n-build-cycle", localId: "build-cycle" },
      max: 5n,
      until: { localId: "ppt-step-check", field: "passed" },
      agent: {
        identity: {
          nodeId: "n-ppt-step-ppt-step-task",
          localId: "ppt-step-task",
        },
        prompt: "制作 季度复盘 的专业 PPT",
      },
      verifier: {
        identity: {
          nodeId: "n-ppt-step-ppt-step-check",
          localId: "ppt-step-check",
        },
        command: "test -f output.pptx",
        accept: "output.pptx 存在且可打开",
      },
      customBlock: {
        nodeId: "n-ppt-step",
        definitionId: "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4",
        slug: "ppt-step",
        version: 1,
        definitionDigest: CUSTOM_BLOCK_DIGEST,
      },
    },
  });
}

function sourceFor(messages: typeof enMessages.loopWorkbench.blockly): string {
  registerLoopBlocks(messages);
  const workspace = new Blockly.Workspace();
  projectProgramToWorkspace(workspace, program());
  return workspaceToLoopSource(workspace).source;
}

const LOCALE_BLOCK_MESSAGES = [
  ["English", enMessages.loopWorkbench.blockly],
  ["Chinese", zhMessages.loopWorkbench.blockly],
  ["German", deMessages.loopWorkbench.blockly],
  ["Spanish", esMessages.loopWorkbench.blockly],
  ["French", frMessages.loopWorkbench.blockly],
  ["Japanese", jaMessages.loopWorkbench.blockly],
  ["Korean", koMessages.loopWorkbench.blockly],
  ["Portuguese", ptMessages.loopWorkbench.blockly],
] as const;

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

  it.each(LOCALE_BLOCK_MESSAGES)("does not expose Worker in the %s toolbox", (_, messages) => {
    const { toolbox } = createLoopBlockCatalog(messages);
    const serialized = JSON.stringify(toolbox).toLowerCase();

    expect(serialized).not.toContain("worker");
    expect(serialized).not.toContain("loop_worker");
  });

  it.each(LOCALE_BLOCK_MESSAGES)("keeps %s projection semantically identical", (_, messages) => {
    const expected = sourceFor(zhMessages.loopWorkbench.blockly);
    registerLoopBlocks(messages);
    const workspace = new Blockly.Workspace();
    projectProgramToWorkspace(workspace, program());
    const result = workspaceToLoopSource(workspace);

    expect(result.source).toBe(expected);
    expect([...result.nodeIndex.keys()].sort()).toEqual([
      "n-checkout-fix",
      "n-fix-cycle",
      "n-fix-tax",
      "n-tests",
    ]);
  });

  it.each([
    ["English", enMessages.loopWorkbench.blockly, "Describe the task to complete", "Verification passes"],
    ["Chinese", zhMessages.loopWorkbench.blockly, "描述要完成的任务", "验证通过"],
    ["German", deMessages.loopWorkbench.blockly, "Beschreiben Sie die zu erledigende Aufgabe", "Verifizierung erfolgreich"],
    ["Spanish", esMessages.loopWorkbench.blockly, "Describe la tarea a completar", "La verificación pasa"],
    ["French", frMessages.loopWorkbench.blockly, "Décrivez la tâche à accomplir", "La vérification réussit"],
    ["Japanese", jaMessages.loopWorkbench.blockly, "完了するタスクを説明してください", "検証が成功"],
    ["Korean", koMessages.loopWorkbench.blockly, "완료할 작업을 설명하세요", "검증 통과"],
    ["Portuguese", ptMessages.loopWorkbench.blockly, "Descreva a tarefa a concluir", "Verificação aprovada"],
  ])("uses %s starter text for newly inserted semantic blocks", (_, messages, prompt, accept) => {
    registerLoopBlocks(messages);
    const workspace = new Blockly.Workspace();
    const agent = workspace.newBlock(LOOP_BLOCK_TYPES.agent);
    const verifier = workspace.newBlock(LOOP_BLOCK_TYPES.verifier);

    expect(agent.getFieldValue("PROMPT")).toBe(prompt);
    expect(verifier.getFieldValue("ACCEPT")).toBe(accept);
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

  it("round-trips a versioned custom block through standard agent and verifier source", () => {
    const definition = pptCustomBlock();
    registerLoopBlocks(zhMessages.loopWorkbench.blockly, [definition]);
    const workspace = new Blockly.Workspace();

    projectProgramToWorkspace(workspace, customProgram(), [definition]);
    const custom = workspace.getBlocksByType(customBlockType(definition), false)[0];
    const result = workspaceToLoopSource(workspace, [definition]);

    expect(custom).toBeDefined();
    expect(workspace.getBlocksByType(LOOP_BLOCK_TYPES.agent, false)).toHaveLength(0);
    expect(result.complete).toBe(true);
    expect(result.nodeIndex.get("n-ppt-step-ppt-step-task")).toBe(custom.id);
    expect(result.source).toContain(
      `custom_block(node_id: n-ppt-step, definition_id: "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4", slug: ppt-step, version: 1, digest: "${CUSTOM_BLOCK_DIGEST}")`,
    );
    expect(result.source).toContain(
      'agent ppt-step-task { prompt """制作 季度复盘 的专业 PPT""" }',
    );
    expect(result.source).toContain(
      'verify ppt-step-check { command "test -f output.pptx" accept "output.pptx 存在且可打开" }',
    );
  });

  it("registers historical definitions but only exposes the latest revision for insertion", () => {
    const v1 = pptCustomBlock();
    const v2 = {
      ...v1,
      definitionId: "c4da5391-1f35-4c9f-9340-f6d3746174b3",
      definitionDigest: "b1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0",
      version: 2,
      label: "制作 PPT v2",
    };
    const catalog = createLoopBlockCatalog(zhMessages.loopWorkbench.blockly, [v1, v2]);

    expect(catalog.definitions.map(({ type }) => type)).toEqual(
      expect.arrayContaining([customBlockType(v1), customBlockType(v2)]),
    );
    expect(JSON.stringify(catalog.toolbox)).toContain(customBlockType(v2));
    expect(JSON.stringify(catalog.toolbox)).not.toContain(customBlockType(v1));
  });

  it("exposes Loop through the reusable host adapter contract", () => {
    const definition = pptCustomBlock();
    const workspace = new Blockly.Workspace();

    loopBlockProgrammingHostAdapter.registerBlocks(
      zhMessages.loopWorkbench.blockly,
      [definition],
    );
    loopBlockProgrammingHostAdapter.projectProgram(workspace, customProgram(), [definition]);

    const catalog = loopBlockProgrammingHostAdapter.createCatalog(
      zhMessages.loopWorkbench.blockly,
      [definition],
    );
    const result = loopBlockProgrammingHostAdapter.workspaceToSource(workspace, [definition]);

    expect(loopBlockProgrammingHostAdapter.namespace).toBe("loop");
    expect(JSON.stringify(catalog.toolbox).toLowerCase()).not.toContain("worker");
    expect(result.complete).toBe(true);
    expect(result.source).toContain("agent ppt-step-task");
  });

  it("inserts custom blocks into the repeat body as standard LoopScript steps", () => {
    const definition = pptCustomBlock();
    const workspace = new Blockly.Workspace();

    registerLoopBlocks(zhMessages.loopWorkbench.blockly, [definition]);
    projectProgramToWorkspace(workspace, program(), [definition]);
    insertLoopBlock({
      workspace,
      type: customBlockType(definition),
      customDefinitions: [definition],
      insertPoint: { menuX: 0, menuY: 0, workspaceX: 0, workspaceY: 0 },
    });

    const result = workspaceToLoopSource(workspace, [definition]);

    expect(result.complete).toBe(true);
    expect(result.source.toLowerCase()).not.toContain("worker");
    expect(result.source).toContain(
      'agent ppt-step-task { prompt """制作 topic 的专业 PPT""" }',
    );
    expect(result.source).toContain("custom_block(node_id:");
    expect(result.source).toContain(
      'verify ppt-step-check { command "test -f file" accept "file 存在且可打开" }',
    );
  });

  it("fails closed when a custom block pin does not match the loaded definition", () => {
    const definition = pptCustomBlock();
    const mismatched = { ...definition, definitionDigest: "c1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0" };
    const workspace = new Blockly.Workspace();

    registerLoopBlocks(zhMessages.loopWorkbench.blockly, [mismatched]);

    expect(() => projectProgramToWorkspace(workspace, customProgram(), [mismatched]))
      .toThrow("custom block definition pin cannot be resolved");
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

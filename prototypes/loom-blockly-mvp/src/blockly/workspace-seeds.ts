import * as Blockly from "blockly";

import { LOOM_BLOCK_TYPES } from "./block-catalog";

function block(
  workspace: Blockly.WorkspaceSvg,
  type: string,
): Blockly.BlockSvg {
  const created = workspace.newBlock(type);
  created.initSvg();
  created.render();
  return created;
}

function connectValue(
  parent: Blockly.BlockSvg,
  input: string,
  child: Blockly.BlockSvg,
): void {
  parent.getInput(input)!.connection!.connect(child.outputConnection!);
}

function connectStatement(
  parent: Blockly.BlockSvg,
  input: string,
  child: Blockly.BlockSvg,
): void {
  parent.getInput(input)!.connection!.connect(child.previousConnection!);
}

export function insertLoomBlock(
  workspace: Blockly.WorkspaceSvg,
  type: string,
  x: number,
  y: number,
): Blockly.BlockSvg {
  const created = block(workspace, type);
  created.moveBy(x, y);
  created.select();
  return created;
}

export function createGoalRoot(
  workspace: Blockly.WorkspaceSvg,
): Blockly.BlockSvg {
  const existing = workspace.getBlocksByType(
    LOOM_BLOCK_TYPES.root,
    false,
  )[0] as Blockly.BlockSvg | undefined;
  if (existing) {
    existing.select();
    workspace.centerOnBlock(existing.id);
    return existing;
  }
  return insertLoomBlock(workspace, LOOM_BLOCK_TYPES.root, 72, 52);
}

export function loadExampleProgram(
  workspace: Blockly.WorkspaceSvg,
): void {
  workspace.clear();
  const root = block(workspace, LOOM_BLOCK_TYPES.root);
  const worker = block(workspace, LOOM_BLOCK_TYPES.worker);
  const task = block(workspace, LOOM_BLOCK_TYPES.instruction);
  const acceptance = block(workspace, LOOM_BLOCK_TYPES.acceptance);
  const verifier = block(workspace, LOOM_BLOCK_TYPES.verifier);
  const limits = block(workspace, LOOM_BLOCK_TYPES.limits);
  const escalation = block(workspace, LOOM_BLOCK_TYPES.escalation);

  root.setFieldValue("结算页修复", "NAME");
  task.setFieldValue("修复税额计算并补充边界测试", "TEXT");
  acceptance.setFieldValue("完整测试集通过", "TEXT");
  connectValue(root, "WORKER", worker);
  connectStatement(root, "INSTRUCTIONS", task);
  connectStatement(root, "ACCEPTANCE", acceptance);
  connectValue(root, "VERIFIER", verifier);
  connectValue(root, "LIMITS", limits);
  connectValue(root, "ESCALATION", escalation);
  root.moveBy(80, 36);
  root.select();
  workspace.centerOnBlock(root.id);
}

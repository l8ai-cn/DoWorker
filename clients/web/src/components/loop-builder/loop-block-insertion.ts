import * as Blockly from "blockly";
import { LOOP_BLOCK_TYPES } from "./loop-block-catalog";
import type { LoopBlockInsertPoint } from "./loop-block-insert-point";
import {
  customBlockType,
  type LoopCustomBlockDefinition,
} from "./loop-custom-block-types";

interface InsertLoopBlockInput {
  workspace: Blockly.Workspace;
  type: string;
  insertPoint: LoopBlockInsertPoint;
  customDefinitions: readonly LoopCustomBlockDefinition[];
}

export function insertLoopBlock({
  workspace,
  type,
  insertPoint,
  customDefinitions,
}: InsertLoopBlockInput) {
  const created = workspace.newBlock(type);
  if (created instanceof Blockly.BlockSvg) created.initSvg();
  const definition = customDefinitions.find((item) => customBlockType(item) === type);
  if (definition && connectCustomBlock(workspace, created, definition)) {
    renderAndSelect(created);
    return;
  }
  renderAndSelect(created);
  if (created instanceof Blockly.BlockSvg) {
    created.moveBy(insertPoint.workspaceX, insertPoint.workspaceY);
  }
}

function renderAndSelect(block: Blockly.Block) {
  if (!(block instanceof Blockly.BlockSvg)) return;
  block.render();
  block.select();
}

function connectCustomBlock(
  workspace: Blockly.Workspace,
  block: Blockly.Block,
  definition: LoopCustomBlockDefinition,
) {
  const repeat = workspace.getBlocksByType(LOOP_BLOCK_TYPES.repeat, false)[0];
  const body = repeat?.getInput("BODY")?.connection;
  if (!body || !block.previousConnection) return false;
  body.targetBlock()?.dispose(false);
  body.connect(block.previousConnection);
  repeat.setFieldValue(definition.expansion.verifierLocalId, "UNTIL_ID");
  repeat.setFieldValue("passed", "UNTIL_FIELD");
  return true;
}

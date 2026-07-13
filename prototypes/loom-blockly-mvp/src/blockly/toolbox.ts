import type * as Blockly from "blockly";

import { LOOM_BLOCK_TYPES } from "./block-catalog";
import {
  customBlockType,
  type CustomBlockDefinition,
} from "../custom-blocks/custom-block-definition";

function block(type: string): Blockly.utils.toolbox.BlockInfo {
  return { kind: "block", type };
}

export function createLoomToolbox(
  customDefinitions: CustomBlockDefinition[],
): Blockly.utils.toolbox.ToolboxDefinition {
  return {
    kind: "categoryToolbox",
    contents: [
      { kind: "category", name: "控制", colour: "#2f4858", contents: [block(LOOM_BLOCK_TYPES.root)] },
      { kind: "category", name: "Worker", colour: "#2d6a4f", contents: [block(LOOM_BLOCK_TYPES.worker)] },
      { kind: "category", name: "任务", colour: "#34699a", contents: [block(LOOM_BLOCK_TYPES.instruction)] },
      { kind: "category", name: "验收", colour: "#7b5e2e", contents: [block(LOOM_BLOCK_TYPES.acceptance)] },
      { kind: "category", name: "验证", colour: "#6b4c7a", contents: [block(LOOM_BLOCK_TYPES.verifier)] },
      { kind: "category", name: "边界", colour: "#8a4f3d", contents: [block(LOOM_BLOCK_TYPES.limits)] },
      { kind: "category", name: "升级", colour: "#8b3a3a", contents: [block(LOOM_BLOCK_TYPES.escalation)] },
      {
        kind: "category",
        name: "我的积木",
        colour: "#4f5d75",
        contents: customDefinitions.map(({ id }) => block(customBlockType(id))),
      },
    ],
  };
}

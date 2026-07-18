import type * as Blockly from "blockly";

export function setLoopBlockAccess(
  workspace: Blockly.WorkspaceSvg,
  readOnly: boolean,
) {
  for (const block of workspace.getAllBlocks(false)) {
    block.setEditable(!readOnly);
    block.setMovable(!readOnly);
    block.setDeletable(!readOnly);
  }
}

import type * as Blockly from "blockly";

export function setWorkspaceEditing(
  workspace: Blockly.Workspace,
  enabled: boolean,
): void {
  workspace.setIsReadOnly(!enabled);
  for (const block of workspace.getAllBlocks(false)) {
    block.setEditable(enabled);
    block.setMovable(enabled);
    block.setDeletable(enabled);
  }
  const svgWorkspace = workspace as Blockly.WorkspaceSvg;
  svgWorkspace.getToolbox?.()?.setVisible(enabled);
}

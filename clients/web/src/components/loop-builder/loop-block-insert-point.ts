import * as Blockly from "blockly";
import type { MouseEvent } from "react";

export interface LoopBlockInsertPoint {
  menuX: number;
  menuY: number;
  workspaceX: number;
  workspaceY: number;
}

export function insertPointFromDoubleClick(
  event: MouseEvent<HTMLDivElement>,
  workspace: Blockly.WorkspaceSvg | undefined,
  readOnly: boolean,
): LoopBlockInsertPoint | undefined {
  const target = event.target;
  if (readOnly || !workspace || !(target instanceof Element) ||
      !target.classList.contains("blocklyMainBackground")) return undefined;
  const bounds = event.currentTarget.getBoundingClientRect();
  const point = Blockly.utils.svgMath.screenToWsCoordinates(
    workspace,
    new Blockly.utils.Coordinate(event.clientX, event.clientY),
  );
  return {
    menuX: event.clientX - bounds.left,
    menuY: event.clientY - bounds.top,
    workspaceX: point.x,
    workspaceY: point.y,
  };
}

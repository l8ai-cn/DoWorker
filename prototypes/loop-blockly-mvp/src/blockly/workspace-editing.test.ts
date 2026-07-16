import * as Blockly from "blockly";
import { beforeAll, describe, expect, it } from "vitest";

import { registerLoopBlocks } from "./block-catalog";
import { setWorkspaceEditing } from "./workspace-editing";

beforeAll(registerLoopBlocks);

describe("setWorkspaceEditing", () => {
  it("freezes and restores every block", () => {
    const workspace = new Blockly.Workspace();
    const block = workspace.newBlock("loop_goal_loop");

    setWorkspaceEditing(workspace, false);
    expect(workspace.isReadOnly()).toBe(true);
    expect(block.isEditable()).toBe(false);
    expect(block.isMovable()).toBe(false);
    expect(block.isDeletable()).toBe(false);

    setWorkspaceEditing(workspace, true);
    expect(workspace.isReadOnly()).toBe(false);
    expect(block.isEditable()).toBe(true);
    expect(block.isMovable()).toBe(true);
    expect(block.isDeletable()).toBe(true);
  });
});

import * as Blockly from "blockly";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import { LOOP_BLOCK_TYPES, registerLoopBlocks } from "./loop-block-catalog";
import { resolvePinnedCustomBlock } from "./loop-custom-block-expansion";
import {
  customBlockType,
  type LoopResolvedCustomBlockDefinition,
} from "./loop-custom-block-types";
import {
  setBlockCustomBlockReference,
  setBlockNodeId,
} from "./loop-node-identity";

function block(workspace: Blockly.Workspace, type: string): Blockly.Block {
  const created = workspace.newBlock(type);
  if (created instanceof Blockly.BlockSvg) created.initSvg();
  return created;
}

function field(block: Blockly.Block, name: string, value: string | number | bigint): void {
  block.setFieldValue(String(value), name);
}

function value(parent: Blockly.Block, input: string, child: Blockly.Block): void {
  if (child.outputConnection) {
    parent.getInput(input)?.connection?.connect(child.outputConnection);
  }
}

function statement(parent: Blockly.Block, input: string, child: Blockly.Block): void {
  if (child.previousConnection) {
    parent.getInput(input)?.connection?.connect(child.previousConnection);
  }
}

export function projectProgramToWorkspace(
  workspace: Blockly.Workspace,
  program: LoopProgram,
  customDefinitions: readonly LoopResolvedCustomBlockDefinition[] = [],
): Map<string, string> {
  const loop = program.loop;
  const limitsNode = program.limits;
  const repeatNode = program.repeat;
  const agentNode = repeatNode?.agent;
  const verifierNode = repeatNode?.verifier;
  if (!loop || !limitsNode || !repeatNode?.identity ||
      !repeatNode.until || !agentNode?.identity || !verifierNode?.identity) {
    throw new Error("循环语法树不完整，无法投影为积木");
  }

  registerLoopBlocks(undefined, customDefinitions);
  Blockly.Events.disable();
  try {
    workspace.clear();
    const custom = resolvePinnedCustomBlock(program, customDefinitions);
    const root = block(workspace, LOOP_BLOCK_TYPES.loop);
    const limits = block(workspace, LOOP_BLOCK_TYPES.limits);
    const repeat = block(workspace, LOOP_BLOCK_TYPES.repeat);
    const agent = custom ? undefined : block(workspace, LOOP_BLOCK_TYPES.agent);
    const verifier = custom ? undefined : block(workspace, LOOP_BLOCK_TYPES.verifier);
    const customBlock = custom ? block(workspace, customBlockType(custom.definition)) : undefined;
    const failure = block(workspace, LOOP_BLOCK_TYPES.failure);

    field(root, "LOCAL_ID", loop.localId);
    field(limits, "ITERATIONS", limitsNode.iterations);
    field(limits, "TOKENS", limitsNode.tokens);
    field(limits, "TIMEOUT", limitsNode.timeoutMinutes);
    field(limits, "NO_PROGRESS", limitsNode.noProgress);
    field(limits, "SAME_ERROR", limitsNode.sameError);
    field(repeat, "LOCAL_ID", repeatNode.identity.localId);
    field(repeat, "MAX", repeatNode.max);
    field(repeat, "UNTIL_ID", repeatNode.until.localId);
    field(repeat, "UNTIL_FIELD", repeatNode.until.field);
    if (agent && verifier) {
      field(agent, "LOCAL_ID", agentNode.identity.localId);
      field(agent, "PROMPT", agentNode.prompt);
      field(verifier, "LOCAL_ID", verifierNode.identity.localId);
      field(verifier, "COMMAND", verifierNode.command);
      field(verifier, "ACCEPT", verifierNode.accept);
    }
    if (customBlock && custom) {
      for (const [name, value] of Object.entries(custom.values)) {
        field(customBlock, name, value);
      }
    }
    field(failure, "POLICY", program.failurePolicy);

    setBlockNodeId(root, loop.nodeId);
    setBlockNodeId(repeat, repeatNode.identity.nodeId);
    if (agent && verifier) {
      setBlockNodeId(agent, agentNode.identity.nodeId);
      setBlockNodeId(verifier, verifierNode.identity.nodeId);
    }
    if (customBlock && custom && repeatNode.customBlock) {
      setBlockNodeId(customBlock, custom.nodeId);
      setBlockCustomBlockReference(customBlock, repeatNode.customBlock);
    }

    value(root, "LIMITS", limits);
    statement(root, "BODY", repeat);
    if (customBlock) statement(repeat, "BODY", customBlock);
    if (agent) statement(repeat, "BODY", agent);
    if (agent?.nextConnection && verifier?.previousConnection) {
      agent.nextConnection.connect(verifier.previousConnection);
    }
    value(root, "FAILURE", failure);

    for (const projected of [limits, repeat, customBlock, agent, verifier, failure, root]) {
      if (projected instanceof Blockly.BlockSvg) projected.render();
    }
    if (root instanceof Blockly.BlockSvg) {
      root.moveBy(48, 36);
    }
    return new Map([
      [loop.nodeId, root.id],
      [repeatNode.identity.nodeId, repeat.id],
      [agentNode.identity.nodeId, customBlock?.id ?? agent?.id ?? ""],
      [verifierNode.identity.nodeId, customBlock?.id ?? verifier?.id ?? ""],
    ]);
  } finally {
    Blockly.Events.enable();
  }
}

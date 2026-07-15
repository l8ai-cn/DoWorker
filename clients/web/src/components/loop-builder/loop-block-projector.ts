import * as Blockly from "blockly";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import { LOOP_BLOCK_TYPES, registerLoopBlocks } from "./loop-block-catalog";
import { setBlockNodeId } from "./loop-node-identity";

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
): Map<string, string> {
  const loop = program.loop;
  const workerNode = program.worker;
  const limitsNode = program.limits;
  const repeatNode = program.repeat;
  const agentNode = repeatNode?.agent;
  const verifierNode = repeatNode?.verifier;
  if (!loop || !workerNode?.identity || !limitsNode || !repeatNode?.identity ||
      !repeatNode.until || !agentNode?.identity || !verifierNode?.identity) {
    throw new Error("Loop AST is incomplete and cannot be projected");
  }

  registerLoopBlocks();
  Blockly.Events.disable();
  try {
    workspace.clear();
    const root = block(workspace, LOOP_BLOCK_TYPES.loop);
    const worker = block(workspace, LOOP_BLOCK_TYPES.worker);
    const limits = block(workspace, LOOP_BLOCK_TYPES.limits);
    const repeat = block(workspace, LOOP_BLOCK_TYPES.repeat);
    const agent = block(workspace, LOOP_BLOCK_TYPES.agent);
    const verifier = block(workspace, LOOP_BLOCK_TYPES.verifier);
    const failure = block(workspace, LOOP_BLOCK_TYPES.failure);

    field(root, "LOCAL_ID", loop.localId);
    field(worker, "LOCAL_ID", workerNode.identity.localId);
    field(worker, "SNAPSHOT_ID", workerNode.snapshotId);
    field(limits, "ITERATIONS", limitsNode.iterations);
    field(limits, "TOKENS", limitsNode.tokens);
    field(limits, "TIMEOUT", limitsNode.timeoutMinutes);
    field(limits, "NO_PROGRESS", limitsNode.noProgress);
    field(limits, "SAME_ERROR", limitsNode.sameError);
    field(repeat, "LOCAL_ID", repeatNode.identity.localId);
    field(repeat, "MAX", repeatNode.max);
    field(repeat, "UNTIL_ID", repeatNode.until.localId);
    field(repeat, "UNTIL_FIELD", repeatNode.until.field);
    field(agent, "LOCAL_ID", agentNode.identity.localId);
    field(agent, "WORKER_REF", agentNode.workerRef);
    field(agent, "PROMPT", agentNode.prompt);
    field(verifier, "LOCAL_ID", verifierNode.identity.localId);
    field(verifier, "COMMAND", verifierNode.command);
    field(verifier, "ACCEPT", verifierNode.accept);
    field(failure, "POLICY", program.failurePolicy);

    setBlockNodeId(root, loop.nodeId);
    setBlockNodeId(worker, workerNode.identity.nodeId);
    setBlockNodeId(repeat, repeatNode.identity.nodeId);
    setBlockNodeId(agent, agentNode.identity.nodeId);
    setBlockNodeId(verifier, verifierNode.identity.nodeId);

    value(root, "WORKER", worker);
    value(root, "LIMITS", limits);
    statement(root, "BODY", repeat);
    statement(repeat, "BODY", agent);
    if (agent.nextConnection && verifier.previousConnection) {
      agent.nextConnection.connect(verifier.previousConnection);
    }
    value(root, "FAILURE", failure);

    for (const projected of [worker, limits, repeat, agent, verifier, failure, root]) {
      if (projected instanceof Blockly.BlockSvg) projected.render();
    }
    if (root instanceof Blockly.BlockSvg) {
      root.moveBy(48, 36);
    }
    return new Map([
      [loop.nodeId, root.id],
      [workerNode.identity.nodeId, worker.id],
      [repeatNode.identity.nodeId, repeat.id],
      [agentNode.identity.nodeId, agent.id],
      [verifierNode.identity.nodeId, verifier.id],
    ]);
  } finally {
    Blockly.Events.enable();
  }
}

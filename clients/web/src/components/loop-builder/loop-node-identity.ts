import type * as Blockly from "blockly";

interface BlockMetadata {
  nodeId?: string;
}

function metadata(block: Blockly.Block): BlockMetadata {
  if (!block.data) return {};
  try {
    return JSON.parse(block.data) as BlockMetadata;
  } catch {
    return {};
  }
}

function slug(value: string): string {
  const normalized = value
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  return normalized || "node";
}

export function setBlockNodeId(block: Blockly.Block, nodeId: string): void {
  block.data = JSON.stringify({ ...metadata(block), nodeId });
}

export function getBlockNodeId(block: Blockly.Block): string | undefined {
  return metadata(block).nodeId;
}

export function ensureBlockNodeId(block: Blockly.Block): string {
  const existing = getBlockNodeId(block);
  if (existing) return existing;
  const localId = String(block.getFieldValue("LOCAL_ID") ?? block.type);
  const suffix = slug(block.id).slice(-12);
  const nodeId = `n-${slug(localId)}-${suffix}`;
  setBlockNodeId(block, nodeId);
  return nodeId;
}

export function findBlockByNodeId(
  workspace: Blockly.Workspace,
  nodeId: string,
): Blockly.Block | undefined {
  return workspace.getAllBlocks(false).find((block) => getBlockNodeId(block) === nodeId);
}

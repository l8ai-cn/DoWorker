import type * as Blockly from "blockly";
import { LOOP_BLOCK_TYPES } from "./loop-block-catalog";
import { ensureBlockNodeId, getBlockNodeId } from "./loop-node-identity";

export interface LoopBlockSourceResult {
  source: string;
  complete: boolean;
  issues: string[];
  nodeIndex: Map<string, string>;
}

function text(block: Blockly.Block | null, field: string): string {
  return String(block?.getFieldValue(field) ?? "");
}

function number(block: Blockly.Block | null, field: string): number {
  return Number(block?.getFieldValue(field) ?? 0);
}

function prompt(value: string): string {
  if (value.includes('"""') || value.endsWith('"') || value.startsWith(" ") ||
      value.endsWith(" ") || value.includes("\n")) {
    return JSON.stringify(value);
  }
  return `"""${value}"""`;
}

function indexNode(block: Blockly.Block | null, index: Map<string, string>): string {
  if (!block) return "";
  const nodeId = ensureBlockNodeId(block);
  index.set(nodeId, block.id);
  return nodeId;
}

export function workspaceToLoopSource(
  workspace: Blockly.Workspace,
): LoopBlockSourceResult {
  const issues: string[] = [];
  const nodeIndex = new Map<string, string>();
  const roots = workspace.getBlocksByType(LOOP_BLOCK_TYPES.loop, false);
  const root = roots[0] ?? null;
  if (!root) return { source: "", complete: false, issues: ["缺少循环根积木"], nodeIndex };
  if (roots.length > 1) issues.push("一个工作区只能有一个循环根积木");

  const limits = root.getInputTargetBlock("LIMITS");
  const repeat = root.getInputTargetBlock("BODY");
  const failure = root.getInputTargetBlock("FAILURE");
  const body = repeat?.getInputTargetBlock("BODY") ?? null;
  const bodyBlocks: Blockly.Block[] = [];
  for (let current = body; current; current = current.getNextBlock()) bodyBlocks.push(current);
  const hasExactBody = bodyBlocks.length === 2 &&
    bodyBlocks[0].type === LOOP_BLOCK_TYPES.agent &&
    bodyBlocks[1].type === LOOP_BLOCK_TYPES.verifier;
  if (bodyBlocks.length > 0 && !hasExactBody) {
    issues.push("重复执行必须且只能按顺序包含一个智能体任务和一个验证步骤");
  }
  const agent = hasExactBody ? bodyBlocks[0] : null;
  const verifier = hasExactBody ? bodyBlocks[1] : null;

  for (const [block, label] of [
    [limits, "执行边界"], [repeat, "重复执行"],
    [agent, "智能体任务"], [verifier, "验证步骤"], [failure, "失败策略"],
  ] as const) {
    if (!block) issues.push(`缺少 ${label} 积木`);
  }
  const connected = new Set(root.getDescendants(false).map(({ id }) => id));
  const loose = workspace.getAllBlocks(false).filter(({ id }) => !connected.has(id));
  if (loose.length > 0) issues.push(`存在 ${loose.length} 个未连接积木`);

  const lines = [`@id(${indexNode(root, nodeIndex)})`, `loop ${text(root, "LOCAL_ID")} {`];
  if (limits) {
    lines.push(
      `  limits(iterations: ${number(limits, "ITERATIONS")}, tokens: ${number(limits, "TOKENS")}, ` +
      `timeout: ${number(limits, "TIMEOUT")}m, no_progress: ${number(limits, "NO_PROGRESS")}, ` +
      `same_error: ${number(limits, "SAME_ERROR")})`,
    );
  }
  if (repeat) {
    lines.push(`  @id(${indexNode(repeat, nodeIndex)})`);
    lines.push(
      `  repeat ${text(repeat, "LOCAL_ID")}(max: ${number(repeat, "MAX")}, ` +
      `until: ${text(repeat, "UNTIL_ID")}.${text(repeat, "UNTIL_FIELD")}) {`,
    );
    if (agent) {
      lines.push(`    @id(${indexNode(agent, nodeIndex)})`);
      lines.push(
        `    agent ${text(agent, "LOCAL_ID")} { prompt ${prompt(text(agent, "PROMPT"))} }`,
      );
    }
    if (verifier) {
      lines.push(`    @id(${indexNode(verifier, nodeIndex)})`);
      lines.push(
        `    verify ${text(verifier, "LOCAL_ID")} { command ${JSON.stringify(text(verifier, "COMMAND"))} ` +
        `accept ${JSON.stringify(text(verifier, "ACCEPT"))} }`,
      );
    }
    lines.push("  }");
  }
  if (failure) lines.push(`  on_failure ${text(failure, "POLICY")}`);
  if (issues.length > 0) lines.push("  invalid-block-structure");
  lines.push("}");
  return { source: lines.join("\n"), complete: issues.length === 0, issues, nodeIndex };
}

export function nodeIdForBlock(block: Blockly.Block): string | undefined {
  return getBlockNodeId(block);
}

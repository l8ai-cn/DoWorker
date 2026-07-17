import type * as Blockly from "blockly";

import { LOOP_BLOCK_TYPES } from "./block-catalog";
import {
  customBlockType,
  expandCustomBlockTemplate,
  type CustomBlockDefinition,
} from "../custom-blocks/custom-block-definition";
import type { Diagnostic, LoopDraft, SourceBlock } from "../domain/loop-types";

const BUILT_IN_TYPES = new Set(Object.values(LOOP_BLOCK_TYPES));

function text(block: Blockly.Block, field: string): string {
  return String(block.getFieldValue(field) ?? "");
}

function number(block: Blockly.Block, field: string): number {
  return Number(block.getFieldValue(field));
}

function sourceText(block: Blockly.Block, field: string): SourceBlock<string> {
  return { blockId: block.id, value: text(block, field) };
}

function readInstructionChain(
  first: Blockly.Block | null,
  customByType: Map<string, CustomBlockDefinition>,
): { diagnostics: Diagnostic[]; instructions: SourceBlock<string>[] } {
  const instructions: SourceBlock<string>[] = [];
  const diagnostics: Diagnostic[] = [];
  for (let block = first; block; block = block.getNextBlock()) {
    if (block.type === LOOP_BLOCK_TYPES.instruction) {
      instructions.push(sourceText(block, "TEXT"));
      continue;
    }
    const definition = customByType.get(block.type);
    if (!definition) continue;
    const values = Object.fromEntries(
      definition.parameters.map((parameter) => [parameter, text(block, parameter)]),
    );
    const expansion = expandCustomBlockTemplate(definition, values);
    for (const parameter of expansion.missingParameters) {
      diagnostics.push({
        code: "missing-custom-parameter",
        message: `自定义积木参数 ${parameter} 不能为空。`,
        blockId: block.id,
        slot: parameter,
      });
    }
    instructions.push({
      blockId: block.id,
      value: expansion.value ?? "",
    });
  }
  return { diagnostics, instructions };
}

function readAcceptanceChain(
  first: Blockly.Block | null,
): SourceBlock<string>[] {
  const criteria: SourceBlock<string>[] = [];
  for (let block = first; block; block = block.getNextBlock()) {
    if (block.type === LOOP_BLOCK_TYPES.acceptance) {
      criteria.push(sourceText(block, "TEXT"));
    }
  }
  return criteria;
}

export function workspaceToDraft(
  workspace: Blockly.Workspace,
  customDefinitions: CustomBlockDefinition[] = [],
): LoopDraft {
  const root = workspace.getBlocksByType(LOOP_BLOCK_TYPES.root, false)[0];
  const allBlocks = workspace.getAllBlocks(false);
  const adapterDiagnostics: Diagnostic[] = [];
  const customByType = new Map<string, CustomBlockDefinition>();
  for (const definition of customDefinitions) {
    const type = customBlockType(definition.id);
    if (customByType.has(type)) {
      adapterDiagnostics.push({
        code: "duplicate-custom-definition",
        message: `自定义积木 ID 重复：${definition.id}。`,
      });
      continue;
    }
    customByType.set(type, definition);
  }
  const connectedIds = new Set(root?.getDescendants(false).map(({ id }) => id) ?? []);
  const customTypes = new Set(
    customDefinitions.map(({ id }) => customBlockType(id)),
  );
  const unknownBlockTypes = allBlocks
    .filter(({ type }) => !BUILT_IN_TYPES.has(type as never) && !customTypes.has(type))
    .map(({ id, type }) => ({ blockId: id, type }));
  const looseBlockIds = allBlocks
    .filter(({ id }) => !connectedIds.has(id))
    .map(({ id }) => id);

  if (!root) {
    return {
      name: "",
      rootBlockId: "workspace",
      instructions: [],
      acceptanceCriteria: [],
      looseBlockIds,
      unknownBlockTypes,
      adapterDiagnostics,
    };
  }

  const worker = root.getInputTargetBlock("WORKER");
  const verifier = root.getInputTargetBlock("VERIFIER");
  const limits = root.getInputTargetBlock("LIMITS");
  const escalation = root.getInputTargetBlock("ESCALATION");
  const instructionResult = readInstructionChain(
    root.getInputTargetBlock("INSTRUCTIONS"),
    customByType,
  );
  return {
    name: text(root, "NAME"),
    rootBlockId: root.id,
    worker: worker ? {
      blockId: worker.id,
      value: {
        snapshotId: number(worker, "SNAPSHOT_ID"),
        label: text(worker, "LABEL"),
      },
    } : undefined,
    instructions: instructionResult.instructions,
    acceptanceCriteria: readAcceptanceChain(
      root.getInputTargetBlock("ACCEPTANCE"),
    ),
    verification: verifier ? sourceText(verifier, "COMMAND") : undefined,
    limits: limits ? {
      blockId: limits.id,
      value: {
        maxIterations: number(limits, "MAX_ITERATIONS"),
        tokenBudget: number(limits, "TOKEN_BUDGET"),
        timeoutMinutes: number(limits, "TIMEOUT_MINUTES"),
        noProgressLimit: number(limits, "NO_PROGRESS_LIMIT"),
        sameErrorLimit: number(limits, "SAME_ERROR_LIMIT"),
      },
    } : undefined,
    escalationPolicy: escalation ? {
      blockId: escalation.id,
      value: text(escalation, "POLICY") as "pause" | "fail",
    } : undefined,
    looseBlockIds,
    unknownBlockTypes,
    adapterDiagnostics: [
      ...adapterDiagnostics,
      ...instructionResult.diagnostics,
    ],
  };
}

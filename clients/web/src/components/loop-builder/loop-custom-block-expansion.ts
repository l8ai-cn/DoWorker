import type * as Blockly from "blockly";
import {
  expandTemplate,
  matchTemplate,
  type LoopResolvedCustomBlockDefinition,
} from "./loop-custom-block-types";
import { referencePinsDefinition } from "./loop-custom-block-definition-digest";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";

export interface CustomBlockExpansion {
  agentLocalId: string;
  verifierLocalId: string;
  prompt: string;
  command: string;
  accept: string;
  issues: string[];
}

export interface CustomBlockMatch {
  definition: LoopResolvedCustomBlockDefinition;
  nodeId: string;
  values: Record<string, string>;
}

export function valuesForCustomBlock(block: Blockly.Block): Record<string, string> {
  const values: Record<string, string> = {};
  for (const input of block.inputList) {
    for (const field of input.fieldRow) {
      const name = field.name;
      if (name) values[name] = String(block.getFieldValue(name) ?? "");
    }
  }
  return values;
}

export function expandCustomBlock(
  definition: LoopResolvedCustomBlockDefinition,
  values: Record<string, string>,
): CustomBlockExpansion {
  const prompt = expandTemplate(definition.expansion.promptTemplate, values);
  const command = expandTemplate(definition.expansion.commandTemplate, values);
  const accept = expandTemplate(definition.expansion.acceptTemplate, values);
  return {
    agentLocalId: definition.expansion.agentLocalId,
    verifierLocalId: definition.expansion.verifierLocalId,
    prompt: prompt.value,
    command: command.value,
    accept: accept.value,
    issues: [...prompt.missing, ...command.missing, ...accept.missing]
      .map((parameter) => `custom block parameter ${parameter} is required`),
  };
}

export function resolvePinnedCustomBlock(
  program: LoopProgram,
  definitions: readonly LoopResolvedCustomBlockDefinition[],
): CustomBlockMatch | undefined {
  const agent = program.repeat?.agent;
  const verifier = program.repeat?.verifier;
  const pin = program.repeat?.customBlock;
  if (!pin) return undefined;
  if (!agent || !verifier) {
    throw new Error("custom block pin requires expanded agent and verifier nodes");
  }
  const definition = definitions.find((candidate) =>
    candidate.definitionId === pin.definitionId &&
    candidate.slug === pin.slug &&
    candidate.version === pin.version,
  );
  if (!definition || !referencePinsDefinition(pin, definition)) {
    throw new Error("custom block definition pin cannot be resolved");
  }
  if (
    agent.identity?.nodeId !== `${pin.nodeId}-${definition.expansion.agentLocalId}` ||
    verifier.identity?.nodeId !== `${pin.nodeId}-${definition.expansion.verifierLocalId}`
  ) {
    throw new Error("custom block pin does not match expanded node identities");
  }
  const prompt = matchTemplate(definition.expansion.promptTemplate, agent.prompt);
  const command = matchTemplate(definition.expansion.commandTemplate, verifier.command);
  const accept = matchTemplate(definition.expansion.acceptTemplate, verifier.accept);
  if (!prompt || !command || !accept) {
    throw new Error("custom block pin does not match the pinned definition");
  }
  const values = { ...prompt, ...command, ...accept };
  if (Object.keys(values).length !== definition.parameters.length) {
    throw new Error("custom block pin has incomplete parameter values");
  }
  return { definition, nodeId: pin.nodeId, values };
}

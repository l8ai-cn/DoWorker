import type * as Blockly from "blockly";
import {
  expandTemplate,
  matchTemplate,
  type LoopCustomBlockDefinition,
} from "./loop-custom-block-types";
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
  definition: LoopCustomBlockDefinition;
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
  definition: LoopCustomBlockDefinition,
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

export function matchCustomBlock(
  program: LoopProgram,
  definitions: readonly LoopCustomBlockDefinition[],
): CustomBlockMatch | undefined {
  const agent = program.repeat?.agent;
  const verifier = program.repeat?.verifier;
  if (!agent || !verifier) return undefined;
  for (const definition of definitions) {
    const agentNodeId = agent.identity?.nodeId ?? "";
    const verifierNodeId = verifier.identity?.nodeId ?? "";
    const agentSuffix = `-${definition.expansion.agentLocalId}`;
    if (!agentNodeId.endsWith(agentSuffix)) continue;
    const nodeId = agentNodeId.slice(0, -agentSuffix.length);
    if (verifierNodeId !== `${nodeId}-${definition.expansion.verifierLocalId}`) continue;
    const prompt = matchTemplate(definition.expansion.promptTemplate, agent.prompt);
    const command = matchTemplate(definition.expansion.commandTemplate, verifier.command);
    const accept = matchTemplate(definition.expansion.acceptTemplate, verifier.accept);
    if (!prompt || !command || !accept) continue;
    const values = { ...prompt, ...command, ...accept };
    if (Object.keys(values).length === definition.parameters.length) {
      return { definition, nodeId, values };
    }
  }
  return undefined;
}

import * as Blockly from "blockly";
import type { LoopBlockCatalogMessages } from "./loop-workbench-messages";

export const LOOP_BLOCK_TYPES = {
  loop: "loop_loop",
  limits: "loop_limits",
  repeat: "loop_repeat",
  agent: "loop_agent",
  verifier: "loop_verifier",
  failure: "loop_failure",
} as const;

function createDefinitions(messages: LoopBlockCatalogMessages) {
  return [
  {
    type: LOOP_BLOCK_TYPES.loop,
    message0: messages.loop.message0,
    args0: [{ type: "field_input", name: "LOCAL_ID", text: "checkout-fix" }],
    message1: messages.loop.message1,
    args1: [{ type: "input_value", name: "LIMITS", check: "LoopLimits" }],
    message2: messages.loop.message2,
    args2: [{ type: "input_statement", name: "BODY", check: "LoopRepeat" }],
    message3: messages.loop.message3,
    args3: [{ type: "input_value", name: "FAILURE", check: "LoopFailure" }],
    colour: 216,
  },
  {
    type: LOOP_BLOCK_TYPES.limits,
    message0: messages.limits.message0,
    args0: [
      { type: "field_number", name: "ITERATIONS", value: 5, min: 1, precision: 1 },
      { type: "field_number", name: "TOKENS", value: 80000, min: 1, precision: 1 },
      { type: "field_number", name: "TIMEOUT", value: 60, min: 1, precision: 1 },
    ],
    message1: messages.limits.message1,
    args1: [
      { type: "field_number", name: "NO_PROGRESS", value: 3, min: 1, precision: 1 },
      { type: "field_number", name: "SAME_ERROR", value: 2, min: 1, precision: 1 },
    ],
    output: "LoopLimits",
    colour: 43,
  },
  {
    type: LOOP_BLOCK_TYPES.repeat,
    message0: messages.repeat.message0,
    args0: [
      { type: "field_input", name: "LOCAL_ID", text: "fix-cycle" },
      { type: "field_number", name: "MAX", value: 5, min: 1, precision: 1 },
    ],
    message1: messages.repeat.message1,
    args1: [
      { type: "field_input", name: "UNTIL_ID", text: "tests" },
      { type: "field_input", name: "UNTIL_FIELD", text: "passed" },
    ],
    message2: messages.repeat.message2,
    args2: [{ type: "input_statement", name: "BODY", check: "LoopAgent" }],
    previousStatement: "LoopRepeat",
    colour: 216,
  },
  {
    type: LOOP_BLOCK_TYPES.agent,
    message0: messages.agent.message0,
    args0: [{ type: "field_input", name: "LOCAL_ID", text: "fix-task" }],
    message1: messages.agent.message1,
    args1: [{ type: "field_input", name: "PROMPT", text: messages.agent.defaultPrompt }],
    previousStatement: "LoopAgent",
    nextStatement: "LoopVerifier",
    colour: 292,
  },
  {
    type: LOOP_BLOCK_TYPES.verifier,
    message0: messages.verifier.message0,
    args0: [{ type: "field_input", name: "LOCAL_ID", text: "tests" }],
    message1: messages.verifier.message1,
    args1: [{ type: "field_input", name: "COMMAND", text: "pnpm test" }],
    message2: messages.verifier.message2,
    args2: [{ type: "field_input", name: "ACCEPT", text: messages.verifier.defaultAccept }],
    previousStatement: "LoopVerifier",
    colour: 122,
  },
  {
    type: LOOP_BLOCK_TYPES.failure,
    message0: messages.failure.message0,
    args0: [{
      type: "field_dropdown",
      name: "POLICY",
      options: [[messages.failure.pause, "pause"], [messages.failure.fail, "fail"]],
    }],
    output: "LoopFailure",
    colour: 8,
  },
  ];
}

export function createLoopBlockCatalog(messages: LoopBlockCatalogMessages) {
  const definitions = createDefinitions(messages);
  const toolbox: Blockly.utils.toolbox.ToolboxDefinition = {
    kind: "categoryToolbox",
    contents: [
      { kind: "category", name: messages.toolbox.loop, colour: "216", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.loop }] },
      { kind: "category", name: messages.toolbox.control, colour: "216", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.repeat }] },
      { kind: "category", name: messages.toolbox.agent, colour: "292", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.agent }] },
      { kind: "category", name: messages.toolbox.verifier, colour: "122", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.verifier }] },
      { kind: "category", name: messages.toolbox.limits, colour: "43", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.limits }] },
      { kind: "category", name: messages.toolbox.failure, colour: "8", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.failure }] },
    ],
  };
  return { definitions, toolbox };
}

export function registerLoopBlocks(messages?: LoopBlockCatalogMessages): void {
  if (messages) {
    for (const type of Object.values(LOOP_BLOCK_TYPES)) delete Blockly.Blocks[type];
    Blockly.common.defineBlocksWithJsonArray(createDefinitions(messages));
    return;
  }
  const missing = Object.values(LOOP_BLOCK_TYPES).filter((type) => !Blockly.Blocks[type]);
  if (missing.length > 0) {
    throw new Error("Loop block messages must be registered before projection");
  }
}

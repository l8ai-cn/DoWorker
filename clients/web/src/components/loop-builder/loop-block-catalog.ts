import * as Blockly from "blockly";

export const LOOP_BLOCK_TYPES = {
  loop: "loop_loop",
  limits: "loop_limits",
  repeat: "loop_repeat",
  agent: "loop_agent",
  verifier: "loop_verifier",
  failure: "loop_failure",
} as const;

const definitions = [
  {
    type: LOOP_BLOCK_TYPES.loop,
    message0: "循环 %1",
    args0: [{ type: "field_input", name: "LOCAL_ID", text: "checkout-fix" }],
    message1: "执行边界 %1",
    args1: [{ type: "input_value", name: "LIMITS", check: "LoopLimits" }],
    message2: "循环步骤 %1",
    args2: [{ type: "input_statement", name: "BODY", check: "LoopRepeat" }],
    message3: "失败处理 %1",
    args3: [{ type: "input_value", name: "FAILURE", check: "LoopFailure" }],
    colour: 216,
  },
  {
    type: LOOP_BLOCK_TYPES.limits,
    message0: "最多 %1 轮 · %2 令牌 · %3 分钟",
    args0: [
      { type: "field_number", name: "ITERATIONS", value: 5, min: 1, precision: 1 },
      { type: "field_number", name: "TOKENS", value: 80000, min: 1, precision: 1 },
      { type: "field_number", name: "TIMEOUT", value: 60, min: 1, precision: 1 },
    ],
    message1: "无进展 %1 次 · 同错 %2 次",
    args1: [
      { type: "field_number", name: "NO_PROGRESS", value: 3, min: 1, precision: 1 },
      { type: "field_number", name: "SAME_ERROR", value: 2, min: 1, precision: 1 },
    ],
    output: "LoopLimits",
    colour: 43,
  },
  {
    type: LOOP_BLOCK_TYPES.repeat,
    message0: "重复执行 %1 · 最多 %2 次",
    args0: [
      { type: "field_input", name: "LOCAL_ID", text: "fix-cycle" },
      { type: "field_number", name: "MAX", value: 5, min: 1, precision: 1 },
    ],
    message1: "直到 %1 . %2",
    args1: [
      { type: "field_input", name: "UNTIL_ID", text: "tests" },
      { type: "field_input", name: "UNTIL_FIELD", text: "passed" },
    ],
    message2: "执行 %1",
    args2: [{ type: "input_statement", name: "BODY", check: "LoopAgent" }],
    previousStatement: "LoopRepeat",
    colour: 216,
  },
  {
    type: LOOP_BLOCK_TYPES.agent,
    message0: "智能体任务 %1",
    args0: [{ type: "field_input", name: "LOCAL_ID", text: "fix-task" }],
    message1: "任务说明 %1",
    args1: [{ type: "field_input", name: "PROMPT", text: "描述要完成的任务" }],
    previousStatement: "LoopAgent",
    nextStatement: "LoopVerifier",
    colour: 292,
  },
  {
    type: LOOP_BLOCK_TYPES.verifier,
    message0: "验证步骤 %1",
    args0: [{ type: "field_input", name: "LOCAL_ID", text: "tests" }],
    message1: "命令 %1",
    args1: [{ type: "field_input", name: "COMMAND", text: "pnpm test" }],
    message2: "通过条件 %1",
    args2: [{ type: "field_input", name: "ACCEPT", text: "测试通过" }],
    previousStatement: "LoopVerifier",
    colour: 122,
  },
  {
    type: LOOP_BLOCK_TYPES.failure,
    message0: "失败后 %1",
    args0: [{
      type: "field_dropdown",
      name: "POLICY",
      options: [["暂停并等待人工", "pause"], ["标记失败", "fail"]],
    }],
    output: "LoopFailure",
    colour: 8,
  },
];

export function registerLoopBlocks(): void {
  const missing = definitions.filter(({ type }) => !Blockly.Blocks[type]);
  if (missing.length > 0) Blockly.common.defineBlocksWithJsonArray(missing);
}

export const loopToolbox: Blockly.utils.toolbox.ToolboxDefinition = {
  kind: "categoryToolbox",
  contents: [
    { kind: "category", name: "循环", colour: "216", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.loop }] },
    { kind: "category", name: "控制", colour: "216", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.repeat }] },
    { kind: "category", name: "智能体", colour: "292", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.agent }] },
    { kind: "category", name: "验证", colour: "122", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.verifier }] },
    { kind: "category", name: "边界", colour: "43", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.limits }] },
    { kind: "category", name: "失败", colour: "8", contents: [{ kind: "block", type: LOOP_BLOCK_TYPES.failure }] },
  ],
};

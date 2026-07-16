import * as Blockly from "blockly";

export const LOOP_BLOCK_TYPES = {
  root: "loop_goal_loop",
  worker: "loop_worker",
  instruction: "loop_instruction",
  acceptance: "loop_acceptance",
  verifier: "loop_verifier",
  limits: "loop_limits",
  escalation: "loop_escalation",
} as const;

const DEFINITIONS = [
  {
    type: LOOP_BLOCK_TYPES.root,
    message0: "Goal Loop %1",
    args0: [{ type: "field_input", name: "NAME", text: "新建 Loop" }],
    message1: "Worker %1",
    args1: [{ type: "input_value", name: "WORKER", check: "LoopWorker" }],
    message2: "任务 %1",
    args2: [{ type: "input_statement", name: "INSTRUCTIONS", check: "LoopInstruction" }],
    message3: "验收 %1",
    args3: [{ type: "input_statement", name: "ACCEPTANCE", check: "LoopAcceptance" }],
    message4: "验证 %1",
    args4: [{ type: "input_value", name: "VERIFIER", check: "LoopVerifier" }],
    message5: "边界 %1",
    args5: [{ type: "input_value", name: "LIMITS", check: "LoopLimits" }],
    message6: "失败处理 %1",
    args6: [{ type: "input_value", name: "ESCALATION", check: "LoopEscalation" }],
    style: "loop_control_blocks",
  },
  {
    type: LOOP_BLOCK_TYPES.worker,
    message0: "使用 Worker 快照 %1 名称 %2",
    args0: [
      { type: "field_number", name: "SNAPSHOT_ID", value: 42, min: 1, precision: 1 },
      { type: "field_input", name: "LABEL", text: "Codex" },
    ],
    output: "LoopWorker",
    style: "loop_worker_blocks",
  },
  {
    type: LOOP_BLOCK_TYPES.instruction,
    message0: "执行任务 %1",
    args0: [{ type: "field_input", name: "TEXT", text: "描述要完成的任务" }],
    previousStatement: "LoopInstruction",
    nextStatement: "LoopInstruction",
    style: "loop_task_blocks",
  },
  {
    type: LOOP_BLOCK_TYPES.acceptance,
    message0: "验收条件 %1",
    args0: [{ type: "field_input", name: "TEXT", text: "定义完成证据" }],
    previousStatement: "LoopAcceptance",
    nextStatement: "LoopAcceptance",
    style: "loop_acceptance_blocks",
  },
  {
    type: LOOP_BLOCK_TYPES.verifier,
    message0: "运行验证命令 %1",
    args0: [{ type: "field_input", name: "COMMAND", text: "pnpm test" }],
    output: "LoopVerifier",
    style: "loop_verifier_blocks",
  },
  {
    type: LOOP_BLOCK_TYPES.limits,
    message0: "最多轮数 %1 Token %2 分钟 %3",
    args0: [
      { type: "field_number", name: "MAX_ITERATIONS", value: 10, min: 1, max: 100 },
      { type: "field_number", name: "TOKEN_BUDGET", value: 80000, min: 1 },
      { type: "field_number", name: "TIMEOUT_MINUTES", value: 60, min: 1 },
    ],
    message1: "无进展 %1 次 同错 %2 次",
    args1: [
      { type: "field_number", name: "NO_PROGRESS_LIMIT", value: 3, min: 1 },
      { type: "field_number", name: "SAME_ERROR_LIMIT", value: 2, min: 1 },
    ],
    output: "LoopLimits",
    style: "loop_limit_blocks",
  },
  {
    type: LOOP_BLOCK_TYPES.escalation,
    message0: "失败后 %1",
    args0: [{
      type: "field_dropdown",
      name: "POLICY",
      options: [["暂停并等待人工", "pause"], ["直接失败", "fail"]],
    }],
    output: "LoopEscalation",
    style: "loop_escalation_blocks",
  },
];

export function registerLoopBlocks(): void {
  const missing = DEFINITIONS.filter(({ type }) => !Blockly.Blocks[type]);
  if (missing.length > 0) Blockly.common.defineBlocksWithJsonArray(missing);
}

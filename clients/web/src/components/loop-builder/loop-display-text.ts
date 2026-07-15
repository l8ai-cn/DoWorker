const parseStatusLabels: Record<string, string> = {
  empty: "等待输入",
  parsing: "校验中",
  valid: "有效",
  "syntax-error": "存在错误",
};

const runStatusLabels: Record<string, string> = {
  draft: "草稿",
  active: "运行中",
  paused: "已暂停",
  verifying: "验证中",
  completed: "已完成",
  failed: "失败",
  cancelled: "已取消",
};

const diagnosticLabels: Record<string, string> = {
  "loop.program.nil": "循环程序不能为空",
  "loop.schema-version.unsupported": "循环脚本版本不受支持",
  "loop.syntax.invalid-token": "循环脚本包含无效字符",
  "loop.syntax.invalid-number": "循环脚本包含无效数字",
  "loop.syntax.unexpected-token": "循环脚本结构不符合语法",
  "loop.syntax.unknown": "循环脚本包含未知语法",
  "loop.node-id.missing": "积木节点缺少唯一标识",
  "loop.node-id.duplicate": "积木节点标识重复",
  "loop.local-id.duplicate": "步骤名称重复",
  "loop.identifier.invalid": "步骤名称格式无效",
  "loop.structure.limits-count": "必须且只能设置一个执行边界",
  "loop.structure.repeat-count": "必须且只能设置一个重复执行步骤",
  "loop.structure.agent-count": "重复执行中必须且只能包含一个智能体任务",
  "loop.structure.verifier-count": "重复执行中必须且只能包含一个验证步骤",
  "loop.structure.failure-count": "必须且只能设置一个失败处理策略",
  "loop.reference.until-invalid": "循环停止条件必须引用验证步骤的通过结果",
  "loop.repeat.max-exceeds-limit": "重复次数不能超过全局最大轮数",
  "loop.failure-policy.invalid": "失败处理策略无效",
  "loop.value.out-of-range": "积木参数超出允许范围",
  "loop.text.empty": "积木文本参数不能为空",
  "loop.secret.literal-forbidden": "循环脚本不能直接写入密钥",
};

export function loopParseStatusLabel(status: string): string {
  return parseStatusLabels[status] ?? "未知状态";
}

export function loopRunStatusLabel(status: string): string {
  return runStatusLabels[status] ?? "未知状态";
}

export function loopDiagnosticLabel(code: string): string {
  return diagnosticLabels[code] ?? "循环脚本校验失败";
}

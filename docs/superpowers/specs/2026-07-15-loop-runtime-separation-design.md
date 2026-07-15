# Loop 与运行环境解耦设计

## 目标

Loop 工作台只声明智能体执行过程，不编排 Worker。Worker 是一次运行所使用的
运行环境快照，在启动 Loop 时选择并独立传递。

## 语义边界

- LoopScript 声明循环、智能体任务、验证、预算边界和失败策略。
- Blockly 是 LoopScript AST 的中文可视化投影。
- Worker 快照不属于 LoopScript AST，不参与积木连接，也不影响编译。
- 运行请求同时提交 LoopScript 和 Worker 快照标识；后端在创建 GoalLoop 前校验快照。
- Web 端以十进制字符串保存运行环境快照标识，编码 protobuf 时再转为
  `int64`，禁止经过 JavaScript `number` 造成精度丢失。
- AI 生成结构化 Loop AST 或语义修改，不直接生成 Blockly 工作区 JSON。

## LoopScript V1

```loopscript
@id(n-checkout-fix)
loop checkout-fix {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax { prompt """修复结算页税额计算，并补充边界测试。""" }
    @id(n-tests)
    verify tests { command "pnpm test --filter billing" accept "完整测试集通过" }
  }
  on_failure pause
}
```

删除语法：

- `worker <alias> = snapshot(<id>)`
- `agent <id>(using: <worker>)`

## 运行流程

1. 用户通过积木、代码或 AI 生成 Loop。
2. 后端编译并返回规范源码、AST 和诊断。
3. 用户点击“运行循环”。
4. 弹窗列出当前组织可执行的运行环境快照。
5. 用户确认后提交 `source + worker_spec_snapshot_id`。
6. 后端校验运行环境，创建并启动 GoalLoop。

## 中文界面

- 所有操作、状态、积木类别和积木字段使用中文。
- `LoopScript` 的界面名称为“循环脚本”，语言关键字继续使用英文。
- 标识符和命令保留原始值，不翻译用户代码。
- `valid`、`syntax-error` 等内部状态映射为中文，不直接展示协议值。

## 验收

- 画布和工具箱中不存在 Worker。
- 规范 LoopScript 中不存在 Worker 声明和 `using` 参数。
- 编译 LoopScript 不依赖可用运行环境。
- 未选择运行环境时不能启动，且显示明确中文提示。
- 运行时后端仍校验快照的组织归属、可用性和新鲜度。
- 积木和代码双向转换保持节点标识及执行语义。
- 浏览器完成中文界面、编辑切换和运行弹窗验证，控制台无相关错误。

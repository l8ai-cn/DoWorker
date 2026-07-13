# Loom Blockly MVP

独立浏览器原型，用 Blockly 编排单 Worker Goal Loop，并生成可预览的
canonical Loop JSON。该目录拥有自己的依赖和锁文件，不接入 AgentsMesh
后端、Runner 或真实 AI。

## 运行

```bash
cd prototypes/loom-blockly-mvp
pnpm install
pnpm dev
```

默认地址：`http://localhost:5173/`

## 验证

```bash
pnpm test
pnpm typecheck
pnpm build
```

## MVP 能力

- Goal Loop、Worker、任务、验收、验证器、边界和失败处理积木。
- 双击画布快速插入积木。
- 通过 `{{parameter}}` 创建声明式自定义任务宏。
- 结构诊断、canonical JSON 生成和逐积木模拟执行。
- 模拟器按迭代、Token、超时、无进展、同错和升级策略停止。
- 工作区、自定义积木和输出标签本地持久化。

## 边界

- 自定义积木不会执行 JavaScript、Shell 或网络请求。
- 模拟运行不会创建 Pod，也不会调用真实 Worker。
- 默认模拟场景预设验证通过，不把模拟结果表述为真实可信证据。
- Blockly workspace JSON 仅用于恢复视觉状态；`GoalLoopProgram` JSON 才表示
  编译后的执行语义。
- 正式接入必须由后端重新编译并执行权限、预算和停止条件校验。

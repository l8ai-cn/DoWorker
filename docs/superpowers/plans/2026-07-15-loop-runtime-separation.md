# Loop 运行环境分离实施计划

**目标：** 从 Loop 编排中移除 Worker，仅在运行时提交环境快照，并将 Loop 工作台完整呈现为中文。

**架构：** Go LoopScript 编译器仍是语义权威，但 AST 只包含执行过程。`RunLoopProgramRequest`
携带 `worker_spec_snapshot_id`；编译不依赖运行环境，执行前校验所选快照。Rust Core 保存语义草稿，
React 负责运行环境选择对话框和中文积木投影。

**技术栈：** Go、ConnectRPC、Protobuf、Rust/WASM、React 19、Next.js、Blockly 12.5、CodeMirror 6、Vitest。

## 任务一：运行环境无关的 LoopScript

涉及 `backend/internal/loopscript/` 下的 AST、解析、格式化、校验、编译和对应测试。

- [x] 为无 Worker 声明、无 Agent `using` 参数的 Loop 增加解析、格式化、校验和编译测试。
- [x] 运行测试并确认旧契约下新增用例失败。
- [x] 从 AST、解析、格式化、校验和启动编译中移除 Worker，同时保留执行限制与验证语义。
- [x] 运行 `go test ./backend/internal/loopscript -count=1` 并确认全部通过。

## 任务二：执行请求契约

涉及 Loop Protobuf、Connect Handler、转换逻辑、生成代码和对应测试。

- [x] 增加接口测试，证明编译不查询 Worker，运行必须提交并校验 `worker_spec_snapshot_id`。
- [x] 运行测试并确认旧契约下新增用例失败。
- [x] 从 `LoopProgram` 移除 Worker 字段，在 `RunLoopProgramRequest` 增加运行环境快照标识。
- [x] 运行 Go、TypeScript 和 Rust Protobuf 生成器。
- [x] 运行 ConnectRPC 定向测试并确认通过。

## 任务三：Rust Core 与 Web 传输

涉及 Rust Core Loop 状态测试、Web Connect 客户端和视图模型。

- [x] 更新 Rust Core 测试数据，使用与运行环境无关的 `LoopProgram`。
- [x] 运行 `cargo test -p agentcloud_state loop_builder -- --nocapture`。
- [x] 将 `runLoopProgram` 改为必须接收运行环境快照标识，并生成无 Worker 的默认 LoopScript。
- [x] 运行 Rust 状态测试和 `pnpm run web:typecheck`。

## 任务四：中文 Blockly 投影

涉及积木目录、投影器、源码生成和投影测试。

- [x] 增加测试，证明工具箱、源码和投影树不包含 Worker 节点或 `using` 参数。
- [x] 运行定向 Vitest 并确认旧实现失败。
- [x] 移除 Worker 积木与分类，将可见积木标签和结构诊断翻译为中文。
- [x] 运行定向 Vitest 并确认通过。

## 任务五：运行环境选择与中文工作台

涉及运行环境对话框、显示文本、工作台工具栏、状态面板、工作台 Hook 和组件测试。

- [x] 增加中文标签、运行环境选择、空状态和已选择提交测试。
- [x] 运行定向组件测试并确认旧实现失败。
- [x] 独立加载运行环境快照，通过“运行循环”打开中文选择对话框。
- [x] 将协议状态映射为中文，移除剩余的界面英文。
- [x] 将对话框遮罩层提升到 Blockly 浮层之上，阻止穿透交互。
- [x] 运行定向测试、类型检查和 lint。

## 任务六：集成验证

- [x] 运行 LoopScript、ConnectRPC、Rust Core、Web 单元测试、类型检查和 lint。
- [x] 验证无 Worker 编译、旧语法拒绝和运行环境必选的接口行为。
- [x] 浏览器验证中文工具箱、无 Worker 源码和双向编辑。
- [x] 浏览器验证运行环境对话框层级、选择与控制台状态。
- [ ] 推送分支、等待 CI 通过并合并到 `main`。

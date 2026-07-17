# LoopScript 语言与 AST 详细设计

- 日期：2026-07-14
- 状态：历史方案，运行环境内嵌部分已废弃
- 依赖：`2026-07-14-loop-bidirectional-feasibility-review.md`

> 2026-07-15 更新：Worker 声明和 Agent 的 `using` 参数已从 LoopScript
> 删除。当前权威边界见
> `2026-07-15-loop-runtime-separation-design.md`；本文涉及运行环境内嵌的
> 示例、语法和 AST 字段仅保留为历史设计记录。

## 1. 语言定位

LoopScript 是声明 Agent 执行过程的受限 DSL。它不是通用编程语言，不允许任意函数、动态加载、反射、文件读写或网络访问。所有外部行为必须通过注册的 Worker、Verifier、Skill、MCP Tool 或平台 Action 节点表达。

## 2. 示例

```loop
@id(n-checkout-fix)
loop checkout-fix {
  @id(n-coder)
  worker coder = snapshot(42)

  limits(
    iterations: 10,
    tokens: 80000,
    timeout: 60m,
    no_progress: 3,
    same_error: 2
  )

  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax(using: coder) {
      prompt """
      修复结算页税额计算，并补充边界测试。
      """
    }

    @id(n-tests)
    verify tests {
      command "pnpm test --filter billing"
      accept "完整测试集通过"
    }
  }

  on_failure pause
}
```

`local_id` 用于可读引用，`@id` 保存稳定 node id。重命名 local id 不改变身份；删除后重新创建才生成新 node id。

## 3. 语法边界

```text
program     := version? use_block* node_meta loop
node_meta   := "@id" "(" IDENT ")"
loop        := "loop" IDENT "{" attributed* failure_policy "}"
attributed  := node_meta? (declaration | statement)
declaration := worker | limits
statement   := agent | verify | sequence | repeat | branch | approval
             | skill | tool | worker_call | custom_call
             | evidence | succeed | fail
use_block   := "use" "block" IDENT "@" INT
custom_call := IDENT IDENT "(" argument_list? ")"
worker      := "worker" IDENT "=" "snapshot" "(" INT ")"
repeat      := "repeat" IDENT "(" "max:" INT "," "until:" REF ")" block
failure     := "on_failure" ("pause" | "fail")
```

规则：

- 一个文件只允许一个顶层 Loop。
- executable/control 节点的 `local_id` 在当前 source scope 内唯一。
- `local_id` 满足 `^[a-z0-9]+(-[a-z0-9]+)*$`，长度 2-100。
- 所有 AST 节点都必须持久化 `@id`，其值同样满足 slugkit。
- 新节点可暂时省略 `@id`；服务端分配后通过 canonical source patch 写回。
- 复制出的重复 `@id` 是阻塞错误，只能显式执行“重新生成复制节点 ID”修复。
- 条件只能引用类型化状态，如 `tests.passed`、`approval.approved`。
- 循环必须同时受局部 `max` 和全局 limits 约束。
- 字符串插值只允许已声明参数，不执行表达式。

## 4. AST

```text
LoopProgram
  schema_version
  loop_id
  workers[]
  limits
  body[]
  failure_policy
  imports[]

LoopNode
  node_id
  local_id
  node_path[]
  kind
  source_range
  leading_comments[]
  config
  children[]
  effects[]
```

`node_id` 是满足 slugkit 的不透明稳定身份，例如 `n-01j2-fix-tax`。`node_path` 是 `local_id` 数组，可显示为 `fix-cycle/fix-tax`，但该显示串不是 identifier 或数据库 lookup key。`source_range` 每次解析重建，不持久化为身份。

Parser 从 `@id` 恢复身份，不使用节点位置、文本相似度或 local id 猜测。CodeMirror 默认把 `@id` 渲染为可折叠元数据，但源文件和版本 hash 始终包含它。

AST 通过 protobuf 定义；JSON 只用于调试和文档。Proto 字段演进遵循新增字段、保留删除编号和显式 `schema_version` 的规则。

## 5. AST 与代码

- 代码编辑：source -> parser -> CST -> typed AST -> diagnostics。
- 积木编辑：semantic command -> AST reducer -> formatter -> source。
- 格式化器产生规范代码，统一缩进、参数顺序和空行。
- `#` 行注释附着到后续节点的 `leading_comments`，积木改写该节点时随节点移动。
- 无法归属的游离注释保存在 program trivia；删除节点时若注释失去归属，要求用户确认。
- 解析失败时保存用户文本，但最后有效 AST 只读，不能发布。

## 6. 类型系统

首期类型：

```text
text
integer
duration
boolean
json
command
worker-ref
node-result<T>
secret-ref
```

条件表达式只允许读取节点输出的已声明字段。禁止隐式字符串转布尔、动态属性访问和运行时类型猜测。

## 7. 诊断

```text
code
severity
message
node_id?
source_range?
field_path?
related[]
```

必须阻止发布的错误：

- 语法错误、未知节点、重复 identifier；
- 缺失 Worker、验证器、limits 或失败策略；
- 无退出条件循环；
- 条件引用不存在或类型不匹配；
- 未发布的自定义积木版本；
- 未声明权限、Secret 字面量或动态命令；
- 编译目标不支持当前节点。

## 8. 编译目标

`goal-loop-v1` 只接受单 Worker、单 Agent 任务、单命令 Verifier 和隐式重复的扁平子集，编译为现有 `CreateGoalLoopRequest`。

`loop-plan-v2` 接受 P0 结构，编译为不可变 `LoopPlan`。当程序包含 V1 不支持的节点时，编译器返回明确错误，不改写程序，也不切换目标。

## 9. Parser 与 Formatter

首期后端使用 Go 实现权威 parser、AST builder、type checker 和 formatter，结构沿用现有 `agentfile/lexer`、`agentfile/parser` 的分层方式，但 LoopScript 使用独立 package 和 token 集。

CodeMirror 使用 Lezer 做高亮、括号匹配和本地语法提示；服务端返回的 AST 与 diagnostics 才能决定发布。客户端提示与服务端结果不一致时，以服务端结果为准并显示差异，不自动接受本地结果。

Parser 必须通过同一 conformance corpus 验证：

- source -> AST fixture；
- source -> canonical source；
- canonical source 再解析后 AST 相同；
- 非法语法产生稳定 diagnostic code；
- schema/compiler version 固定到发布版本。

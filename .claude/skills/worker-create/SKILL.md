---
name: worker-create
description: |
  通过对话式 intake 结构化地创建 worker（Task 子代理），并显式绑定 skill、模型、运行模式与验收标准。
  适用于「创建 worker / 发起任务 / 委派 / multitask / 开一个 bug 猎人批次 / 帮我并行做几件事」等场景。
  也可作为多 worker 批次的协调器入口（scope 分配、去重、汇总）。
user-invocable: true
---

# Worker 创建（对话式 + Skill 绑定）

把用户的自然语言诉求，翻译成一个或多个**结构完整、显式绑定 skill** 的 Task 子代理调用。

## 何时用

- 用户说：创建 worker、发起任务、委派、并行做、multitask、开一个 xxx 批次
- 用户描述了一件（或几件）可以交给子代理独立完成的工作
- 需要为 worker 指定专业 skill（bug 猎人 / 合并 / e2e / 安全审查…）时

## 核心约束：Task API 没有 skill 参数

Cursor 的 `Task` 工具只有 `description` / `subagent_type` / `model` / `prompt` / `run_in_background` / `readonly`，**没有 `skill` 字段**。

**Workaround（本 skill 的关键）**：把 skill 绑定写进 `prompt` 里——**prompt 首行声明 `Skill: <绝对路径>`，并要求 worker 在做任何事之前先 READ 该 skill 文件并严格遵循**。这样 skill 就通过 prompt 注入生效。

## 对话式 Intake 流程

创建前，先从对话推断、缺失的用一句话补问（不要逐项审讯，能推断就别问）。需要凑齐 6 要素：

| 要素 | 含义 | 缺失时如何处理 |
|------|------|----------------|
| **目标 (goal)** | worker 要达成什么 | 必问，无法推断则问 |
| **范围 (scope)** | 目录/模块/文件边界 | 可从目标推断；多 worker 时必须切分 |
| **模式 (skill)** | 绑定哪个 skill | 用下方「Skill 选择表」路由 |
| **模型 (model)** | 用哪个模型 | 用下方「模型选择」默认值 |
| **运行模式** | background 与否、readonly 与否 | multitask/并行 → background=true；只读调查 → readonly=true |
| **验收 (accept)** | 完成的判定标准 | 缺省给「产出 + 证据 + 残余风险」 |

补问模板（仅在 2+ 要素缺失时，一次问清）：

> 我来帮你开 worker。确认一下：**目标**是 X，**范围**限定在 `Y/` 吗？要不要**后台并行**跑？完成标准我按「给出修改 + 验证证据」来算，可以吗？

## Skill 选择表（任务 → 应绑定的 skill）

| 用户诉求 | 绑定 skill（prompt 首行 `Skill:`） | Task subagent_type |
|----------|-----------------------------------|--------------------|
| 查 bug / 缺陷 / 回归 / 权限漏洞 / 性能 / 索引 / 安全边界 / 日志审计 / 超大文件 / 耦合 / 设计模式 / 旧代码清理 | `/Users/wwyz/.codex/skills/bug-hunter/SKILL.md`（对应 defect/security/permission/performance/… 模式） | `generalPurpose`（改代码）或 `explore`（只读扫描） |
| 简化 / 删冗余 / 找一个可删复杂度源头 | `/Users/wwyz/.codex/skills/bug-hunter/SKILL.md`（subtraction 模式） | `explore` |
| 跑 E2E 测试用例 | `<repo>/.claude/skills/e2e/SKILL.md` | `generalPurpose` |
| 合并当前分支到 main（GitLab） | `<repo>/.claude/skills/merge/SKILL.md` | `shell` |
| 合并到 GitHub / 处理 PR CI | `<repo>/.claude/skills/gh-merge/SKILL.md` | `shell` |
| GitLab ↔ GitHub 双向同步 | `<repo>/.claude/skills/gl-gh-sync/SKILL.md` | `shell` |
| 开 worktree 隔离开发 | `<repo>/.claude/skills/worktree/SKILL.md` | `shell` |
| 安全审查本地改动 | 无需 skill 文件 | `security-review`（内置） |
| Bugbot 式代码评审 | 无需 skill 文件 | `bugbot`（内置） |
| 查 CI 失败原因 | 无需 skill 文件 | `ci-investigator`（内置） |
| 探索代码库 / 回答「在哪」 | 无需 skill 文件 | `explore`（只读） |

> `<repo>` = `/Users/wwyz/Documents/code/Agent Cloud`。项目 skill 用仓库相对/绝对路径都可，但 worker 可能在别的 cwd 启动，**优先写绝对路径**。
> `security-review` / `bugbot` / `ci-investigator` 是 Task 内置 subagent_type，本身即专业化，不需要额外绑定 skill 文件。

## 模型选择

| 场景 | 模型 |
|------|------|
| 缺陷定位/修复、架构、需要推理 | `claude-opus-4-8-thinking-high` |
| 中等实现、常规功能 | `claude-sonnet-5-thinking-high` |
| lint / 小改 / 机械重构 / 批量替换 | `composer-2.5-fast` |
| 大范围只读探索 | `gemini-3.1-pro` / `explore` 默认即可 |

不确定就继承父代理模型（不传 `model`）。用户明确点名的模型必须在 Task 工具允许列表内，否则回退并说明。

## Task 调用模板

在 prompt **首行**放 `Skill:` 绑定 + READ 指令，随后给足上下文（worker 看不到父对话）。

```text
description: <3-5 词标题>
subagent_type: <generalPurpose | explore | shell | security-review | ...>
model: <见模型选择>
run_in_background: <true 用于 multitask/并行；否则 false>
readonly: <true 用于只读调查>
prompt: |
  Skill: <绝对路径到 SKILL.md>
  在做任何事之前，先用 Read 工具读取上面的 skill 文件并严格遵循其流程。

  仓库: /Users/wwyz/Documents/code/Agent Cloud
  目标: <goal>
  范围: <scope，明确目录/文件边界；批次中此 worker 独占的 scope>
  约束: 不要改动 scope 之外的代码；不要 git commit（除非明确要求）。
  验收: <产出物 + 证据类型（测试/日志/code-path）+ 残余风险>
  输出: 用 skill 规定的报告模板返回一份紧凑报告。
```

## 批次协调（多 worker）

参考 bug-hunter 的 batch 模式。作为 lead：

1. **切 scope**：给每个 worker 一个**互不重叠**的范围 + 独立输出路径，避免写冲突。
2. **同一批一次并行**：在**一条消息里**发多个 Task 调用（`run_in_background: true`）。
3. **覆盖优先于深挖**：先覆盖用户点名的所有方向，再深挖偏好类别。
4. **去重按边界不按标题**：合并结果时以「模块/边界」判重。
5. **汇总**：只把已确认/证据充分的发现提升进最终总结；需要建 issue 时统一在汇总后建。
6. **排除 secret 类**：API key / token / TLS secret 不在 bug-hunter 猎杀范围。

## 示例对话

### 示例 1：单 worker — 查 backend 权限

用户：「帮我查一下 backend 的权限有没有漏」

产出的 Task 调用：

```text
description: Backend 权限完整性猎杀
subagent_type: explore
model: claude-opus-4-8-thinking-high
run_in_background: true
readonly: true
prompt: |
  Skill: /Users/wwyz/.codex/skills/bug-hunter/SKILL.md
  在做任何事之前，先用 Read 读取上面的 skill 并按其 "Permission Completeness Mode Workflow" 执行。

  仓库: /Users/wwyz/Documents/code/Agent Cloud
  目标: 猎杀 backend 的权限完整性缺口（role-resource-action-state 覆盖、前后端权限漂移、缺失的后端强制、越权 affordance）。
  范围: 只看 backend/internal/api/ 与 backend/internal/service/、backend/internal/middleware/；排除 secret/API-key/TLS 处理。
  约束: 只读，不改代码，不 commit。
  验收: 用 skill 的 "Permission Completeness Report" 模板逐项给出，含 role/resource/action/state、期望 vs 实际、证据、修复方向、残余风险。
```

### 示例 2：批次 — 三个并行猎人

用户：「multitask 开三个 worker，分别查 backend 性能、web 前端可见性、runner 的日志完整性」

在一条消息里发 3 个 `Task`（均 `run_in_background: true`），各绑定 bug-hunter skill、互斥 scope：

```text
# Worker A
description: Backend 性能索引猎杀
subagent_type: explore
model: claude-opus-4-8-thinking-high
run_in_background: true
readonly: true
prompt: |
  Skill: /Users/wwyz/.codex/skills/bug-hunter/SKILL.md
  先 Read skill，按 "Performance Mode Workflow" 执行。
  仓库: /Users/wwyz/Documents/code/Agent Cloud
  范围: backend/（N+1、缺失/无用索引、无界扫描、重复网络请求）。
  验收: 每条发现用 "Performance Report" 模板，需 code-path/基准证据。

# Worker B
description: Web 前端可见性猎杀
subagent_type: explore
model: gemini-3.1-pro
run_in_background: true
readonly: true
prompt: |
  Skill: /Users/wwyz/.codex/skills/bug-hunter/SKILL.md
  先 Read skill，按 "Frontend Visibility Mode Workflow" 执行。
  仓库: /Users/wwyz/Documents/code/Agent Cloud
  范围: clients/web/（title/meta/OG、favicon/PWA、badge/未读数）。
  验收: 每条发现用 "Frontend Visibility Report" 模板。

# Worker C
description: Runner 日志完整性猎杀
subagent_type: explore
model: claude-sonnet-5-thinking-high
run_in_background: true
readonly: true
prompt: |
  Skill: /Users/wwyz/.codex/skills/bug-hunter/SKILL.md
  先 Read skill，按 "Logging Completeness Mode Workflow" 执行。
  仓库: /Users/wwyz/Documents/code/Agent Cloud
  范围: runner/（关键变更/失败/后台任务/集成缺失的审计日志、correlation id）。
  验收: 每条发现用 "Logging Completeness Report" 模板。
```

三者返回后，lead 按边界去重、汇总确认项。

### 示例 3：单 worker — 合并分支

用户：「把当前分支合并到 main」

```text
description: 合并分支到 main
subagent_type: shell
run_in_background: false
prompt: |
  Skill: /Users/wwyz/Documents/code/Agent Cloud/.claude/skills/merge/SKILL.md
  先 Read skill 并严格按其流程执行（提交 → 建 MR → 监控 Pipeline → 修错直到通过 → 合并 → 清理）。
  仓库: /Users/wwyz/Documents/code/Agent Cloud
  目标: 把当前分支合并到 main。
  验收: MR 已合并、Pipeline 通过；按 skill 的「完成后输出」格式回报。
```

## 备选：MCP delegate_task

若本会话启用了 `delegate_task` MCP 工具（见 `am-delegate` skill），可改用它把任务委派给 Agent Cloud 平台内的其他 agent；skill 绑定同样写进委派 prompt 首行。默认优先用 Cursor `Task` 工具。

## 注意事项

- **首行必须是 `Skill: <路径>` + READ 指令**——这是绑定生效的唯一机制。
- prompt 要自包含：worker 看不到父对话，目标/范围/约束/验收都要写全。
- 并行批次务必 scope 互斥 + 独立输出，避免写冲突。
- 只读调查用 `readonly: true`；multitask 用 `run_in_background: true`。
- 不要让 worker 越界改代码或擅自 commit，除非用户明确要求。

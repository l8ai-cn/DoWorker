# Agent Cloud 产品、品牌与 Logo 定义

- 日期：2026-07-14
- 状态：待确认
- 用途：首页内容、品牌表达、Logo 设计与产品口径统一
- 事实原则：区分已实现、已接入但受发布门槛限制、正在统一和规划中

## 1. 产品本质
Agent Cloud 是一个以 Expert 为业务入口、以 Worker 为执行单元的 AI 工作编排与
交付平台。它把模型、Skill、组织知识、工具、凭证、运行环境和流程组织为可复用
能力，在受控基础设施中执行，并通过人工确认、确定性验证和交付证据，把目标推进
到可验收结果。

更短的定义：

> 把分散的 AI 与组织能力组合成 Expert，把业务目标推进到可验证结果。

Agent Cloud 不应被定义为聊天机器人集合、Agent 启动器或多个角色头像组成的虚拟
团队。用户购买和管理的不是更多对话窗口，而是一种可以反复调用、受权限约束、
能够持续执行并对结果负责的组织能力。

## 2. 产品为什么存在
单个 AI Agent 已经能完成局部任务，但组织真正缺少的是将能力变成稳定交付的
控制层：

1. 模型、Prompt、Skill、知识、工具和凭证分散，无法形成稳定能力。
2. 多个 Agent 各自执行，目标、上下文、权限和结果容易割裂。
3. 长任务会停滞、重复消耗或错误自报完成，需要预算、停止条件和外部验证。
4. 跨部门工作依赖人工转述、协调和交接，过程无法沉淀为可复用方法。
5. 代码、数据和凭证需要留在组织控制的基础设施与权限边界中。

Agent Cloud 的价值不是让 Agent 数量更多，而是让分散能力形成一个可治理、可运行、
可验证、可复用的 Expert。

## 3. 核心产品模型
```text
Expert = WorkerSpec + Model + Skills + Knowledge + Tools + Workflow + Governance
```

| 对象 | 用户理解 | 产品职责 |
| --- | --- | --- |
| Expert | 对业务结果负责的专家入口 | 固化能力、上下文、权限、运行规则与交付标准 |
| Worker | 一次真实执行 | 在隔离工作区运行具体 Agent，并产生过程状态和结果 |
| WorkerSpec | Worker 的不可变能力快照 | 固定运行时、模型、仓库、Skill、知识、环境和生命周期 |
| Skill | 可复用的专业方法 | 约束执行步骤、工具使用和交付标准 |
| Knowledge | 组织上下文 | 提供文件、标准、历史资料和领域事实 |
| Tool / MCP | 对外行动能力 | 连接浏览器、数据库、Git、媒体工具和业务系统 |
| Workflow | 可重复任务 | 处理手动、API、Cron、并发、超时和运行历史 |
| Goal Loop | 一次有边界的目标闭环 | 使用验收标准、验证命令、预算和无进展检测推进目标 |
| Runner | 组织控制的执行节点 | 在自有机器或集群中启动隔离 Worker |
| Ticket / Channel / Mesh | 协作与管理面 | 承载任务、沟通、委托关系和运行拓扑 |
| Artifact / Block | 可查验交付物 | 承载代码、文件、文档、图片、音视频、数据和证据 |

## 4. 一条完整工作链
```text
业务目标
  -> 选择或定义 Expert
  -> 固定 Worker 能力与所需资源
  -> 一个或多个 Worker 在隔离环境中执行
  -> Expert 保持目标、上下文和任务状态
  -> 高风险或关键决策进入人工确认
  -> 验证命令或外部证据检查结果
  -> 交付文件、代码、系统动作、状态、风险和证据
  -> 将有效方法沉淀为 Skill、Expert 或 Workflow
```

“打通部门”不是让多个部门 Agent 同时聊天，而是让目标、上下文、权限和交付物
在同一条工作链中连续流动。部门能力按需进入，用户不再承担反复转述和人工拼接。

## 5. 产品差异化
### 5.1 Expert-first
前台保持一个稳定的业务责任主体。不同模型、Worker 和工具留在后台，避免把底层
技术选择转嫁给业务用户。

### 5.2 能力组合，不是角色堆叠
Expert 可以同时组合研发、研究、文档、数据、内容和系统操作能力。专业性来自
Skill、知识、工具、权限和验证标准的组合，不来自角色名称。

### 5.3 受控执行，不是只生成答案
Worker 在真实工作区和真实基础设施中执行，可以修改代码、读取文件、调用工具和
产生交付物；同时受隔离、凭证引用、权限、预算、停止条件和人工接管约束。

### 5.4 证据交付，不接受自报完成
Goal Loop 使用验证命令退出码判断完成，工作区产物可以作为 Artifact 展示。最终
交付应同时包含结果、执行状态、验证证据、未解决风险和下一步。

### 5.5 自托管执行底座
Backend 负责控制面，Runner 负责执行，Relay 负责终端数据面。组织可以控制代码、
网络、运行节点和凭证边界，而不是把全部工作上下文交给单一云端聊天产品。

## 6. 当前事实与表达边界
### 已有坚实实现
- Runner、Relay、Backend 构成的控制面与数据面分离。
- 隔离工作区、实时终端或 ACP 会话、人工输入与权限处理。
- WorkerSpec、不可变快照、模型与工具模型绑定、Skill、知识和环境引用。
- Goal Loop 的验收标准、验证命令、迭代与时间预算、无进展和同错检测。
- Workflow 的触发、Cron、并发、超时、沙箱和运行记录。
- Ticket、Channel、Mesh、知识库和工作区 Artifact。

### 已接入但不能写成“全部正式可用”
- 当前目录包含 13 个 Worker Definition，包括 Codex CLI、Claude Code、
  Gemini CLI、OpenClaw、Do Agent 和 Seedance Expert。
- 正式运行时发布目录中的镜像仍为禁用状态，README 明确记录当前没有 Worker
  类型完成全部正式可部署门槛。
- 对外应使用“已接入 Worker 类型目录，具体可用性以环境和发布验证为准”，不能
  使用“13 种 Worker 已可直接部署”。

### 正在统一
- Expert 是已确认的产品主入口，但当前创建、编辑、发布和运行仍存在新旧配置
  路径差异；可运行 Expert 必须绑定 WorkerSpec 快照。
- Workflow 仍保留部分可变运行字段，目标架构是绑定不可变 WorkerSpec revision。
- 当前市场有三个应用条目，但安装路径未形成完整的快照绑定执行闭环，不能写成
  “开箱即用且已经验证完成”。
- 资源原生 Validate、Plan、Apply 已覆盖部分资源，Expert、Workflow 和 GoalLoop
  的完整资源化仍在后续接入范围。

### 规划中
- 完整行业 Expert 市场、Connector、Entitlement、Quota Ledger 和两阶段安装。
- 跨境增长、AI 教育与 AI 伙伴的经过验证的行业应用包。

## 7. 品牌主张

### 核心主张

> 把能力组织成专家，把目标推进到结果。

### 首页主标题

> 把分散的 AI 能力，组织成真正完成工作的专家。

### 首页说明

Agent Cloud 将模型、Skills、组织知识、工具和运行规则组合为一个可复用 Expert。
它可以在受控基础设施上调用一个或多个 Worker，跨越研发、内容、数据与运营持续
执行，在关键节点接受人工确认，并交付可检查的文件、代码、系统动作与证据。

### 英文表达

- Category: `AI Expert Orchestration and Delivery Platform`
- Brand line: `Compose capability. Deliver outcomes.`
- Existing company line can remain: `Where teams scale beyond headcount.`

## 8. 品牌词汇

优先使用：

- 组织能力、Expert、工作链、受控执行、持续推进、人工确认
- 可验证结果、交付物、证据、复用、版本、权限边界
- Worker 执行、隔离工作区、自托管 Runner、模型与工具装配

避免使用：

- 无限人力、全自动、无需人工、支持所有系统、一键完成一切
- 多个 Agent 聊天、智能体矩阵、拟人化 AI 军团、替代所有岗位
- 将已接入目录写成已通过正式发布验证

## 9. Logo 设计语义

Logo 以“能力合芯”表达 Agent Cloud 的核心机制：

1. **能力来源不同**：四个方向和尺寸不完全相同的结构模块代表运行时、专业方法、
   组织上下文以及行动与治理。
2. **Expert 是组合结果**：四个模块以错缝方式共同形成一个稳定外轮廓，Expert
   就是组合后的整体，不是模块之外的第五个对象。
3. **统一责任主体**：完整轮廓代表统一目标、上下文、任务状态和交付责任；薄荷色
   模块只表示当前被激活的能力面。
4. **装配后整体运行**：标志完成砌合后应被理解为一个执行组件，而不是多个角色、
   节点或部门的并列集合。

图形原则：

- 使用正交、厚实、略微非对称的错缝砌合结构，不使用品牌首字母作为主题。
- 一个能力模块使用薄荷绿；其余模块使用石墨黑或暖白，完整轮廓共同代表 Expert。
- 不设置独立中央核心，避免把 Expert 错误表达成被能力围绕的第五个对象。
- Image2 生成稿只用于轮廓探索，最终标志必须重绘为无渐变、无阴影的原生 SVG。
- 图形必须支持单色版，并在 16px、24px、32px 和 64px 下保持识别。
- 避免网络节点、机器人头、脑、聊天气泡、蜂窝、花朵、齿轮、无限符号、十字、
  普通播放键、箭头和流程图结构；也不使用任何隐藏字母替代产品理念。

## 10. Logo 验收

- 第一眼识别为多个结构咬合成一个整体，不应先读成任何字母或常见符号。
- 轮廓保持紧凑方形，四个模块通过错位接缝形成一个完整结构，没有细碎缝隙。
- 遮掉薄荷色后，黑白版本仍能看出四个模块共同构成一个主体。
- 深色与浅色背景均清晰，16px 下不退化成花、十字、二维码或普通方块。
- 能同时用于导航栏、favicon、PWA 图标、文档和社交头像。

## 11. 事实依据

- 产品与发布状态：`README.md`、`config/worker-types/catalog.json`、
  `backend/internal/domain/workerruntime/runtime_catalog.lock.json`
- 执行能力：`backend/internal/domain/workerspec/`、
  `backend/internal/domain/agentpod/autopilot_controller.go`
- Expert 与市场：`backend/internal/domain/expert/expert.go`、
  `backend/internal/service/expert/run.go`、`backend/internal/service/expert/marketplace.go`
- 持续执行：`backend/internal/domain/goalloop/goal_loop.go`、
  `backend/internal/domain/workflow/workflow.go`
- 能力与界面：`backend/internal/domain/skill/skill.go`、
  `clients/web/src/components/pod/CreatePodForm/`、
  `clients/web/src/components/workspace/agent-ui/`

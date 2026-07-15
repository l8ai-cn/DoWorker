# Do Worker 首页内容迁移审计

- **日期：** 2026-07-13
- **状态：** 已实施并完成浏览器验收
- **上游设计：** `2026-07-13-product-capability-homepage-content-design.md`
- **审计范围：** `/`、首页导航、页脚、结构化数据、首页多语言内容

## 1. 当前首页结论

当前首页已经具备五个可以保留的产品事实：目标输入、执行过程、人工确认、结果
交付、自托管治理。但页面仍围绕“组建多个拟人化角色”组织内容，与新的
Expert 统一入口存在根本冲突。

主要差距：

1. Hero 标题是“设定目标，组建一支 AI 团队”，没有建立 Expert 概念。
2. Mission Console 以三名 Worker 角色卡为视觉中心，强化了 N 个角色的理解。
3. 六个通用场景没有对应跨境电商、AI 教育、AI 伙伴和市场四个产品入口。
4. 没有解释 Worker、Skill、Knowledge、MCP、Workflow 如何组成 Expert。
5. 文档、办公、编剧编导、图片、音频、视频等能力没有边界清晰的能力地图。
6. 市场只存在于导航链接，没有成为 Expert 和能力分发的产品闭环。
7. 首页和页脚仍包含定价入口，结构化数据仍包含免费 Offer。
8. 当前兼容性区域只展示 7 个 Worker 和 Custom Agent，与正式 12 类目录不一致。
9. `WorkforceBackdrop` 和 `FinalCTA` 使用模糊光球，不符合新 UI 规范。
10. 大量胶囊按钮、负字距和单色青绿强调需要按新视觉系统收敛。

## 2. 迁移原则

- 页面第一层对象从“多个 Agent 角色”改为“一个承担结果的 Expert”。
- Worker 协作保留为底层能力，不作为用户必须先理解的产品入口。
- 四个业务菜单展示完整链路和交付物，不展示虚构的系统连接。
- 所有能力标注为已实现、需安装能力或规划中。
- 复用当前人工确认、执行证据、自托管和多运行时的真实能力。
- 首页删除定价，不删除独立计费能力和企业页面。
- 营销页继续使用 `useLightSession`，禁止引入 Rust/WASM 和业务 Zustand store。
- 全部首页组件保持 200 行以内，拆分数据、视觉和交互职责。

## 3. 页面级迁移表

| 当前文件/区域 | 处理 | 目标 |
| --- | --- | --- |
| `app/page.tsx` | 修改 | 移除 Pricing，更新 JSON-LD 定位、关键词和 Offer |
| `Navbar.tsx` | 修改 | 导航改为跨境电商、AI 教育、AI 伙伴、市场、产品能力、文档 |
| 旧 Workforce 首页组件 | 保留但不再组合 | 避免在本次首页改版中扩大清理范围 |
| `landing/expert-home/` | 新增 | 承载 Hero、控制面、四个业务专区、能力、市场和治理 |
| 新增 Expert 区域 | 新增 | 解释 Expert 的组成和复用逻辑 |
| 新增 Workflow 区域 | 新增 | 手动/API/Cron、持续执行、预算、接管和失败证据 |
| 新增 Market 区域 | 新增 | 展示现有三个专家应用和市场分发逻辑 |
| `FinalCTA.tsx` | 改写 | 创建 Expert、浏览市场；移除免费套餐导向 |
| `Footer.tsx` | 修改 | 删除定价，增加四个解决方案和专家市场入口 |
| `globals.css` 营销样式 | 外科式修改 | 移除光球，建立石墨/近白/薄荷/琥珀四层语义 |

旧版 `HeroSection`、`CoreFeatures`、`WhyTerminalBased` 等当前没有被首页组合，
本次不顺手删除；只有确认无引用并单独安排清理时才能移除。

## 4. 目标首页组件树

```text
Home
├── Navbar
├── ExpertHero
│   └── ExpertControlSurface
├── ExpertFormula
├── SolutionDomains
│   ├── CrossBorderCommerce
│   ├── AIEducation
│   ├── AIPartners
│   └── Marketplace
├── CapabilitySpectrum
├── UnifiedWorkChain
├── ExpertSystem
├── WorkflowOperations
├── HumanReviewAndDelivery
├── TrustAndDeployment
│   └── WorkerRuntimeCatalog
├── MarketplacePreview
├── FinalCTA
└── Footer
```

这些是职责边界，不要求每个节点都成为一个大卡片或独立视觉容器。

## 5. Hero 内容合同

主标题：

> 把分散的 AI 能力，组织成真正完成工作的专家

说明：

> 在一个工作空间里连接模型、Worker、Skill、知识与业务系统，从目标拆解、
> 执行协作、人工确认到结果交付，打通完整工作链路。

主操作为“创建 Expert”，次操作为“浏览专家市场”。演示控制面必须包含：

- 一个业务目标和一个 Expert 名称。
- Worker 类型、模型、Skill、知识和系统连接的装配状态。
- 四到五个 Workflow 步骤和当前运行状态。
- 至少一个琥珀色人工确认点。
- 一个可检查的交付物区域。
- 暂停、继续和重放控制，并支持 `prefers-reduced-motion`。

## 6. 四个菜单合同

| 菜单 | 首屏承诺 | 示例交付 |
| --- | --- | --- |
| 跨境电商 | 把研究、商品内容、素材、审核和经营复盘接成一条链 | 市场报告、Listing、素材包、日报 |
| AI 教育 | 从教学目标推进到经过教师审核的课程资产 | 大纲、课件、实验、练习、多媒体资源 |
| AI 伙伴 | 与团队共享上下文、跨部门组合能力并在关键节点确认 | 协作结果、系统动作、进展说明、可查验证据 |
| 市场 | 按业务结果发现、启用和复用专家能力 | Expert、Skill；连接与资源标注为规划方向 |

首版菜单可以锚定首页对应区域；只有在内容和真实功能足以支撑独立页面时再增加
路由，不能创建只有营销文案的空壳页面。

## 7. 能力展示合同

能力光谱包含编程研发、文档办公、研究知识、内容创作、编剧编导、图片、音频、
视频、数据运营和教育培训。展示规则：

- 编程研发标记为“已验证应用”。
- 平台运行、Workflow、Expert、Skill、Knowledge、媒体承载标记为“平台能力”。
- 依赖 Skill/MCP/外部工具的能力使用“可装配”。
- 多租户市场、Space、连接和资源 Listing 使用“规划中”。
- 不出现“原生生成所有媒体”“支持所有办公格式”“自动接入全部电商平台”。

## 8. 多语言与文案迁移

首页由 `landing.json`、`workforce.json` 和 `expert-home.json` 共同提供文案，覆盖 `de`、`en`、`es`、
`fr`、`ja`、`ko`、`pt`、`zh` 八个 locale。实现时：

- 新首页内容集中到 `expert-home.json`，运行时合并到
  `landing.workforce.expertHome`；导航和共用页脚保留在 `landing.json`。
- 先建立类型化 key 清单，再同步八个 locale，禁止运行时缺 key。
- 保留当前中文文件中用户已修改的“Workflow 调度”。
- 删除页面引用后再决定是否清理旧翻译，避免无关大规模 churn。

## 9. 验收证据

- 首页不再出现定价板块、`/#pricing` 导航和页脚链接。
- Hero、场景区和控制面不再以多个角色卡为主要叙事。
- 四个菜单均可从桌面和移动导航到达。
- Expert 公式、能力边界、Workflow、人工确认、交付、自托管和 12 类 Worker 可见。
- 当前三个市场应用名称与后端市场数据一致。
- 八个 locale 的新 key 完整，中文内容符合已确认文案。
- 单元测试覆盖菜单切换、控制面播放、暂停、重放和减弱动画。
- 浏览器自动化覆盖桌面与移动主路径、导航、状态和无重叠截图。
- 浏览器 console 无错误，关键请求无失败，页面无横向溢出。
- `pnpm run web:lint`、`web:typecheck`、`web:test` 通过。
- `check-no-wasm-in-marketing.sh` 的营销页负向检查通过；其生产构建正向检查依赖
  chunk 中保留 `WasmProvider` 字面量，当前压缩产物不满足该实现假设。源代码导入链
  已确认 Dashboard 和 Popout 仍通过 `AuthBootstrap` 挂载 `WasmProvider`。

## 10. 实施边界

本次只改首页及其直接依赖的营销组件、文案和测试。不修改 Dashboard、Rust Core、
Backend 合同、Marketplace 后端或计费业务。现有脏工作区中的其他改动必须保留。

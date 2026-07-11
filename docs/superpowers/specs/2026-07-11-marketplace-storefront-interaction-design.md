# Marketplace Storefront Interaction Design

- **Date:** 2026-07-11
- **Scope:** 市场前台、内容详情、获取流程、我的应用和额度

## 1. Terms and Actions

| Technical term | Chinese UI |
| --- | --- |
| Marketplace | 市场 |
| Space | 专区 |
| Application | 应用 |
| Skill | Skill |
| MCP Connector | 系统连接 |
| Resource | 资源 |
| Publisher | 发布方 |
| Entitlement | 使用权限 |
| Installation | 启用实例 |
| Marketplace Credit | 市场额度 |
| API Token | 访问令牌 |

主按钮按类型固定：应用“启用应用”、Skill“安装 Skill”、系统连接“连接系统”、
资源“申请使用”或“添加资源”。

## 2. Page Tree

```text
/
/spaces/[spaceSlug]
/catalog?type=&space=&q=
/listings/[listingSlug]
/acquire/[listingSlug]
/library
/library/[installationId]
/usage
/requests
```

## 3. Market Home

用户角色是消费者，主要对象是专区和 Listing，主操作是发现能力。

页面区域：

1. Header：品牌、专区、全部内容、我的应用、额度、账户。
2. 市场说明：市场名称和 summary。
3. 专区入口：名称、摘要和已上架数量。
4. 精选集合：仅由管理员配置。
5. 最新上架：按 published_at 排序。

文案：

- 搜索：“搜索应用、Skill、系统连接或资源”
- 空市场：“这个市场还没有可用内容”
- 空说明：“市场管理员完成上架后，内容会显示在这里。”
- 暂停：“市场暂时停止服务”
- 暂停说明：“你仍可查看已启用内容，但不能获取或安装新内容。”

## 4. Space, Catalog, and Cards

筛选项为内容类型、适用对象、发布方、权限模式和额度范围。URL query 是筛选
SSOT，刷新和分享后结果一致；移动端使用筛选 Sheet。

卡片显示图标、名称、tagline、类型、发布方、验证状态、额度摘要和维护状态。
Skill、权限和依赖数量只在详情页高级信息展示。

列表必须有 loading skeleton、空筛选、加载失败和 market suspended 状态。

## 5. Listing Detail

信息顺序：

1. 名称、类型、发布方、认证标识、所属专区。
2. tagline、最多三个 outcome、主按钮。
3. 示例输入和示例结果。
4. 适用对象和使用场景。
5. 所需账号、资源、权限和额度。
6. 最近验证、版本和维护状态。
7. 依赖、Skill、MCP tools 等高级装配信息。
8. 文档、支持、版本历史和变更说明。

状态文案：

- “已获得使用权限”
- “已在「{organization}」启用”
- “需要组织管理员批准”
- “发布方已暂停此内容，暂时不能新启用”
- “当前版本已停止维护，请升级后继续使用”

## 6. Acquisition Wizard

```text
登录 -> 选择组织 -> 检查条件 -> 配置
     -> 确认权限与额度 -> 执行 -> 完成
```

### Select Organization

- 标题：“选择使用组织”
- 说明：“应用和连接会安装到你选择的组织。创建后不能直接移动。”
- 无组织按钮：“创建组织”
- 无权限：“你可以向该组织管理员发送启用申请。”
- 禁止自动选择第一个组织。

### Preflight

结果按“可继续”“需要配置”“阻塞问题”分组：

- “运行条件已满足”
- “还需要完成 {count} 项配置”
- “暂时不能启用”
- 操作：“重新检查”
- 过期：“市场内容或组织资源已变化，请重新检查后确认。”

每个问题包含对象、原因和下一步。例如：

```text
缺少兼容模型
这个应用需要支持工具调用的文本模型。请先让组织管理员添加兼容模型。
```

### Configuration and Confirmation

配置 schema 来自 Runtime Preflight，常见字段是模型、仓库范围、Worker 类型、
计算目标、MCP 授权、密钥引用和默认任务。密钥仅选择 Runtime 已有引用。

确认页展示目标组织、创建资源、权限、外部系统、预估额度、版本和更新策略。
主按钮固定“确认并启用”，不能写“确定”或“提交”。

高风险权限文案：

```text
此应用可以修改仓库内容并创建合并请求。
确认后，这些权限仅在「{organization}」及所选仓库范围内生效。
```

### Progress and Completion

阶段为验证权限、预占额度、安装依赖、创建实例、运行验证、结算额度。关闭页面
不取消操作；通过 operation ID 恢复状态。

- 成功：“应用已启用”
- 成功说明：“已在「{organization}」创建可用实例。”
- 主操作：“运行示例任务”
- 次操作：“查看启用实例”
- 失败：“启用未完成”
- 原因：“失败阶段：{stage}。{reason}”
- 已补偿：“本次创建内容已清理，预占额度已返还。”
- 补偿中：“正在清理已创建内容，暂时不要重复启用。”

## 7. Library and Usage

`/library` 展示启用实例，筛选 active、needs_configuration、
upgrade_available、suspended 和 failed。详情显示来源市场、版本、组织、配置
摘要、验证、最近使用、额度、权限和操作历史。

`/usage` 展示可用额度、已预占、周期消费和下次刷新，再按应用、成员和 meter
展示明细。禁止把余额与模型 Token 合并为一个数字。

## 8. Error Contract

| Code | Title | Recovery |
| --- | --- | --- |
| `MARKET_NOT_FOUND` | 找不到这个市场 | 返回市场列表 |
| `MARKET_SUSPENDED` | 市场暂时停止服务 | 查看我的应用 |
| `LISTING_NOT_AVAILABLE` | 此内容当前不可获取 | 浏览同专区内容 |
| `ORG_SELECTION_REQUIRED` | 请选择使用组织 | 选择组织 |
| `APPROVAL_REQUIRED` | 需要管理员批准 | 发送启用申请 |
| `QUOTA_INSUFFICIENT` | 市场额度不足 | 查看额度详情 |
| `RUNTIME_INCOMPATIBLE` | 组织资源不满足运行条件 | 查看阻塞问题 |
| `PLAN_EXPIRED` | 启用计划已过期 | 重新检查 |
| `INSTALLATION_CONFLICT` | 已存在冲突的启用实例 | 查看现有实例 |
| `COMPENSATION_PENDING` | 正在清理未完成的启用 | 查看进度 |

前端按稳定 code 选择标题，服务端 detail 只作为具体原因展示。

## 9. Required States

所有列表、详情、表单和命令覆盖 loading、empty、error、permission denied、
disabled、selected、pending approval、suspended、conflict 和 retryable failure。
移动端不得隐藏权限、额度或阻塞原因。

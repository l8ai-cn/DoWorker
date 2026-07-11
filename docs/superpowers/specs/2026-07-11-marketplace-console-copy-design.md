# Marketplace Console and Copy Design

- **Date:** 2026-07-11
- **Scope:** 市场运营台、发布台、字段和中文文案
- **Visual direction:** 现有语义 token + data-admin

## 1. Page Tree

```text
/overview
/market/general
/market/brand
/market/domain
/market/identity
/spaces
/catalog
/listings
/submissions
/publishers
/members
/quota/plans
/quota/accounts
/usage
/audit
```

Console 使用固定侧边栏和表格/详情模式，不设计营销 Hero。每页只有一个主要
工作，重复操作使用紧凑表格，破坏性操作必须二次确认。

## 2. Overview

展示配置完成度、待审核数量、发布 Listing、活跃安装、安装成功率、额度消耗和
异常操作。主按钮依据市场状态显示“继续配置市场”或“创建上架项”。

不展示无业务意义的总注册数或装饰性趋势图。异常卡片必须能进入具体记录。

## 3. Create Marketplace

| Field | Label | Helper |
| --- | --- | --- |
| `template_key` | 市场模板 | “模板只提供初始专区和页面结构，创建后可以调整。” |
| `name` | 市场名称 | “例如：海贸通跨境电商 AI 市场” |
| `slug` | 市场标识 | “用于平台域名，创建后不可修改。” |
| `summary` | 市场简介 | “用一句话说明这个市场服务谁、解决什么问题。” |
| `visibility` | 访问范围 | “公开市场可被任何人浏览；私有市场仅成员可见。” |
| `registration_mode` | 加入方式 | “选择用户如何成为市场成员。” |

主按钮：“创建市场”。成功后进入配置检查清单，不直接发布。

## 4. Market Settings

### General

字段：市场名称、简介、完整介绍、默认语言、访问范围、加入方式。slug 只读。
保存按钮：“保存基本信息”。

### Brand

字段：品牌名、Logo、Favicon、主色、首页主视觉、布局模板和首页区块排序。
主色实时检查文本对比度；不允许上传 CSS、HTML 或脚本。

按钮：“保存品牌设置”；预览操作：“预览市场首页”。

### Domain

列表显示域名、类型、验证状态、主域名和最近错误。操作为“添加自定义域名”、
“检查验证状态”、“设为主域名”、“停用域名”。

验证说明必须显示具体 DNS 记录名和值，失败显示 error code 和下一步。

### Identity

字段：公开注册、邀请、SSO provider、允许邮箱域和默认成员角色。启用 SSO 前
必须通过 discovery、callback 和测试登录。

## 5. Spaces

列表字段：名称、标识、状态、维护者、公开上架数和排序。主按钮：“创建专区”。

编辑字段：

- 专区名称：“例如：商品运营”
- 专区标识：“用于专区 URL，创建后不可修改。”
- 简介：“说明专区解决的业务问题。”
- 完整介绍、图标、维护者、审核者、排序和状态。

空状态：“还没有专区。先创建专区，再把上架内容组织到业务场景中。”

## 6. Catalog and Publisher

Catalog 列表显示资源名称、类型、发布方、来源、最新版本、验证状态和引用市场。
主按钮：“登记资源”。来源只能从 Runtime 可发布资源中选择。

发布方字段：类型、展示名、标识、简介、Logo、平台用户或组织引用。认证操作为
“提交认证”、“通过认证”和“撤销认证”，均写审计。

## 7. Listing Editor

编辑器分区：

1. 来源版本：资源、版本、digest，只读。
2. 市场介绍：展示名、tagline、描述、outcomes、use cases、适用对象。
3. 专区与标签：至少一个专区，最多 12 个标签。
4. 使用条件：账号、模型、算力、地域和其他前置条件。
5. 权限：仓库、工具、MCP scope 和外部系统。
6. 额度：方案、meter、预估示例。
7. 媒体：图标、主视觉、最多 6 个截图。
8. 支持：文档 URL、支持 URL。
9. 版本说明：release notes。

操作文案：

- “保存草稿”
- “预览上架页”
- “提交审核”
- “撤回审核”
- “批准上架”
- “要求修改”
- “暂停上架”

要求修改必须填写“需要修改的内容”。禁止任意 HTML、脚本、CSS、明文凭证和
自由格式安装命令。

## 8. Submission Review

审查页左侧显示提交内容，右侧显示 schema、依赖、权限、安全、安装和验收结果。
Reviewer 必须逐项确认 blocking checks 全部通过。

权限比上一版本扩大时显示独立警告：

```text
这个版本新增了 {count} 项权限。
现有用户升级时需要重新获得组织管理员批准。
```

发布方不能审核自己的提交；无负责 Space 权限时按钮 disabled 并说明原因。

## 9. Members and Roles

列表字段：成员、角色、状态、加入方式和最近活动。操作是“邀请成员”、
“修改角色”、“暂停成员”和“移除成员”。

移除最后一个 owner 必须阻止，文案：“市场必须至少保留一名所有者。”

Marketplace role 与 Runtime organization role 分开显示，避免误认为市场管理员
自动拥有组织运行权限。

## 10. Quota Console

额度方案分为“周期与发放额度”“计量规则”“适用内容”。人工调整必须填写原因，
确认页展示调整前后 available、reserved 和 consumed。

- 主按钮：“创建额度方案”
- 调整按钮：“调整额度”
- 确认按钮：“确认调整额度”
- 空状态：“还没有额度方案。市场发布前至少需要一个有效方案。”

## 11. Usage and Audit

Usage 支持按周期、Listing、组织、用户和 meter 筛选；显示 accepted、settled、
rejected 和 settlement shortfall。

Audit 显示时间、操作人、动作、对象、结果和 request ID。详情显示 old/new data，
但默认遮蔽敏感字段。

## 12. State Contract

每个页面覆盖 loading、empty、error、permission denied、dirty、saving、saved、
revision conflict、pending、disabled 和 destructive confirmation。PATCH 冲突文案：

“这条记录已被其他管理员更新。请刷新后比较变化，再重新提交。”

# AgentCloud Embedded Marketplace Design

- **Date:** 2026-07-12
- **Status:** Superseded
- **Superseded by:** `2026-07-12-public-app-marketplace-closed-loop-design.md`

## Product Decision

This historical MVP made Marketplace a AgentCloud capability page. It is no
longer the product direction and must not be used for new implementation.

| Concern | Decision |
| --- | --- |
| Canonical route | `/{org}/marketplace` |
| Navigation | Dedicated “市场” activity in the AgentCloud activity bar |
| Installation target | URL organization only |
| Resources | Application, Skill, system connector, and resource |
| Completion | Same organization’s Experts page |
| Legacy host | Redirect `market.l8ai.cn` to `https://dowork.l8ai.cn/dev-org/marketplace` |
| Service boundary | Marketplace API and its database remain independent |

`dev-org` is the seeded production organization used by the former public market.
No anonymous organization inference is added in this MVP.

## User Flow

```text
AgentCloud organization
  -> 市场
  -> search/filter by Space and resource type
  -> resource detail
  -> 检查启用条件
  -> confirm permissions and market credits for current organization
  -> enable
  -> Experts in the same organization
```

The server remains authoritative for membership, plan digest, quota reservation,
and installation. The user never selects a second organization in this flow.

## UI Contract

- Header: “应用市场”, current organization context, search, and catalog count.
- Filters: “全部”, “应用”, “Skill”, “系统连接”, “资源”, then Space filters.
- Cards: type, verified publisher, Space, estimated market credits, and details.
- Detail: outcomes, use cases, requirements, permissions, and current version.
- Confirm: exact target organization, credits, immutable plan expiry, permissions,
  and `确认并启用`.

No independent site header, fake account state, English navigation, or duplicate
marketing hero is rendered.

## MVP Capability Boundary

The runtime bridge currently installs application Listings as Experts. Skill,
connector, and resource cards remain discoverable but state that their runtime
integration is not yet available; they do not expose a succeeding action.

## Acceptance Criteria

1. `/{org}/marketplace` renders inside `DashboardShell`, highlights “市场”, and
   preserves the current organization.
2. Catalog and detail use `/api/marketplace/v1`.
3. `/{org}/marketplace/acquire` does not display an organization picker.
4. Successful application enablement leads to `/{org}/experts`.
5. `market.l8ai.cn` serves neither standalone frontend nor Marketplace API routes.

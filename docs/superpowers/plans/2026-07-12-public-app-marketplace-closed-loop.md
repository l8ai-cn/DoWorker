# Public App Marketplace Closed-Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the Agent Cloud Marketplace into a public outcome-oriented
storefront with server-side taxonomy filtering and an authenticated
organization application center that continues into first use.

**Architecture:** Keep the Marketplace API and database as catalog and
installation SSOT. Re-enable `clients/marketplace-web` as the standalone
public storefront at `market.l8ai.cn`; use the Agent Cloud web application only
for authenticated acquisition and organization-owned applications. Extend API
data before replacing frontend routes so both surfaces share the same contract.

**Tech Stack:** Go, Gin, GORM/PostgreSQL migrations, Next.js App Router,
TypeScript, Tailwind CSS, Vitest, Go test, Kubernetes Kustomize, doops.

---

## File Ownership

| Area | Primary files |
| --- | --- |
| Taxonomy and public API | `marketplace/migrations/`, `marketplace/internal/service/`, `marketplace/internal/api/public/`, `marketplace/internal/infra/postgres/` |
| Public storefront | `clients/marketplace-web/src/` |
| Acquisition and application center | `clients/web/src/app/`, `clients/web/src/components/marketplace/`, `clients/web/src/lib/marketplace/` |
| GitOps routing | `deploy/kubernetes/cluster-oilan/` |

### Task 1: Model And Expose Marketplace Taxonomy

**Files:**
- Create: `marketplace/migrations/000011_listing_taxonomy.up.sql`
- Create: `marketplace/migrations/000011_listing_taxonomy.down.sql`
- Modify: `marketplace/internal/service/storefront_types.go`
- Modify: `marketplace/internal/infra/postgres/storefront_listing_repository.go`
- Modify: `marketplace/internal/api/public/storefront_handler.go`
- Modify: `marketplace/internal/api/public/storefront_response.go`
- Test: `marketplace/internal/api/public/storefront_handler_test.go`
- Test: `marketplace/internal/infra/postgres/storefront_repository_test.go`

- [ ] Write a failing list-handler test for `scene=software-delivery` and
  `industry=enterprise-services`, asserting response tags and no unrelated item.
- [ ] Run `go test ./marketplace/internal/api/public -run TestListListings` and
  verify the new expectation fails because query filtering and tags are absent.
- [ ] Add taxonomy tables, data migration from `listing_versions.tags`, tag
  constraints, indexes, service query fields, repository filtering, response
  tags, and cursor parsing.
- [ ] Run API and repository tests until the new server-side filter test passes.
- [ ] Commit with `feat(marketplace): add public listing taxonomy filters`.

### Task 2: Define Package And First-Run Listing Contract

**Files:**
- Create: `marketplace/migrations/000012_listing_activation_content.up.sql`
- Create: `marketplace/migrations/000012_listing_activation_content.down.sql`
- Modify: `marketplace/internal/service/storefront_types.go`
- Modify: `marketplace/internal/infra/postgres/storefront_listing_repository.go`
- Modify: `marketplace/internal/api/public/storefront_response.go`
- Test: `marketplace/internal/api/public/storefront_handler_test.go`
- Test: `marketplace/internal/infra/postgres/storefront_repository_test.go`

- [ ] Write a failing detail-handler test that expects `package_summary`,
  `activation_requirements`, `first_run_templates`, `documentation_url`, and
  `support_url`.
- [ ] Run the targeted public API test and verify it fails on the missing fields.
- [ ] Add listing-version columns and map the immutable fields through
  StorefrontService without duplicating runtime secrets or credentials.
- [ ] Update the seeded software-delivery application with its package,
  readiness requirements, documentation, and one executable first task.
- [ ] Run the targeted API and repository tests and commit with
  `feat(marketplace): publish application activation guidance`.

### Task 3: Restore The Public Storefront

**Files:**
- Modify: `clients/marketplace-web/src/lib/marketplace-types.ts`
- Modify: `clients/marketplace-web/src/lib/marketplace-api.ts`
- Modify: `clients/marketplace-web/src/lib/listing-filters.ts`
- Modify: `clients/marketplace-web/src/components/catalog-page-content.tsx`
- Modify: `clients/marketplace-web/src/components/catalog-filters.tsx`
- Create: `clients/marketplace-web/src/components/outcome-space-grid.tsx`
- Modify: `clients/marketplace-web/src/components/listing-card.tsx`
- Modify: `clients/marketplace-web/src/components/detail-content.tsx`
- Modify: `clients/marketplace-web/src/components/detail-hero.tsx`
- Modify: `clients/marketplace-web/src/components/site-header.tsx`
- Modify: `clients/marketplace-web/src/styles/*.css`
- Test: `clients/marketplace-web/src/lib/listing-filters.test.ts`
- Test: `clients/marketplace-web/src/components/detail-content.test.tsx`

- [ ] Write failing filter parser tests for `scene`, `industry`, `audience`,
  `integration`, `readiness`, and `sort`.
- [ ] Run the tests and verify they fail because the current client only accepts
  query, type, and Space.
- [ ] Implement URL-preserving filters, outcome Space navigation, honest
  availability labels, package composition, first-run detail content, and
  Chinese public navigation without placeholder entries.
- [ ] Run marketplace-web unit tests, typecheck, and production build.
- [ ] Commit with `feat(storefront): build outcome-led public marketplace`.

### Task 4: Make Public Acquisition Preserve Intent

**Files:**
- Modify: `clients/marketplace-web/src/lib/acquire-link.ts`
- Modify: `clients/marketplace-web/src/components/acquire-button.tsx`
- Modify: `clients/web/src/components/marketplace/acquire/MarketplaceAcquireFlow.tsx`
- Modify: `clients/web/src/lib/marketplace/acquire-api.ts`
- Test: `clients/marketplace-web/src/lib/acquire-link.test.ts`
- Test: `clients/web/src/components/marketplace/acquire/MarketplaceAcquireFlow.test.tsx`

- [ ] Write a failing acquisition-link test proving the public listing URL and
  market/listing/version intent survive the Core Web handoff.
- [ ] Write a failing flow test proving a user with multiple organizations sees
  an explicit selector and no installation mutation happens before preflight.
- [ ] Implement public-to-core return intent, login redirect preservation,
  explicit organization selection, preflight blockers, and plan confirmation.
- [ ] Run targeted tests and commit with
  `feat(marketplace): preserve public acquisition intent`.

### Task 5: Add Organization Application Center And First Use

**Files:**
- Create: `clients/web/src/app/(dashboard)/[org]/applications/page.tsx`
- Create: `clients/web/src/app/(dashboard)/[org]/applications/[installationId]/page.tsx`
- Create: `clients/web/src/components/applications/OrganizationApplicationList.tsx`
- Create: `clients/web/src/components/applications/ApplicationFirstRun.tsx`
- Create: `clients/web/src/lib/marketplace/installation-api.ts`
- Modify: `marketplace/internal/api/consumer/installation_handler.go`
- Modify: `marketplace/internal/service/installation_orchestration_service.go`
- Modify: `clients/web/src/components/marketplace/acquire/MarketplaceAcquireFlow.tsx`
- Test: `marketplace/internal/api/consumer/installation_handler_test.go`
- Test: `clients/web/src/components/applications/OrganizationApplicationList.test.tsx`

- [ ] Write a failing authenticated API test for listing only the caller's
  organization installations and returning status, quota, first-run template,
  and runtime reference.
- [ ] Write a failing component test asserting active and needs-attention
  applications show different primary actions.
- [ ] Implement read-only installation queries, organization API client,
  application list/detail pages, and success navigation into the application
  page with one `Start first task` action.
- [ ] Run targeted API and web tests, then commit with
  `feat(applications): continue enabled apps into first use`.

### Task 6: Switch Canonical Routes And Deploy

**Files:**
- Modify: `deploy/kubernetes/cluster-oilan/32-web.yaml`
- Modify: `deploy/kubernetes/cluster-oilan/38-marketplace.yaml`
- Modify: `deploy/kubernetes/cluster-oilan/40-ingress.yaml`
- Modify: `deploy/kubernetes/cluster-oilan/kustomization.yaml`
- Modify: `clients/web/src/app/(dashboard)/[org]/marketplace/page.tsx`
- Modify: `clients/web/src/components/ide/ActivityBar.tsx`
- Modify: `clients/web/src/lib/ide-route.ts`
- Test: `clients/web/src/lib/ide-route.test.ts`

- [ ] Write a failing route test showing `/{org}/marketplace` is no longer a
  dashboard activity and `/{org}/applications` resolves to the organization
  application center.
- [ ] Re-enable the standalone marketplace deployment at `market.l8ai.cn`,
  remove the permanent redirect, configure public Core Web acquisition URL, and
  redirect obsolete dashboard market routes to the public storefront.
- [ ] Run `pnpm` typechecks, marketplace-web tests/build, `go test
  ./marketplace/...`, Kustomize rendering, and image release verification.
- [ ] Build online through doops using official registries, update immutable
  GitOps image digests, push, wait for rollout, check health endpoints, and
  perform browser verification for anonymous discovery and authenticated
  enablement preflight.
- [ ] Commit release manifests with `chore(oilan): publish public app marketplace`.

## Plan Self-Review

- The six tasks cover the public route, taxonomy, filtering, evaluation,
  acquisition, enablement continuation, organization management, tests, and
  GitOps deployment.
- Public API changes precede both frontends, so the contract is implemented
  once and consumed by both surfaces.
- Unsupported component runtimes remain explicitly unavailable; the plan does
  not invent fallback acquisition behavior.

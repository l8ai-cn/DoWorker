# Video Expert Marketplace Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship tag-based Skill grouping, a reviewed immutable Expert Marketplace, and the first operator-owned video production, editing, and directing experts on an executable video runtime.

**Architecture:** Git-backed Skill metadata remains authoritative while PostgreSQL indexes normalized tags. Marketplace applications own stable identity and immutable release snapshots; installation and upgrades copy an approved release into an organization-owned Expert. A dedicated `video-studio` Worker type supplies FFmpeg, Chromium, Remotion, Python media tools, and CJK fonts.

**Tech Stack:** Go, Gin, GORM, PostgreSQL, Rust/WASM DTOs, Next.js, TypeScript, Vitest, Docker, FFmpeg, Remotion, Playwright.

---

### Task 1: Persist and synchronize Skill tags

**Files:**
- Create: `backend/migrations/000207_skill_tags.up.sql`
- Create: `backend/migrations/000207_skill_tags.down.sql`
- Create: `backend/migrations/skill_tags_test.go`
- Modify: `backend/internal/domain/skill/skill.go`
- Modify: `backend/internal/service/skill/authoring.go`
- Modify: `backend/internal/service/skill/authoring_test.go`
- Modify: `backend/internal/service/skill/upstream_sync.go`
- Modify: `backend/internal/service/extension/skill_importer_scan.go`
- Modify: Skill repository implementation returned by `rg "func .*ListCatalog" backend/internal`

- [ ] Write failing migration/model tests asserting `tags TEXT[] DEFAULT '{}'`, a GIN index, JSON output, lowercase trim/dedup normalization, and `skill.json` schema `2`.
- [ ] Run `go test ./backend/migrations ./backend/internal/service/skill/...` and confirm failures mention the missing column/schema.
- [ ] Add `Tags pq.StringArray`, `NormalizeTags([]string)`, create/update request fields, and schema-2 rendering:

```go
type skillConfig struct {
    Schema int `json:"schema"`
    Slug string `json:"slug"`
    Tags []string `json:"tags,omitempty"`
}
```

- [ ] Preserve catalog tags during imported upstream synchronization; never replace curator tags with upstream frontmatter.
- [ ] Run the focused Go tests and `gofmt` changed Go files; expect PASS.
- [ ] Commit: `feat(skills): add tag metadata`

### Task 2: Expose tag editing and grouping contracts

**Files:**
- Modify: `backend/internal/api/rest/v1/skill_handler_types.go`
- Modify: `backend/internal/api/rest/v1/skill_handler.go`
- Modify: `clients/core` Skill DTO/service files located with `rg "SkillCatalog" clients/core`
- Modify: `clients/web/src/lib/api/skillCatalogApi.ts`
- Modify: `clients/web/src/components/settings/organization/extensions/SkillCatalogSettings.tsx`
- Create: `clients/web/src/components/settings/organization/extensions/SkillTagEditor.tsx`
- Create: `clients/web/src/components/settings/organization/extensions/SkillTagFilters.tsx`
- Modify: `clients/web/src/components/settings/organization/extensions/CatalogSkillList.tsx`
- Create: `clients/web/src/components/settings/organization/extensions/__tests__/SkillCatalogTags.test.tsx`

- [ ] Read `frontend-design-skill` data-admin, engineering, product, and QA references before editing UI.
- [ ] Write failing handler, Rust contract, API conversion, and component tests for update, multi-tag filter, grouped view, untagged group, saving, empty, and error states.
- [ ] Run focused Go, Cargo, and Vitest commands; confirm missing `tags` failures.
- [ ] Implement the smallest contract and UI: tags are editable chips, filters are multi-select, and grouping is a flat/tag segmented control.
- [ ] Verify tag edits leave Expert `skill_slugs` and WorkerSpec snapshots unchanged with a backend regression test.
- [ ] Run focused tests, `pnpm run build:wasm`, `pnpm run web:typecheck`, and `pnpm run web:lint`; expect PASS.
- [ ] Commit: `feat(skills): manage catalog groups by tag`

### Task 3: Add immutable marketplace persistence

**Files:**
- Create: `backend/migrations/000208_expert_marketplace.up.sql`
- Create: `backend/migrations/000208_expert_marketplace.down.sql`
- Create: `backend/migrations/expert_marketplace_test.go`
- Create: `backend/internal/domain/expertmarket/application.go`
- Create: `backend/internal/domain/expertmarket/release.go`
- Create: `backend/internal/domain/expertmarket/repository.go`
- Create repository files under the existing GORM infrastructure package found with `rg "type .*Repository struct" backend/internal/infra`
- Modify: `backend/internal/domain/expert/expert.go`

- [ ] Write failing tests for identifier checks, unique `(application_id, version)`, lifecycle constraints, JSONB snapshots, review fields, and installed Expert source IDs.
- [ ] Run migration/domain tests and confirm the schema is absent.
- [ ] Implement application identity plus release statuses `draft`, `pending_review`, `published`, `rejected`, `withdrawn`; store expert, WorkerSpec, Skill dependency, and presentation snapshots.
- [ ] Add nullable `source_market_application_id` and `source_market_release_id` to Experts with organization/application installation uniqueness.
- [ ] Run focused migration and repository tests; expect PASS.
- [ ] Commit: `feat(marketplace): persist expert releases`

### Task 4: Implement submission, review, install, and upgrade services

**Files:**
- Delete: hard-coded catalog body from `backend/internal/service/expert/marketplace.go`
- Create: `backend/internal/service/expert/market_submission.go`
- Create: `backend/internal/service/expert/market_review.go`
- Create: `backend/internal/service/expert/market_installation.go`
- Create: `backend/internal/service/expert/market_queries.go`
- Create: `backend/internal/service/expert/marketplace_test.go`
- Modify: `backend/internal/service/expert/service.go`
- Modify: expert and Skill repository interfaces only where required

- [ ] Write failing service tests for snapshot immutability, inactive/non-platform dependency rejection, rejection reason, resubmission versioning, approval, withdrawal, idempotent install, upgrade discovery, and explicit upgrade.
- [ ] Run `go test ./backend/internal/service/expert/...` and confirm failures use the new service methods.
- [ ] Implement transactional transitions and return a typed dependency error containing sorted missing Skill slugs.
- [ ] On install, create an independent Expert bound to the release WorkerSpec snapshot and source IDs; on upgrade, replace release-owned fields only after an explicit request.
- [ ] Run service tests with race detection where supported; expect PASS.
- [ ] Commit: `feat(marketplace): add expert publishing workflow`

### Task 5: Add publisher, reviewer, and public APIs

**Files:**
- Modify: `backend/internal/api/rest/v1/routes_expert.go`
- Create: `backend/internal/api/rest/v1/expert_market_handler.go`
- Modify: `backend/internal/api/rest/v1/public_market_handler.go`
- Modify: `backend/internal/api/rest/v1/admin/routes.go`
- Create: `backend/internal/api/rest/v1/admin/expert_market_handler.go`
- Create: handler tests beside each handler
- Modify: `clients/web/src/lib/api/expertApi.ts`
- Create: `clients/web-admin/src/lib/api/adminExpertMarket.ts`

- [ ] Write failing route/handler tests for publisher submit/status/resubmit/withdraw, admin queue/detail/approve/reject, public published-only list/detail, install, upgrade availability, and upgrade.
- [ ] Run focused handler tests; confirm 404/409/422 mappings are absent.
- [ ] Implement tenant authorization, system-admin review authorization, typed validation responses, pagination, and published-only public responses.
- [ ] Add TypeScript clients and conversion tests matching the wire contracts exactly.
- [ ] Run Go tests plus web and web-admin API tests/typechecks; expect PASS.
- [ ] Commit: `feat(marketplace): expose expert release APIs`

### Task 6: Build marketplace operations UI

**Files:**
- Modify: `clients/web/src/components/experts/ExpertDetailPane.tsx`
- Create: `clients/web/src/components/experts/ExpertMarketSubmissionDialog.tsx`
- Create: `clients/web/src/components/experts/ExpertMarketStatus.tsx`
- Modify: `clients/web/src/components/marketplace/MarketplaceApplicationBrowser.tsx`
- Modify: `clients/web/src/components/marketplace/MarketplaceApplicationCard.tsx`
- Modify: `clients/web/src/components/marketplace/MarketplaceInstallButton.tsx`
- Create: `clients/web/src/components/marketplace/MarketplaceUpgradeButton.tsx`
- Modify: `clients/web-admin/src/components/layout/sidebar.tsx`
- Create: `clients/web-admin/src/app/(dashboard)/expert-reviews/page.tsx`
- Create: focused components under `expert-reviews/_components/`
- Create: Vitest tests beside the new user and admin flows

- [ ] Write failing UI tests for submission metadata, dependency error, pending/rejected states, admin rejection reason, approval, install provenance, and explicit upgrade confirmation.
- [ ] Run focused Vitest suites and confirm expected missing-component failures.
- [ ] Implement compact operational screens with status badges, dependency lists, loading/empty/error/disabled states, and no nested cards.
- [ ] Run web/web-admin tests, lint, typecheck, and production builds; expect PASS.
- [ ] Commit: `feat(marketplace): add publishing and review screens`

### Task 7: Add the video runtime and curated Skills

**Files:**
- Modify: `docker/agent-runtime/Dockerfile`
- Modify: `docker/agent-runtime/build.sh`
- Create: `docker/agent-runtime/video_contract_test.sh`
- Modify: `backend/internal/domain/workerruntime/catalog.go`
- Modify: `backend/internal/domain/workerruntime/catalog_test.go`
- Modify: local runner compose/Kubernetes manifests located with `rg "runner-codex-cli" deploy/dev deploy/kubernetes/local`
- Create: operator Skill seeds for `video-delivery-qa` and `video-storyboard-director` in the repository's existing seed mechanism
- Import only license-verified GitHub Skills: `remotion-best-practices`, `video-use`, `image2`, `motion-layer-animation`

- [ ] Write failing catalog and shell contract tests for `video-studio`, FFmpeg `libass`, Chromium, Node/Remotion, Python, and Noto CJK.
- [ ] Run catalog and shell tests; confirm the runtime is missing.
- [ ] Add an isolated video build target based on `codex-cli`; do not add video packages to other runtime images.
- [ ] Build the image and run an ASS subtitle burn plus a minimal 1080x1920 Remotion render; probe the MP4 with `ffprobe`.
- [ ] Confirm every redistributed external Skill has an explicit compatible license; exclude `remotion-video-director` unless verified.
- [ ] Commit: `feat(video): add studio runtime and curated skills`

### Task 8: Publish the first operator-owned expert suite

**Files:**
- Create: deterministic seed/release migration or bootstrap files following the existing platform seed pattern
- Create: browser tests under the repository's existing Playwright/E2E location
- Modify: user-visible marketplace documentation or release notes

- [ ] Seed three source Experts using `video-studio` and explicit Skill slugs: production, editing, and directing.
- [ ] Submit version `1`, approve as platform reviewer, mark the three applications operator-owned, and feature the video production expert.
- [ ] Run backend, Cargo/WASM, web, web-admin, runtime contract, and migration suites.
- [ ] Start the isolated dev stack and execute browser paths: tag/group, submit/reject/resubmit/approve, public install, and explicit upgrade; capture screenshots and check console/network errors.
- [ ] Launch the installed production expert and render a playable 9:16 MP4 with burned Chinese subtitles.
- [ ] Run independent requirement and code-quality reviews; fix findings and rerun affected verification.
- [ ] Commit: `feat(video): publish operator expert suite`
- [ ] Push `codex/video-expert-marketplace`, verify the remote contains the full commit SHA, and confirm CI checks.
- [ ] After the user confirms the target environment, update GitOps image digests/config, verify rollout health, and rerun the critical browser path before declaring release complete.

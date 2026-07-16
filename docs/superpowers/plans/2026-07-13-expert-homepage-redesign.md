# Expert Homepage Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the multi-role workforce homepage with an Expert-first product experience covering four product entries, composable capabilities, workflows, delivery, governance, marketplace, and the 12 supported Worker types.

**Architecture:** Add a focused `landing/expert-home` component group and make it the only homepage content composition. Keep marketing routes WASM-free, retain `useLightSession`, and load the dedicated `expert-home.json` namespace under `landing.workforce.expertHome` with an exact typed key contract across eight locales.

**Tech Stack:** Next.js App Router, React, TypeScript, Tailwind CSS, next-intl, Lucide React, Vitest, Testing Library, Playwright browser QA.

---

## File Map

- Create `clients/web/src/components/landing/expert-home/expert-home-content.ts`: solution, capability, workflow, market, safeguard, and Worker catalog IDs.
- Create `clients/web/src/components/landing/expert-home/expert-home-message-keys.ts`: exact production translation key list.
- Create `clients/web/src/components/landing/expert-home/ExpertHome.tsx`: section composition.
- Create `clients/web/src/components/landing/expert-home/ExpertHero.tsx`: headline, actions, and control-surface playback.
- Create `clients/web/src/components/landing/expert-home/ExpertControlSurface.tsx`: Expert assembly, workflow states, checkpoint, and deliverables.
- Create `clients/web/src/components/landing/expert-home/SolutionDomains.tsx`: four product menu panels.
- Create `clients/web/src/components/landing/expert-home/CapabilitySpectrum.tsx`: implemented, composable, and planned capability levels.
- Create `clients/web/src/components/landing/expert-home/ExpertOperatingModel.tsx`: Expert formula, workflow, human review, and delivery.
- Create `clients/web/src/components/landing/expert-home/ExpertMarketplace.tsx`: current market apps.
- Create `clients/web/src/components/landing/expert-home/ExpertGovernance.tsx`: safeguards and 12 Worker types.
- Create focused tests under `clients/web/src/components/landing/expert-home/__tests__/`.
- Modify `clients/web/src/app/page.tsx`, `Navbar.tsx`, `FinalCTA.tsx`, `Footer.tsx`, `globals.css`, i18n configuration, and workforce i18n tests; add eight `expert-home.json` files.

### Task 1: Lock Content Contracts

- [x] Add typed IDs for four solutions, ten capability groups, four workflow stages, four safeguards, three current market apps, and the 12 catalog Worker slugs.
- [x] Add `requiredExpertHomeMessageKeys` generated from the typed IDs plus fixed Hero, Expert, Workflow, Market, Trust, and CTA keys.
- [x] Replace `workforceMessages.test.ts` imports and assertions so every locale must contain exactly the new production key set.
- [x] Preserve the Chinese term `Workflow 调度`; remove the obsolete channel-slug assertion because the new homepage has no Channel fragment.
- [x] Run the focused message contract test.

### Task 2: Add Eight Locale Files

- [x] Add each `messages/{en,zh,ja,ko,es,fr,de,pt}/expert-home.json` with the same structural keys.
- [x] Use the approved Chinese Hero:

```json
{
  "badge": "专家驱动的 AI 工作平台",
  "title": "把分散的 AI 能力，组织成真正完成工作的专家",
  "description": "在一个工作空间里连接模型、Worker、Skill、知识与业务系统，从目标拆解、执行协作、人工确认到结果交付，打通完整工作链路。",
  "primaryAction": "创建 Expert",
  "secondaryAction": "浏览专家市场"
}
```

- [x] Translate equivalent meaning for all seven other locales without adding locale-only keys.
- [x] Run the workforce message test; all exact-key and non-empty-value assertions pass.

### Task 3: Build Expert Hero and Control Surface

- [x] Add deterministic pause, resume, next-step, and replay behavior.
- [x] Test the single-Expert narrative, workflow state, review checkpoint, and replay behavior.
- [x] Implement `ExpertControlSurface` with semantic Lucide controls and stable dimensions.
- [x] Implement `ExpertHero` actions without gradient text or decorative orbs.
- [x] Run the Hero tests and verify they pass.

### Task 4: Build Product and Capability Sections

- [x] Test all four solution anchors and deliverables.
- [x] Implement `SolutionDomains` with accessible tabs and four stable anchors.
- [x] Test capability labels for `implemented`, `composable`, and `planned`.
- [x] Implement the complete `CapabilitySpectrum`.
- [x] Run the section tests and verify selection and fact-level labels.

### Task 5: Build Operating Model, Market, and Governance

- [x] Test the formula `Expert = Worker + Model + Skills + Knowledge + Tools + Workflow`.
- [x] Implement `ExpertOperatingModel` with triggers, budgets, stop conditions, human takeover, and inspectable delivery.
- [x] Implement `ExpertMarketplace` with the three current applications and planned capability labels.
- [x] Render exactly the 12 Worker types and four safeguards in `ExpertGovernance`.
- [x] Run focused tests and verify all market and Worker names.

### Task 6: Replace Homepage Shell

- [x] Render `ExpertHome`, remove `PricingSection`, and remove JSON-LD `offers`.
- [x] Update JSON-LD description and keywords.
- [x] Update navigation for the four product entries, capabilities, and docs.
- [x] Update the final CTA and remove pricing claims.
- [x] Update footer product links and tagline.
- [x] Add narrowly scoped `.expert-home-*` tokens and remove active orb usage.

### Task 7: Regression and Browser QA

- [x] Keep required behavior in production homepage component tests.
- [x] Run lint, typecheck, focused tests, full tests, and production build.
- [x] Run the no-WASM helper; its marketing negative check passes, while its production positive assertion depends on an unminified symbol. Verify the Dashboard/Popout source import chain separately.
- [x] Test `/` at 1440×1000, 1024×900, and 390×844.
- [x] Verify navigation, solution selection, pause/resume/replay, overflow, overlap, and console state.
- [x] Capture desktop and mobile screenshots and compare them with the approved V2 storyboard.
- [x] Review `git diff` and preserve unrelated dirty-worktree changes.

## Delivery Constraint

This shared worktree already contains unrelated user changes. Implementation will not create commits or stage files unless the user explicitly requests a Git delivery action.

# Agent Supply and OPC Marketing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reframe the Agent Cloud marketing site around enterprise Agent supply, OPC incubation, and higher-education digital employees.

**Architecture:** Reuse the existing marketing shell, expert-home visual tokens, and focused landing components. Replace the Expert-first content model with a supply-first model, consolidate the story routes into Product and Solutions, and preserve backend resource names.

**Tech Stack:** Next.js App Router, React, TypeScript, Tailwind CSS, next-intl, Vitest, Testing Library.

---

### Task 1: Product Content Contract

**Files:**
- Modify: `clients/web/src/messages/en/expert-home.json`
- Modify: `clients/web/src/messages/zh/expert-home.json`
- Modify: `clients/web/src/messages/en/landing.json`
- Modify: `clients/web/src/messages/zh/landing.json`
- Modify: `clients/web/src/components/landing/expert-home/expert-home-content.ts`
- Modify: `clients/web/src/components/landing/expert-home/expert-home-message-keys.ts`

- [ ] Replace Expert-first hero and section copy with Agent supply terminology.
- [ ] Define exactly three solution entries: enterprise supply, OPC incubation, and higher-education digital employees.
- [ ] Remove Marketplace from the solution taxonomy.
- [ ] Keep English as the configured fallback for untranslated locales and provide complete Chinese content.
- [ ] Update message-key tests to require the production key set.

### Task 2: Homepage Composition

**Files:**
- Modify: `clients/web/src/components/landing/expert-home/ExpertHome.tsx`
- Modify: `clients/web/src/components/landing/expert-home/ExpertHero.tsx`
- Modify: `clients/web/src/components/landing/expert-home/ExpertControlSurface.tsx`
- Modify: `clients/web/src/components/landing/expert-home/SolutionDomains.tsx`
- Modify: `clients/web/src/components/landing/expert-home/ExpertOperatingModel.tsx`
- Modify: `clients/web/src/components/landing/expert-home/CapabilitySpectrum.tsx`
- Modify: `clients/web/src/components/landing/expert-home/ExpertMarketplace.tsx`
- Modify: `clients/web/src/components/landing/expert-home/ExpertGovernance.tsx`
- Modify: `clients/web/src/components/landing/FinalCTA.tsx`
- Modify: `clients/web/src/components/landing/Footer.tsx`
- Modify: `clients/web/src/app/page.tsx`

- [ ] Make the hero show an Agent supply network and goal-to-delivery path.
- [ ] Present the six-stage supply lifecycle before the solution directions.
- [ ] Present product foundations as Agent factory, market, workspace, automation, and governance.
- [ ] Keep market applications and trust controls aligned with implemented product capabilities.
- [ ] Remove all remaining pricing and free-tier claims from the homepage.
- [ ] Update homepage structured data and metadata language.

### Task 3: Product and Solutions Pages

**Files:**
- Create: `clients/web/src/app/product/page.tsx`
- Create: `clients/web/src/app/product/layout.tsx`
- Modify: `clients/web/src/app/how-it-works/page.tsx`
- Modify: `clients/web/src/app/solutions/page.tsx`
- Modify: `clients/web/src/components/landing/marketing-routes.ts`
- Modify: `clients/web/src/components/landing/expert-pages/marketing-page-config.ts`
- Modify: `clients/web/src/components/landing/expert-pages/MarketingPageHero.tsx`
- Modify: `clients/web/src/app/__tests__/ExpertMarketingPages.test.tsx`

- [ ] Add `/product` as the dedicated product page.
- [ ] Redirect `/how-it-works` to `/product`.
- [ ] Keep `/solutions` as the dedicated three-direction solution page.
- [ ] Reduce first-level navigation to Home, Product, Solutions, Agent Market, and Documentation.
- [ ] Update route composition tests for the new information architecture.

### Task 4: Static Verification

**Files:**
- Test: `clients/web/src/lib/i18n/__tests__/workforceMessages.test.ts`
- Test: `clients/web/src/lib/i18n/__tests__/marketingPricingMessages.test.ts`
- Test: `clients/web/src/components/landing/expert-home/__tests__/*`
- Test: `clients/web/src/app/__tests__/ExpertMarketingPages.test.tsx`

- [ ] Run focused Vitest tests for marketing routes, content, and navigation.
- [ ] Run `pnpm run web:lint`.
- [ ] Run `pnpm run web:typecheck`.
- [ ] Run `bash clients/web/scripts/check-no-wasm-in-marketing.sh` after a production build when the worktree can build.

### Task 5: Browser Acceptance

**Routes:**
- `/`
- `/product`
- `/solutions`
- `/marketplace`
- `/docs`

- [ ] Verify desktop at 1440x900.
- [ ] Verify mobile at 390x844.
- [ ] Exercise navigation, mobile menu, solution tabs, and primary calls to action.
- [ ] Check loading overlays, browser console errors, failed requests, text clipping, and horizontal overflow.
- [ ] Capture screenshots for the homepage and solutions page.

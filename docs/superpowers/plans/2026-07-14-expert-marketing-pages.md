# Expert Marketing Pages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Replace homepage hash navigation with an explicit Home item plus five independent marketing destinations, and remove public marketing pricing from rendered pages and localized payloads.

**Architecture:** Keep marketing navigation in one route definition, compose new pages from a shared shell and hero, and reuse the existing translated Expert sections as the content source. Preserve the marketplace data flow and specialized documentation shell.

**Tech Stack:** Next.js App Router, React 19, TypeScript, Tailwind CSS, next-intl, Vitest, Testing Library, Playwright CLI.

---

### Task 1: Lock Navigation And Homepage Behavior

**Files:**
- Modify: `clients/web/src/components/landing/__tests__/Navbar.test.tsx`
- Create: `clients/web/src/app/__tests__/HomePage.test.tsx`
- Create: `clients/web/src/components/landing/marketing-routes.ts`
- Modify: `clients/web/src/components/landing/Navbar.tsx`
- Modify: `clients/web/src/components/landing/expert-home/ExpertHero.tsx`
- Modify: `clients/web/src/app/page.tsx`

- [x] Change the Navbar test to require `/`, `/solutions`, `/how-it-works`, `/capabilities`, `/marketplace`, and `/docs`, and assert that Pricing is absent.
- [x] Add a Home page regression test that mocks `PricingSection` and proves the homepage no longer renders it.
- [x] Run the two tests and confirm they fail against the current hash links and Pricing section.
- [x] Add the six-item route SSOT and consume it from Navbar.
- [x] Point the homepage hero secondary action to `/solutions`.
- [x] Remove `PricingSection` from the homepage.
- [x] Run the two tests and confirm they pass.

### Task 2: Add Shared Independent Page Structure

**Files:**
- Create: `clients/web/src/components/landing/expert-pages/MarketingPageShell.tsx`
- Create: `clients/web/src/components/landing/expert-pages/MarketingPageHero.tsx`
- Create: `clients/web/src/components/landing/expert-pages/marketing-page-config.ts`
- Create: `clients/web/src/components/landing/expert-pages/__tests__/MarketingPageHero.test.tsx`
- Modify: `clients/web/src/components/landing/expert-home/SolutionDomains.tsx`
- Modify: `clients/web/src/components/landing/expert-home/ExpertOperatingModel.tsx`
- Modify: `clients/web/src/components/landing/expert-home/CapabilitySpectrum.tsx`

- [x] Add a failing hero test for translated title, description, primary action, and next-page link.
- [x] Run the test and confirm the shared page hero is not implemented.
- [x] Add typed page definitions for solutions, operating model, and capabilities using existing translation keys.
- [x] Implement the shared shell and hero using existing Expert tokens and responsive constraints.
- [x] Add `showIntro?: boolean` to the three reused sections so subpages have one H1 and no duplicated introduction.
- [x] Run the focused component tests and confirm they pass.

### Task 3: Create The Three New Routes

**Files:**
- Create: `clients/web/src/app/solutions/layout.tsx`
- Create: `clients/web/src/app/solutions/page.tsx`
- Create: `clients/web/src/app/how-it-works/layout.tsx`
- Create: `clients/web/src/app/how-it-works/page.tsx`
- Create: `clients/web/src/app/capabilities/layout.tsx`
- Create: `clients/web/src/app/capabilities/page.tsx`
- Create: `clients/web/src/app/__tests__/ExpertMarketingPages.test.tsx`

- [x] Add failing route composition tests for the hero and correct content regions on all three pages.
- [x] Run the tests and confirm the route files are absent.
- [x] Add metadata layouts with canonical URLs and page-specific descriptions.
- [x] Compose Solutions from `MarketingPageHero` and `SolutionDomains`.
- [x] Compose How It Works from `MarketingPageHero` and `ExpertOperatingModel`.
- [x] Compose Capabilities from `MarketingPageHero`, `CapabilitySpectrum`, and `ExpertGovernance`.
- [x] Run the route tests and confirm they pass.

### Task 4: Align Existing Destinations

**Files:**
- Modify: `clients/web/src/components/marketplace/MarketplaceHeader.tsx`
- Modify: `clients/web/src/app/marketplace/page.tsx`
- Test: `clients/web/src/components/landing/__tests__/Navbar.test.tsx`

- [x] Add a failing assertion that the marketplace header exposes Home and the five marketing destinations.
- [x] Run the focused test and confirm the existing two-link header fails it.
- [x] Reuse the shared Navbar in the marketplace header and provide the Expert token scope required by it.
- [x] Preserve the marketplace loading, empty, error, filtering, and install behavior.
- [x] Run the focused tests and confirm they pass.

### Task 5: Verify The Complete Experience

**Files:**
- Verify all files changed by Tasks 1-4.

- [x] Run focused Vitest tests for navigation, homepage, hero, and route composition.
- [x] Run `pnpm run web:typecheck`.
- [x] Run `pnpm run web:test`.
- [x] Run ESLint for the changed frontend files.
- [x] Start the web development server on an available port.
- [x] Use Playwright to validate `/`, `/solutions`, `/how-it-works`, `/capabilities`, `/marketplace`, and `/docs` at desktop and mobile widths.
- [x] Confirm links, menu closing, responsive wrapping, page identity, absence of Pricing, and lack of horizontal overflow.
- [x] Inspect console and network output; distinguish unavailable local backend services from frontend regressions.
- [x] Run targeted `git diff --check` and review the final diff without altering unrelated worktree changes.

### Final Review Corrections

- Remove the retired pricing component, enterprise quote block, public pricing terms, and pricing keys from every marketing locale.
- Keep H1 → H2 → H3 semantics when reused sections hide their visible introductions.
- Publish the five content destinations through the sitemap and expose Home plus all five through the documentation shell.
- Mark active navigation with `aria-current` across marketing and documentation pages.
- Distinguish organization discovery states so network failures remain on the current page with a visible error.
- Verify 390px, 1100px, and 1440px layouts, including a scrollable mobile documentation drawer.
- Record the unrelated Worker documentation catalog and `WorkerTypeConfigStep.tsx` typecheck blockers.
- Remove billing FAQ content, price-bearing download JSON-LD, free-tier claims, and free-start CTA copy from public locale payloads.

# Expert Marketing Pages Design

## Goal

Turn the primary menu into an explicit homepage entry plus five real destinations:

- `/`
- `/solutions`
- `/how-it-works`
- `/capabilities`
- `/marketplace`
- `/docs`

The homepage remains the product overview. Public marketing pricing is removed from rendered pages, navigation, enterprise messaging, and localized marketing payloads.

## Product Frame

The visitor is evaluating whether Agent Cloud can own cross-functional work. Each page answers one decision:

- **Solutions:** Which business outcomes can an Expert carry?
- **How it works:** How does one Expert coordinate capabilities and human control?
- **Capabilities:** What can the platform execute, compose, and govern?
- **Marketplace:** Which verified Experts can be reused now?
- **Docs:** How is the product installed, configured, and operated?

## Information Architecture

The logo and an explicit Home menu item link to `/`. The primary navigation uses independent routes rather than homepage hashes.

The homepage keeps the existing Expert hero and overview sections. Its secondary hero action links to `/solutions`. Detailed pages reuse the same translated content and section components so marketing claims have one source of truth.

### Solutions

The page introduces outcome-first adoption, then renders the four existing solution domains with delivery chains and inspectable outcomes. Its final action lets the visitor create an Expert or continue to the operating model.

### How It Works

The page explains the six layers behind an Expert, continuous execution, human checkpoints, evidence-backed delivery, and reuse of validated practice.

### Capabilities

The page distinguishes implemented, composable, and planned capability levels. Governance follows capability breadth so the page does not imply unlimited permissions.

### Marketplace

The existing dynamic catalog remains authoritative. Its header joins the shared six-item marketing navigation while preserving loading, empty, and error states. Authenticated console access distinguishes an empty organization list from an unavailable organization lookup.

### Docs

The existing documentation shell remains specialized for reading and navigation. It exposes the six-item marketing navigation at desktop breakpoints and inside the scrollable mobile documentation drawer.

## Component Boundaries

- `marketing-routes.ts` is the navigation route SSOT.
- `MarketingPageShell.tsx` composes Navbar, page content, CTA, and Footer.
- `MarketingPageHero.tsx` provides a consistent subpage introduction using existing translation keys.
- Existing Expert homepage sections accept an optional intro visibility prop so subpages avoid duplicate headings.
- Route files only provide metadata and compose page regions.

## Pricing Removal

The Navbar no longer renders pricing. `PricingSection` and the enterprise quote block are deleted, retired pricing and free-tier keys are removed from all public marketing locale bundles and FAQ content, and public terms no longer claim that current prices are published on the website. Download structured data does not publish an offer or price. No `/pricing` route is created. Existing billing code and authenticated product billing remain untouched.

## Validation

- Regression tests lock Home plus the five content destinations and absence of pricing.
- Route tests verify the three new pages render their correct content regions.
- Type checking and the full web test suite are run; unrelated worktree blockers are documented rather than hidden.
- Browser checks cover homepage plus all five destinations at desktop and mobile widths.
- The browser must show no relevant framework error, clipping, horizontal overflow, or broken navigation.

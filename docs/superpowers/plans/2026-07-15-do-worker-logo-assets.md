# Do Worker Logo Assets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the rejected keystone mark with the approved Do Worker capability-core mark across the web app, PWA icons, favicon, and documentation.

**Architecture:** Use four unequal capability modules in an offset bond so the combined silhouette itself is the Expert. Keep the React component and static SVG geometry aligned, provide transparent and monochrome variants plus a dark app-icon container, and generate every raster derivative from the vector source.

**Tech Stack:** SVG, React, TypeScript, Sharp, Pillow, Vitest, Next.js, in-app browser.

---

### Task 1: Explore and Lock the Capability-Core Contract

**Files:**
- Create: `clients/web/src/components/common/__tests__/Logo.test.tsx`
- Modify: `clients/web/src/components/common/Logo.tsx`

- [x] Generate Image2 explorations of four-way capability interlock and non-radial composition.
- [x] Reject variants that read as a letter, flower, cross, gear, network, gem, or fractured object.
- [x] Update the component test to require four capability modules and no standalone Expert core.
- [x] Run the focused test, confirm the rejected keystone fails, and make the offset-bond component pass.

### Task 2: Create Vector Brand Assets

**Files:**
- Create: `clients/web/public/brand/do-worker-mark.svg`
- Create: `clients/web/public/brand/do-worker-mark-mono.svg`
- Create: `clients/web/public/brand/do-worker-app-icon.svg`
- Modify: `clients/web/public/icons/icon.svg`
- Modify: `docs/images/logo.svg`

- [x] Draw the selected four-module capability-core mark as deterministic SVG geometry.
- [x] Draw a one-color mark whose module interlock remains identifiable without color.
- [x] Draw the app icon with a solid `#0B0F14` rounded-square background and the same geometry.
- [x] Copy the app-icon vector contract to the PWA icon and documentation logo assets.
- [x] Validate every SVG as XML and scan for gradients, filters, text, embedded bitmaps, and scripts.

### Task 3: Generate Raster Assets

**Files:**
- Create: `clients/web/public/brand/do-worker-mark.png`
- Create: `clients/web/public/brand/do-worker-app-icon.png`
- Modify: `clients/web/public/icons/icon-192.png`
- Modify: `clients/web/public/icons/icon-512.png`
- Modify: `clients/web/src/app/favicon.ico`

- [x] Use bundled Sharp to render the transparent mark and app icon at 1024x1024.
- [x] Use the same SVG source to render exact 192x192 and 512x512 PWA icons.
- [x] Use Pillow to package 16x16, 32x32, 48x48, and 64x64 favicon layers from the app icon.
- [x] Inspect dimensions, alpha channels, file signatures, and the rendered 16px favicon.

### Task 4: Verify Product Integration

**Files:**
- No additional production files unless verification exposes a defect.

- [x] Run the focused Logo component test and targeted ESLint.
- [x] Run the full Web TypeScript checking command.
- [x] Open the solutions page in the in-app browser and check the navigation and footer on desktop.
- [x] Check the mobile navigation and favicon-sized rendering.
- [x] Inspect browser console errors and capture screenshots showing the final logo in context.

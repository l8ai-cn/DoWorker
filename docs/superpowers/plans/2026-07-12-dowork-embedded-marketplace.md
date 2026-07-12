# DoWorker Embedded Marketplace Implementation Plan

**Goal:** close Marketplace inside the DoWorker organization workflow while
retaining independent Marketplace API data ownership.

## Acceptance Scenarios

1. Given an authenticated member opens `/{org}/marketplace`, when the catalog
   loads, then the activity bar highlights “市场” and content is Chinese.
2. Given a Listing is an application, when enablement is chosen, then the plan
   targets URL organization and confirmation exposes credits and permissions.
3. Given a Listing is not runtime-installable, when a member views it, then the
   UI states the missing integration and never presents a succeeding action.
4. Given a legacy market URL, when visited, then it redirects to canonical
   DoWorker organization market.

## Steps And Checks

1. Add market activity and nested routes.
   - Change activity union, resolver, activity bar, and sidebar.
   - Verify `resolveActivityFromPathname`.
2. Build dashboard catalog and detail.
   - Add authenticated same-origin catalog API client and no Zustand catalog cache.
   - Verify request path and encoded detail identifier with Vitest.
3. Bind acquisition to organization route.
   - Resolve matching organization membership and omit picker for nested route.
   - Verify plan/apply API unit tests and success destination.
4. Replace public traffic.
   - Redirect root compatibility page internally.
   - Delete standalone workload and redirect `market.l8ai.cn` with ingress.
   - Update release verification and deployment scripts.
5. Validate and publish.
   - Run Vitest, lint, TypeScript, production build, Kustomize render, doops
     deploy/health, and authenticated browser smoke test.

# Agent Workbench Phase 0B Preview Security Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move live Pod content to a dedicated origin and make preview bootstrap credentials single-use.

**Architecture:** Backend issues a five-minute bootstrap JWT. Relay validates its exact origin and atomically consumes its JTI through a Redis-backed backend endpoint before setting a separate 15-minute HttpOnly session cookie.

**Tech Stack:** Go, Gin, Relay HTTP proxy, Redis, JWT, React, Vitest, Playwright.

---

### Task 1: Configure And Enforce The Preview Origin

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/internal/api/rest/v1/pods.go`
- Modify: `backend/internal/api/rest/v1/routes_pod_queue.go`
- Modify: `backend/internal/api/rest/v1/pod_preview.go`
- Modify: `backend/internal/api/rest/v1/pod_preview_test.go`
- Modify: `relay/internal/config/config.go`
- Modify: `relay/internal/config/config_test.go`
- Modify: `relay/internal/server/handler_preview.go`
- Modify: `relay/internal/server/handler_preview_test.go`
- Modify: `deploy/dev/lib/config_gen.sh`
- Modify: `backend/.env.example`
- Modify: `relay/.env.example`

- [ ] **Step 1: Add fail-closed tests**

```go
require.Equal(t, "https://preview.example.com", cfg.PreviewPublicOrigin)
require.Equal(t, http.StatusMisdirectedRequest, requestWithHost("app.example.com").Code)
require.Equal(t, "https://preview.example.com/preview/pod1/", response.PreviewBaseURL)
```

- [ ] **Step 2: Verify tests fail**

Run: `go test ./backend/internal/config ./backend/internal/api/rest/v1 ./relay/internal/config ./relay/internal/server`
Expected: FAIL because `PREVIEW_PUBLIC_ORIGIN` and Host enforcement do not exist.

- [ ] **Step 3: Implement explicit origin wiring**

Add required `PREVIEW_PUBLIC_ORIGIN`; production startup fails when it equals an authenticated application origin. `PodHandler` receives the parsed origin and constructs preview URLs only from it. Relay requires the exact configured host. Dev generates `http://preview.localhost:${HTTP_PORT}` and Traefik exposes only `/preview/*` on that host.

- [ ] **Step 4: Run tests and commit**

Run: `go test ./backend/internal/config ./backend/internal/api/rest/v1 ./relay/internal/config ./relay/internal/server`
Expected: PASS.

```bash
git add backend/internal/config backend/internal/api/rest/v1/pods.go backend/internal/api/rest/v1/routes_pod_queue.go backend/internal/api/rest/v1/pod_preview.go backend/internal/api/rest/v1/pod_preview_test.go relay/internal/config relay/internal/server/handler_preview.go relay/internal/server/handler_preview_test.go deploy/dev/lib/config_gen.sh backend/.env.example relay/.env.example
git commit -m "feat(preview): isolate the public preview origin"
```

### Task 2: Single-Use Preview Bootstrap

**Files:**
- Create: `backend/internal/service/relay/preview_bootstrap_store.go`
- Create: `backend/internal/service/relay/preview_bootstrap_store_test.go`
- Modify: `backend/internal/service/relay/token.go`
- Modify: `backend/internal/service/relay/token_test.go`
- Create: `backend/internal/api/rest/internal/preview_bootstrap.go`
- Create: `backend/internal/api/rest/internal/preview_bootstrap_test.go`
- Modify: `backend/internal/api/rest/router.go`
- Modify: `relay/internal/backend/client.go`
- Modify: `relay/internal/backend/client_test.go`
- Modify: `relay/internal/auth/token.go`
- Modify: `relay/internal/auth/token_test.go`
- Modify: `relay/internal/server/handler_preview_session.go`
- Modify: `relay/internal/server/handler_preview_test.go`

- [ ] **Step 1: Write replay and cookie tests**

```go
require.NoError(t, store.Consume(ctx, jti, podKey))
require.ErrorIs(t, store.Consume(ctx, jti, podKey), ErrPreviewBootstrapConsumed)
require.True(t, cookie.HttpOnly)
require.True(t, cookie.Secure)
require.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
require.Equal(t, "/preview/pod1", cookie.Path)
```

- [ ] **Step 2: Verify failure**

Run: `go test ./backend/internal/service/relay ./backend/internal/api/rest/internal ./relay/internal/auth ./relay/internal/server`
Expected: FAIL because bootstrap JTI, token use, and atomic consumption are absent.

- [ ] **Step 3: Implement redemption**

Mint `preview_bootstrap` with audience, JTI, exact origin, pod, user, org, target, and path. Relay consumes JTI through the authenticated backend endpoint backed by Redis `SET NX`, then signs `preview_session` for the HttpOnly host-only cookie. `HandlePreview` accepts only `preview_session`.

- [ ] **Step 4: Run suites and commit**

Run: `go test ./backend/internal/service/relay ./backend/internal/api/rest/internal ./relay/internal/auth ./relay/internal/server`
Expected: PASS including concurrent redemption and cross-Pod rejection.

```bash
git add backend/internal/service/relay backend/internal/api/rest/internal backend/internal/api/rest/router.go relay/internal/backend relay/internal/auth relay/internal/server/handler_preview_session.go relay/internal/server/handler_preview_test.go
git commit -m "fix(preview): redeem bootstrap tokens exactly once"
```

### Task 3: Preview Frame Policy And Browser Verification

**Files:**
- Modify: `clients/web-user/src/components/PreviewPanel.tsx`
- Modify: `clients/web-user/src/components/PreviewPanel.test.tsx`
- Modify: `clients/web-user/src/hooks/usePodPreview.ts`
- Modify: `clients/web-user/src/hooks/usePodPreview.test.tsx`
- Create: `clients/web-user/e2e/preview-security.spec.ts`

- [ ] **Step 1: Add frame policy tests**

```ts
expect(frame).toHaveAttribute("sandbox", "allow-scripts allow-same-origin allow-forms allow-downloads");
expect(frame).toHaveAttribute("referrerpolicy", "no-referrer");
expect(frame).toHaveAttribute("allow", expect.not.stringContaining("camera"));
```

- [ ] **Step 2: Implement the `pod-live` descriptor**

`usePodPreview` validates both URLs against `PREVIEW_PUBLIC_ORIGIN`. `PreviewPanel` applies the exact sandbox, Permissions Policy, and referrer policy. New-window opening targets a trusted shell route with `noopener,noreferrer`.

- [ ] **Step 3: Run unit and browser tests**

Run: `pnpm --dir clients/web-user test -- PreviewPanel usePodPreview`
Expected: PASS.

Run: `pnpm --dir clients/web-user exec playwright test e2e/preview-security.spec.ts`
Expected: one successful redemption, replay 401, blocked popup/navigation/camera, no application cookies, and no console errors.

- [ ] **Step 4: Commit**

```bash
git add clients/web-user/src/components/PreviewPanel.tsx clients/web-user/src/components/PreviewPanel.test.tsx clients/web-user/src/hooks/usePodPreview.ts clients/web-user/src/hooks/usePodPreview.test.tsx clients/web-user/e2e/preview-security.spec.ts
git commit -m "fix(web-user): enforce the live preview security profile"
```

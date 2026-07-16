# Agent Embed Context Security Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** allow an authenticated session manager to mint a short-lived, origin-bound embed context that an iframe redeems into a restricted single-session access token.

**Architecture:** The backend creates two distinct signed JWT uses. `embed_context` is accepted only by inspect and proof-backed redemption endpoints; `embed_session` is accepted only by `/v1/embed/sessions/:id` routes. The parent retains an independent one-time proof, Redis atomically compares its hash before consuming the context, and normal authentication middleware rejects both embed token uses.

**Tech Stack:** Go, Gin, `golang-jwt/jwt/v5`, Testify.

---

## File Structure

| Path                                                         | Responsibility                                            |
| ------------------------------------------------------------ | --------------------------------------------------------- |
| `backend/pkg/embedtoken/token.go`                            | Context/session claims and service contract               |
| `backend/pkg/embedtoken/token_signing.go`                    | JWT signing and validation                                |
| `backend/pkg/embedtoken/context_store.go`                    | Proof generation, Redis inspection, and atomic redemption |
| `backend/pkg/embedtoken/token_test.go`                       | Token purpose, expiry, issuer, and claims tests           |
| `backend/internal/api/rest/v1/session/embed_context.go`      | Create, inspect, and redeem HTTP handlers                 |
| `backend/internal/api/rest/v1/session/embed_auth.go`         | Restricted token middleware and capability checks         |
| `backend/internal/api/rest/v1/session/embed_context_test.go` | Handler and restricted API authorization tests            |
| `backend/internal/api/rest/v1/session/routes.go`             | Normal and restricted route registration                  |
| `backend/internal/api/rest/v1/session/deps.go`               | Token service dependency                                  |
| `backend/internal/api/rest/router.go`                        | Session dependency wiring                                 |
| `backend/internal/middleware/auth.go`                        | Reject non-login token uses from normal endpoints         |

### Task 1: Define distinct token uses

**Files:**

- Create: `backend/pkg/embedtoken/token_test.go`
- Create: `backend/pkg/embedtoken/token.go`

- [x] Write a failing test that proves an `embed_context` validates only as a context, expires at five minutes, and redeems into an `embed_session` with the same session id, capabilities, and origin set.
- [x] Run `go test ./pkg/embedtoken -run TestContextRedeemsToRestrictedSession -count=1` and confirm it fails because the package does not exist.
- [x] Implement `Service.IssueContext`, `Service.RedeemContext`, and `Service.ValidateSession` with `token_use`, `session_id`, `org_id`, `org_slug`, `user_id`, capabilities, and allowed origins claims.
- [x] Run the package test and confirm it passes.

### Task 2: Prevent normal JWT acceptance

**Files:**

- Modify: `backend/internal/middleware/auth.go`
- Test: `backend/internal/middleware/auth_test.go`

- [x] Write a failing middleware test for a signed `embed_session` token sent to `AuthMiddleware`.
- [x] Confirm normal authentication returns 401 rather than exposing the token's user id.
- [x] Add a `token_use` claim to normal middleware claims and reject any non-empty use.
- [x] Run the focused middleware test and confirm it passes.

### Task 3: Create and redeem a session-scoped context

**Files:**

- Create: `backend/internal/api/rest/v1/session/embed_context.go`
- Create: `backend/internal/api/rest/v1/session/embed_context_test.go`
- Modify: `backend/internal/api/rest/v1/session/deps.go`
- Modify: `backend/internal/api/rest/v1/session/routes.go`
- Modify: `backend/internal/api/rest/router.go`

- [x] Write failing handler tests for:
  - a manager issuing an exact-origin read token;
  - malformed, wildcard, path-bearing, or duplicate origins returning 400;
  - an edit-only user receiving 403 because bearer delegation requires `levelManage`;
  - redeeming a context into an `embed_session`.
- [x] Register `POST /v1/sessions/:id/embed-context` behind normal session auth and `POST /v1/embed-contexts/redeem` without normal auth.
- [x] Require `levelManage`, validate exact `http`/`https` origins, require an explicit non-empty capability set, and issue a five-minute context.
- [x] Run the focused handler tests and confirm they pass.
- [x] Return a separate parent-held redemption proof, expose a non-consuming
      inspect endpoint, and atomically compare-and-delete the proof hash in Redis.
- [x] Prove a wrong proof does not consume a valid context and replay fails.

### Task 4: Restrict embedded session access

**Files:**

- Create: `backend/internal/api/rest/v1/session/embed_auth.go`
- Modify: `backend/internal/api/rest/v1/session/routes.go`
- Modify: `backend/internal/api/rest/v1/session/session_authz.go`
- Test: `backend/internal/api/rest/v1/session/embed_context_test.go`

- [x] Write failing tests proving an `embed_session` can read only its exact session, cannot read another session, and cannot post an event without `write`.
- [x] Implement embedded-token middleware that sets the tenant only from signed claims and stores the verified claims in Gin context.
- [x] Register only `GET /v1/embed/sessions/:id`, `GET /v1/embed/sessions/:id/items`, `GET /v1/embed/sessions/:id/stream`, and `POST /v1/embed/sessions/:id/events`.
- [x] Apply token session-id matching in `authorizeSession` and token capability matching before mutating routes.
- [x] Run `go test ./internal/api/rest/v1/session ./internal/middleware ./pkg/embedtoken -count=1` and confirm it passes.

### Task 5: Verify and document the iframe boundary

**Files:**

- Modify: `docs/superpowers/specs/2026-07-13-agent-conversation-integration-design.md`

- [x] Record the two-token flow and the currently exposed embedded API subset.
- [ ] Run `git diff --check` for changed paths.
- [x] Start the local backend and use an authenticated API test to confirm normal JWT routes reject embed tokens.

## Review Checklist

- [x] Context tokens cannot be used as normal user JWTs.
- [x] Session tokens cannot access routes outside `/v1/embed`.
- [x] No token expands access beyond one session, its issuing user, allowed capabilities, and expiry.
- [x] Parent origins are exact origins, not wildcards or URL prefixes.
- [x] A read-only token cannot send, interrupt, stop, or resolve permissions.

`git diff --check` is still red because unrelated concurrently generated
`proto/gen/ts/*` files have blank lines at EOF. The changed embed paths were
checked individually and have no whitespace errors.

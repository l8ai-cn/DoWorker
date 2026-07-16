# Agent Workbench Security Profiles Design

**Status:** normative appendix
**Date:** 2026-07-15
**Parent:** `2026-07-15-agent-workbench-v2-design.md`

## 1. Trust Domains

Authenticated application pages, static artifacts, live Pod previews, embedded workbenches, remote renderer frames, and terminal data are distinct trust domains. Sharing a product domain or Relay endpoint does not make their active content equivalent.

Production and development configuration must provide a dedicated preview origin that is not an origin used by Web, Web User, Web Admin, Backend cookies, or any authenticated API. Server routing rejects preview sessions on an authenticated application origin.

## 2. Static Artifact Profile

Untrusted HTML, SVG, and generated web artifacts default to `artifact-static`:

- opaque-origin sandbox without `allow-same-origin`;
- no scripts, forms, popups, top navigation, downloads without user activation, pointer lock, presentation, or modals;
- Content Security Policy denies network, frames, workers, objects, and form actions;
- referrer policy is `no-referrer`;
- source is delivered through a sandboxed shell, not opened as a top-level `text/html` Blob.

Opening a static artifact in a new window opens the same isolated shell on the dedicated artifact/preview origin. It never navigates directly to a Blob inheriting the authenticated application origin.

## 3. Interactive Preview Profile

`artifact-interactive` is a separate explicit capability for generated applications that need scripts or network access. It runs only on the dedicated preview origin, uses exact destination allowlists for `connect-src`, `img-src`, `media-src`, and frames, and denies forms, popups, top navigation, and browser APIs unless separately granted.

An artifact cannot promote itself from static to interactive. The backend issues the profile and policy after checking the resource and user grant.

Live Pod applications use a third `pod-live` profile on the dedicated preview origin. Its iframe sandbox is `allow-scripts allow-same-origin allow-forms allow-downloads`; popups, top navigation, modals, pointer lock, and presentation remain denied. Permissions Policy disables camera, microphone, geolocation, clipboard write, and sensors unless the session grant enables an exact feature. Referrer policy is `no-referrer`.

Opening a Pod preview in a new window opens a trusted shell that embeds the same exact preview URL with the same sandbox and Permissions Policy. The opener uses `noopener,noreferrer`; it never navigates the new top-level document directly to untrusted Pod content.

## 4. Markdown And Media

Markdown remote images, embeds, styles, and links that auto-fetch are blocked by default. A host may activate a resource through a backend proxy or exact allowlist after user action. Renderers do not load arbitrary remote URLs during initial render.

SVG is treated as active content. It is rasterized, served as a download, or rendered through a sandbox profile; raw SVG markup is never injected into the host DOM.

## 5. Preview Session Exchange

A backend-issued `preview_bootstrap` token has a short expiry, exact audience, `token_use=preview_bootstrap`, single-use `jti`, user/embed subject, session and resource scope, requested profile, and exact preview origin.

The preview origin atomically redeems the `jti`, then sets a browser-only `preview_session` cookie with `HttpOnly`, `Secure`, no `Domain` attribute, and `Path=/preview/{podKey}`. Same-site deployments use `SameSite=Strict`; an explicitly cross-site deployment must use `SameSite=None; Partitioned` and may not issue an unpartitioned third-party cookie. Browser endpoints never return the bearer credential to Pod JavaScript.

Trusted non-browser clients use a separately issued bearer token with a different audience and cannot exchange a browser bootstrap. Preview sessions expire within 15 minutes. Renewal requires a new authenticated bootstrap; pod stop, user logout, embed revocation, and grant revocation invalidate active session IDs.

Tests cover concurrent redemption, replay, expiry, wrong audience, wrong origin, wrong resource, cross-profile use, script readability, cookie path isolation, renewal, and revocation.

## 6. Remote Renderer Contract

A restricted remote renderer descriptor includes:

- renderer ID and protocol version;
- exact URL and origin;
- sandbox and Permissions Policy;
- CSP, network destination, navigation, and referrer policies;
- session, artifact, representation, and revision binding;
- maximum inbound and outbound message size.

The handshake uses a channel nonce. Every message includes channel ID, message ID, protocol version, bound session/artifact/revision, type, and payload. The host also requires `event.source === iframe.contentWindow` and the exact registered origin; the sender always uses the exact `targetOrigin`, never `"*"`. Both sides reject duplicate IDs, stale revisions, unknown types, oversized payloads, sibling-frame messages, and messages before handshake completion.

Remote frames receive capability declarations separately from server-issued authorization grants. They receive only scoped data and short-lived action grants, never host credentials.

## 7. Artifact Actions

An actionable descriptor states issuer, action type, required grant, target artifact and representation, target/base revision, risk class, confirmation requirement, input schema, idempotency key rules, and expiry.

The server validates authorization and expected revision for every action. UI confirmation does not authorize an action by itself. Stale actions fail visibly with the current revision; renderers do not retry against a newer revision without user intent.

## 8. Terminal Control

Terminal mutation requires a current lease containing holder, resource, state, expiry, and fencing epoch. Input, paste, resize, signal, interrupt, and every other state-changing frame carry the same resource-level fencing epoch. Relay or Runner rejects frames from an older epoch even if an old browser connection remains open.

Observer mode may subscribe to output but cannot write, resize, signal, or claim that it controls the terminal. Lease loss is published as durable runtime state and immediately disables input.

## 9. Outer Embed Authorization

The existing one-time embed context remains the outer iframe authorization boundary: exact parent origin, parent-held redemption proof, session-bound token, nonce, expiry, and route capability enforcement.

Embed grants are intersected with session capabilities and resource authorization. Renderer configuration from the parent is data-only and can select prebundled renderer IDs; it cannot supply executable code or broaden grants.

## 10. Phase Zero Defects

Implementation starts by removing current escape paths before adding richer renderers:

1. replace top-level `text/html` Blob opening in `clients/web/src/components/media/HtmlPreviewCard.tsx`;
2. remove app-origin authority from `clients/web-user/src/components/PreviewPanel.tsx` and its new-window path;
3. configure and enforce the dedicated preview origin;
4. replace reusable preview bootstrap credentials with atomic single-use redemption;
5. block remote Markdown resources by default in shared and Web User renderers.

No compatibility branch keeps the unsafe behavior available.

## 11. Verification

Browser tests assert frame origins, sandbox tokens, CSP, Permissions Policy, referrer policy, blocked network requests, blocked popup/form/navigation attempts, bootstrap replay rejection, remote-frame message replay rejection, stale artifact action rejection, and terminal fencing.

Security completion requires network traces and browser screenshots for static HTML, interactive preview, Markdown with remote resources, embedded workbench, and terminal lease loss.

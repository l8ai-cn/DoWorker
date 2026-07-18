# Acceptance Checklist

Loop: Codex ACP Precreation Verification Loop

- [x] `accept-contract`: Codex uses `codex-app-server` with ACP and requires an `openai-compatible` model binding.
- [x] `accept-runtime`: Catalog, Docker image, and Runner report Codex CLI `0.144.5`.
- [x] `accept-precreation`: Browser template and launch plans pass; no Worker launch or Pod exists.
- [x] `accept-review`: Precreation-only evidence rejects plaintext runtime preferences as credentials.
- [ ] `accept-live-launch`: Named approval, model-resource credential use, Pod, ACP, provider smoke, and cleanup.

The unchecked live-launch item blocks terminal success and any supported claim.

# Codex ACP Loop Rules

Canonical contract: `loop.json`.

- Reload durable state before every cycle.
- Run `bash scripts/verify.sh` before checking an acceptance item.
- Treat `blocked_human_gate` as a pause, never as support.
- Do not validate a provider connection, use credentials, create a Pod, or use ACP
  before the named approval in `DECISIONS.md`.
- Do not weaken the protected verifier paths.

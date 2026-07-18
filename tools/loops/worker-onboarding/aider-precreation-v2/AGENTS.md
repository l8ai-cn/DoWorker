# AGENTS.md

This file adapts the Aider precreation verification loop for Codex.

Canonical loop contract: `loop.json`

## Execution Rules

- Re-read `loop.json`, `worker.json`, `state.json`, `PROGRESS.md`, `ACCEPTANCE.md`, and `DECISIONS.md` before each loop cycle.
- Mark `ACCEPTANCE.md` items checked only after their criteria, verifier refs, and evidence refs are satisfied.
- Treat `blocked_human_gate` as a pause, not a successful integration result.
- Use `decision_policy` when blocked: `decision-proxy` may only order read-only evidence work; `loop-supervisor` enforces the live-launch gate.
- Do not weaken `verification_policy.protected_paths` or terminal verifier commands.
- Stop and report when a requested capability is unsupported by Codex.
- Gate irreversible actions listed in `human_gates.irreversible_actions`.

## Verification

Run `bash scripts/verify.sh` for precreation evidence. Run
`bash scripts/verify-live-launch.sh` only after named human approval.

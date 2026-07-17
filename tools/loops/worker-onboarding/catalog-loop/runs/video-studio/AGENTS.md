# AGENTS.md

This file adapts `Worker Onboarding Loop Template` for Codex.

Canonical loop contract: `loop.json`

## Execution Rules

- Re-read `loop.json`, `state.json`, `PROGRESS.md`, `ACCEPTANCE.md`, `DECISIONS.md`, `tasks.json`, and `agents.json` before each loop cycle.
- Mark `ACCEPTANCE.md` items checked only after their criteria, verifier refs, and evidence refs are satisfied.
- Use `decision_policy` when blocked: `decision-proxy` may decide only inside confirmed low-risk authority; `loop-supervisor` must review goal drift on cadence.
- Use subagents only when `collaboration_policy.subagent_activation.allowed_when` applies.
- Do not weaken `verification_policy.protected_paths` or terminal verifier commands.
- Stop and report when a requested capability is unsupported by Codex.
- Gate irreversible actions listed in `human_gates.irreversible_actions`.

## Verification

Run `bash scripts/verify.sh` for terminal verification unless the manifest defines
a stricter runtime-specific verifier.

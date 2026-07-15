# Loop V1 Progress

- Goal status: complete
- Branch: `codex/loop-blockly-mvp`
- Started: 2026-07-14
- Maximum implementation iterations: 12
- Maximum wall time per iteration: 45 minutes
- No-progress rule: two consecutive iterations without a new passing test, browser assertion, or resolved review finding
- Escalation: stop on architecture/security/data ambiguity or repeated no-progress; record the exact blocker and required human decision

## Machine-Checkable Done

- [x] Go parser/compiler round-trip and rejection tests pass.
- [x] Compile and run Connect handler tests pass.
- [x] Rust Core LoopState tests and WASM build pass.
- [x] Web projection/editor tests, typecheck and lint pass.
- [x] Browser E2E proves both edit directions, invalid-code lockout and real GoalLoop creation/start.
- [x] Final review has no blocking findings.
- [x] Fix commit is pushed and visible on the remote branch.

## Iterations

| Iteration | Scope | Evidence | Status |
| --- | --- | --- | --- |
| 1 | Plan and repository mapping | Plan line checks and `git diff --check` pass | complete |
| 2 | Authoritative LoopScript core | Go tests, spec review and quality review pass | complete |
| 3 | Connect compile/run contract | Handler/service tests and two-stage review pass | complete |
| 4 | Rust Core LoopState and WASM bridge | State/service tests, Cargo checks and WASM build pass | complete |
| 5 | Blockly and CodeMirror workbench | Projection tests, typecheck, lint and responsive browser QA pass | complete |
| 6 | Real execution integration | GoalLoop `checkout-fix-2` and Pod `7-standalone-62c1f8c9` were created and started | complete |
| 7 | Final review and delivery | Independent review found no P1/P2; clean candidate tree passed all checks; commit `3b23b774290b4c7ba30ed3d1a159d5029360e556` is visible on `origin/codex/loop-blockly-mvp` | complete |
| 8 | Single-runtime deterministic controller | Race tests, real browser integration and independent review cover exact verification consumption, durable retry commands, replay filtering and recovery | complete; merge and deployment gated on migration sequence |

## Integration Evidence

- Worker snapshot `23` passed backend freshness validation and was compiled into the launch spec.
- The browser exercised blocks to code, code to blocks, invalid source lockout and recovery without console warnings or errors.
- The authenticated Run path reached GoalLoop creation, Pod creation and Runner target startup.
- A clean candidate tree, built only from the exact delivery index, passed all focused tests and the three-test Playwright suite on port `10027`.
- Start/cancel races and verification cleanup failures now persist an explicit retryable Pod cleanup state; timeout sweeping processes the full batch even when one Runner remains unavailable.
- The target `openclaw` process started, then the existing Autopilot controller attempted its fixed `claude` control command. The single-runtime runner image does not contain that binary, so the existing same-error circuit breaker paused the loop after three errors.
- GoalLoop V1 now uses a deterministic verifier-driven controller and sends failed verification evidence back to the same Worker. It no longer requires a second control-plane CLI inside the single-runtime runner image.
- Failed verification evidence and a stable retry command ID are persisted together; the pending Runner command queue and Runner command-ID deduplication make dispatch recoverable.
- Runner releases a reserved command ID when prompt delivery fails, so recovery retries are not absorbed as false duplicates.
- Runner persists accepted or uncertain prompt command IDs and completed verifier results outside the Worker sandbox; restart replay does not repeat their side effects.
- Persisted verifier requests are redispatched with the original request ID, closing the backend crash window between state claim and command delivery.
- Persisted Pod agent status rejects reordered events; an `executing` callback activates a queued retry without re-enqueuing it.
- Timeout sweeping terminates expired Loops before retry recovery, and every Pod termination path removes pending commands for that Pod.
- Duplicate command IDs are detected before per-Runner capacity checks, preserving idempotent retry semantics when a queue is full.
- GoalLoop `checkout-fix-7` and Pod `7-standalone-616af48d` exercised the repaired path in the browser without creating an Autopilot controller.
- Independent final review found no P0, P1 or P2 issue after the Runner handler and deterministic controller test files were split by responsibility.
- Fix commit `bd3136d5132bca49579f0708ce5b5eb93ed969fe` is visible on `origin/codex/loop-blockly-mvp`.
- Merge and deployment remain gated until migrations `000207` through `000212` are published into the target history; deploying `000213` and `000214` first would make the lower migration versions permanently ineligible.

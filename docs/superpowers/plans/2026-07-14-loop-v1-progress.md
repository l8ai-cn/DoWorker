# Loop V1 Progress

- Goal status: active
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
- [ ] Commit is pushed and visible on the remote branch.

## Iterations

| Iteration | Scope | Evidence | Status |
| --- | --- | --- | --- |
| 1 | Plan and repository mapping | Plan line checks and `git diff --check` pass | complete |
| 2 | Authoritative LoopScript core | Go tests, spec review and quality review pass | complete |
| 3 | Connect compile/run contract | Handler/service tests and two-stage review pass | complete |
| 4 | Rust Core LoopState and WASM bridge | State/service tests, Cargo checks and WASM build pass | complete |
| 5 | Blockly and CodeMirror workbench | Projection tests, typecheck, lint and responsive browser QA pass | complete |
| 6 | Real execution integration | GoalLoop `checkout-fix-2` and Pod `7-standalone-62c1f8c9` were created and started | complete |
| 7 | Final review and delivery | Independent review found no P1/P2; clean candidate tree passed Go, Rust/WASM, TypeScript, Vitest, prototype and E2E checks | delivery in progress |

## Integration Evidence

- Worker snapshot `23` passed backend freshness validation and was compiled into the launch spec.
- The browser exercised blocks to code, code to blocks, invalid source lockout and recovery without console warnings or errors.
- The authenticated Run path reached GoalLoop creation, Pod creation, Runner target startup and Autopilot startup.
- A clean candidate tree, built only from the exact delivery index, passed all focused tests and the three-test Playwright suite on port `10027`.
- Start/cancel races and verification cleanup failures now persist an explicit retryable Pod cleanup state; timeout sweeping processes the full batch even when one Runner remains unavailable.
- The target `openclaw` process started, then the existing Autopilot controller attempted its fixed `claude` control command. The single-runtime runner image does not contain that binary, so the existing same-error circuit breaker paused the loop after three errors.
- The runner image contract intentionally forbids silently bundling every CLI. The required follow-up is an explicit control-plane runner or a pluggable controller transport, not a compatibility fallback inside Loop.

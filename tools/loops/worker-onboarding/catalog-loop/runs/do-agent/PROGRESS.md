# Progress

Loop: Worker Onboarding Loop Template

## Current Status

- Status: blocked on the configured model endpoint, after a real Worker run.
- Active loop node: `worker-goal`
- Active atomic task: `verify-worker-flow`
- Last verifier result: Do Agent `0.2.7` completed ACP initialization, accepted
  the Prompt resource, and then failed its OpenAI-compatible request during
  TLS connection setup.
- Last no-progress fingerprint: `do-agent|openai-endpoint-unreachable|0.2.7`

## Verified

- The local Do Agent image is immutable at
  `sha256:b652fdfcb27a7cb08c9c030b5e433d513ad903d2a13019905e4e50633f43778b`
  and runs `do-agent 0.2.7`.
- A browser-created WorkerTemplate bound the `settings` configuration document
  to `do-agent-e2e-config`, then applied snapshot `15`.
- The real Worker `1-standalone-ccf70573` started on Runner `9`, entered
  `running/idle`, created an ACP session, and sent its initial Prompt resource.
- After Relay was restored for this worktree, the browser obtained terminal
  control and the Runner maintained one Relay connection.
- The Web form now blocks Plan when a Definition-required configuration document
  has no EnvironmentBundle reference. The browser showed the Chinese
  `settings` warning and disabled `生成计划`; the same check rejects direct
  Plan invocation before the API call.
- `pnpm run web:test` passed 308 test files and 2283 tests. `pnpm run
  web:typecheck` passed.

## Blocker

`dev-org` has one enabled ModelBinding, `qa-gpt-5`, configured for
`https://api.openai.com/v1`. From the Do Agent Runner, that endpoint resolves
to `198.18.1.2` and TLS fails before an HTTP response. The same Runner can
reach Alibaba DashScope and Tencent LKEAP endpoints, but no corresponding
credentialed ModelBinding exists in this organization. No provider credential
was read, changed, or replaced.

## Next Cycle

1. Create or select one non-production Alibaba DashScope or Tencent-compatible
   ModelBinding with an approved credential.
2. Apply a new Do Agent WorkerTemplate that pins that compatible binding.
3. Create a disposable Worker, send a browser prompt, and require a completed
   provider reply.
4. Only then rerun the terminal verifier and independent acceptance review.

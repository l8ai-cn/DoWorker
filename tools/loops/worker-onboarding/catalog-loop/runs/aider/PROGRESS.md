# Progress

Loop: Worker Onboarding Loop Template

## Current Status

- Status: blocked on an Aider provider credential bundle.
- Active loop node: `worker-goal`
- Active atomic task: `verify-worker-flow`
- Last verifier result: real Aider starts but waits for interactive provider login
  when no key is injected.
- Last no-progress fingerprint: `aider|missing-provider-credential|0.86.2`

## Verified

- Aider Definition version `2` declares the exact `aider-pty` PTY contract.
- `OPENAI_API_KEY` and `ANTHROPIC_API_KEY` are credential-bundle references,
  and the new `provider-api-key` group requires one of them before planning.
- Definition schema, backend projection, Connect options JSON, and the Web form
  propagate the credential group. Targeted Go tests, the configured Web test
  suite, and `pnpm run web:typecheck` passed.
- The new Definition was synchronized to the local database, then the backend
  was restarted and passed its health check on port `12415`.
- Browser creation selected Aider, Local Runner Pool, and Standard Profile.
  With both credential fields empty, the form displayed the credential warning
  and kept `生成计划` disabled; browser console had no warnings or errors.
- A real authenticated `PlanResource` request using the browser-generated YAML
  returned one blocking issue and no plan. The database confirms zero resources
  and zero plans named `aider-no-credential-live`.
- The dedicated Runner image is running Aider `0.86.2`. A no-credential CLI
  invocation timed out after 15 seconds while waiting for provider login; it
  did not produce a synthetic model reply.

## Blocker

`dev-org` has no compatible credential EnvironmentBundle for Aider. The next
real prompt requires one non-production bundle that injects either
`OPENAI_API_KEY` or `ANTHROPIC_API_KEY`. No raw secret was read, logged, or
changed.

## Next Cycle

1. Create or select an encrypted Aider credential EnvironmentBundle.
2. Bind it to one of the two Aider credential fields.
3. Generate and apply a new immutable WorkerTemplate.
4. Create the Worker, send a browser prompt, and require a completed reply.

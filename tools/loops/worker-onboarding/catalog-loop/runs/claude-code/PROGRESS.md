# Progress

Loop: Worker Onboarding Loop Template

## Current Status

- Status: blocked on a valid non-production Anthropic ModelBinding.
- Active loop node: `worker-goal`
- Active atomic task: `verify-worker-flow`
- Last verifier result: Claude stream-json initialization succeeded; the CLI reported no login or API key.
- Last no-progress fingerprint: `claude-code|missing-anthropic-model-binding|2.1.211`

## Verified

- The dedicated `claude-code` Runner is online with `claude --version` equal to
  `2.1.211 (Claude Code)`.
- The canonical contract uses executable `claude`, adapter `claude-stream-json`,
  and `-p --verbose --input-format stream-json --output-format stream-json`.
- A real stream-json invocation emitted a Claude `init` record before returning
  `Not logged in · Please run /login`; it did not simulate a successful reply.
- The only available ModelBinding resolves to `openai / gpt-5`.
- Browser plan generation for Claude rejected that binding at
  `/spec/modelResourceId` with the safe incompatible-model message. No
  template, Worker, or Pod was created.

## Blocker

The development organization has no Anthropic-protocol ModelBinding. A valid
encrypted non-production Anthropic credential must be supplied through resource
management; no raw credential was read, logged, or changed.

## Next Cycle

1. Create or select a valid Anthropic model resource and immutable ModelBinding.
2. Generate and apply the Claude WorkerTemplate using that binding.
3. Create the Worker, send a browser prompt, and require a completed reply.

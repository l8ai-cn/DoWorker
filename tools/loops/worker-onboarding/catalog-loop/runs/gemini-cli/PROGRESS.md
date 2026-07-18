# Progress

Loop: Worker Onboarding Loop Template

## Current Status

- Status: blocked on a valid non-production Gemini ModelBinding.
- Active loop node: `worker-goal`
- Active atomic task: `verify-worker-flow`
- Last verifier result: Gemini ACP initialization succeeded; session creation rejected the missing API key.
- Last no-progress fingerprint: `gemini-cli|missing-gemini-model-binding|gemini-0.50.0`

## Verified

- The dedicated `gemini-cli` Runner is online, and its image runs
  `gemini --version` as `0.50.0`.
- The canonical contract uses executable `gemini`, adapter `gemini-acp`, and
  ACP mode `gemini --experimental-acp`.
- The real CLI accepts ACP `initialize` without a credential and advertises
  Gemini API key, OAuth, Vertex AI, and gateway authentication methods.
- A real ACP `session/new` call without `GEMINI_API_KEY` returns
  `Gemini API key is missing or not configured.`
- `GEMINI_API_KEY SECRET OPTIONAL` is model-resource managed; the definition
  still requires a Gemini protocol ModelBinding and injects the key at runtime.
- The only available `ModelBinding` resolves to enabled `openai / gpt-5`.
  Browser plan generation rejects it at `/spec/modelResourceId` with the
  safe incompatibility message. No orchestration resource or Pod was created.

## Blocker

The development organization has no Gemini-protocol ModelBinding. A valid
encrypted non-production Gemini credential must be created through resource
management; no raw credential was read, logged, or changed.

## Next Cycle

1. Create or select a validated Gemini model resource and immutable
   ModelBinding through the resource-management flow.
2. Generate and apply the Gemini Worker template using that binding.
3. Create the Worker, send a browser prompt, and require a completed ACP
   session plus an agent reply.

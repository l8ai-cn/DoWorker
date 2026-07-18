# Progress

Loop: Worker Onboarding Loop Template

## Current Status

- Status: blocked on a valid named non-production OpenAI credential.
- Active loop node: `worker-goal`
- Active atomic task: `verify-worker-flow`
- Last verifier result: provider authentication rejected after a real browser prompt.
- Last no-progress fingerprint: `opencode|provider-auth-rejected|model-resource-1|definition-v4`

## Verified

- Definition v4 requires an `openai-compatible` model resource and injects
  `OPENAI_API_KEY` only through the model-resource binding.
- The browser created `opencode-openai-model-live-e2e-v2`
  (`1-standalone-a70efd0b`) from WorkerSpec snapshot `13`.
- Runner started `opencode acp`, created the sandbox `opencode.json` with
  `model: "openai/gpt-5"`, completed ACP initialization, and subscribed to
  Relay.
- ACP reports streaming unsupported for this runtime, so the AgentFile now
  declares `CAPABILITY streaming false`.
- The browser prompt reached OpenAI and returned an explicit invalid-API-key
  error. No credential value was read or recorded.

## Blocker

`QA Primary Model` is marked valid in stored metadata but its current OpenAI
credential is rejected by the provider. This is a credential lifecycle issue,
not an OpenCode launch, configuration, model-selection, or ACP adapter issue.

## Next Cycle

1. Update the named non-production provider connection through the encrypted
   resource-management flow with a valid API key.
2. Re-send `Reply with exactly: READY` to this Worker.
3. Keep OpenCode unsupported until the browser displays the agent reply and
   Runner logs confirm the completed prompt.

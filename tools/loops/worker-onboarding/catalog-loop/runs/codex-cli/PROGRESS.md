# Progress

Loop: Worker Onboarding Loop Template

## Current Status

- Status: blocked by the configured OpenAI provider location.
- Active loop node: `worker-goal`
- Active atomic task: `verify-worker-flow`
- Last verifier result: real Codex ACP prompt reached the OpenAI provider and received HTTP 403.
- Last no-progress fingerprint: `codex-cli|openai-provider-region-forbidden|0.144.5`

## Verified

- The dedicated `codex-cli` Runner image contains `codex-cli 0.144.5`.
- The canonical contract uses executable `codex`, adapter `codex-app-server`,
  ACP mode `codex app-server`, and an `OPENAI_API_KEY` model-resource binding.
- Browser plan/apply created WorkerTemplate `codex-real-worker-e2e` and
  immutable WorkerSpec snapshot `14`.
- Browser Worker creation produced Pod `1-standalone-fd17dc70` on
  `dev-runner-codex`; the Pod is running with its pinned WorkerSpec snapshot.
- Runner evidence confirms the exact `codex app-server` command, ACP
  initialization, session creation, and Relay connection.
- The ACP popup originally rendered a terminal for `app-server`. It now renders
  the ACP `AgentPanel`, obtains a control lease, and exposes the real prompt UI.
- The browser sent `Reply with exactly: READY`; the provider rejected it with
  HTTP 403 due to the Runner location, after the request left the adapter.

## Blocker

The selected OpenAI provider endpoint rejects requests from this Runner
location. The next verification requires a non-production OpenAI-compatible
provider reachable from the Runner, or a Runner deployed in a supported region.
No raw credential was read, printed, or changed.

## Next Cycle

1. Bind a reachable non-production OpenAI-compatible model resource.
2. Create a new immutable Codex WorkerTemplate and Worker using that binding.
3. Send a browser prompt and require the exact completed reply before marking
   Codex as supported.

# OpenClaw and Hermes Worker Status

OpenClaw and Hermes have formal Worker Definitions, but neither has a published
immutable runtime image lock. They are therefore unavailable in the Worker
creation flow and must not be described as built-in runnable agents.

| Area | OpenClaw | Hermes |
| --- | --- | --- |
| Worker slug | `openclaw` | `hermes` |
| Executable | `openclaw` | `hermes` |
| Adapter | `openclaw-pty` | `hermes-pty` |
| Interaction mode | PTY | PTY |
| Model resource | OpenAI-compatible | OpenAI-compatible |
| Credential injection | `OPENAI_API_KEY` | `OPENAI_API_KEY` |
| Immutable runtime image | Missing | Missing |
| Product-path evidence | Missing | Missing |

Their authoritative contracts are in:

```text
config/worker-types/openclaw/definition.json
config/worker-types/openclaw/AgentFile
config/worker-types/hermes/definition.json
config/worker-types/hermes/AgentFile
```

Before either type becomes selectable, its own evidence run must prove:

1. A real non-mock image builds and passes the Definition version probe.
2. The image is published with an immutable digest lock.
3. Runner resolves an explicit non-interactive PTY command that accepts the
   Worker prompt and does not enter setup/onboarding.
4. An authorized compatible model resource reaches preflight, creation, a real
   terminal interaction, and cleanup.

The current OpenClaw image passes `openclaw --version`, but its declared bare
`openclaw` command exits outside a TTY because it enters onboarding. Its
AgentFile is therefore not a runnable Worker launch contract. No inferred
`agent --local`, provider configuration, or credential fallback is accepted
until the exact upstream command and product path are tested.

Hermes previously failed its runtime build because the upstream Debian package
repository repeatedly returned HTTP 502 while installing `python3-pip`. That is
a blocker to resolve at the real upstream path; no substitute image is accepted.

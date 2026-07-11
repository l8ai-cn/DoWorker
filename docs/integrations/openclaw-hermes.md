# OpenClaw and Hermes Workers

OpenClaw and Hermes are available as builtin Worker agents.

## Runtime Contract

| Area | OpenClaw | Hermes |
| --- | --- | --- |
| Agent slug | `openclaw` | `hermes` |
| Executable | `openclaw` | `hermes` |
| Image | `do-worker/runner-openclaw:latest` | `do-worker/runner-hermes:latest` |
| Mode | PTY | PTY + ACP |
| Launch | `openclaw` | `hermes` / `hermes acp` |
| Home env | `OPENCLAW_HOME` | `HERMES_HOME` |

Both agents accept optional provider credentials through EnvBundle fields:

```text
OPENAI_API_KEY
ANTHROPIC_API_KEY
XAI_API_KEY
GOOGLE_API_KEY
GEMINI_API_KEY
```

Exact model resources are injected as ephemeral env bundles. OpenClaw also
receives the selected model through AgentFile `CONFIG model`, which turns into
`--model <model>`.

Hermes also exposes TUI Gateway JSON-RPC and an OpenAI-compatible HTTP API.
Those are separate transports and should be wired through dedicated runner
protocol support instead of being disguised as ACP.

## Local Validation

```bash
bash docker/agent-runtime/build.sh openclaw
bash docker/agent-runtime/build.sh hermes
cd deploy/dev
docker compose up -d runner-openclaw runner-hermes
```

For local Kubernetes runners, set `RUNNERS_LAUNCHER=k8s`; the generated runner
manifest includes `runner-openclaw` and `runner-hermes`.

## Implementation Index

```text
agents table rows are managed through DoSQL-controlled changes
runner/internal/agents/openclaw/
runner/internal/agents/hermes/
docker/agent-runtime/Dockerfile
deploy/dev/docker-compose.runners.yml
deploy/kubernetes/cluster-oilan/39-runner-openclaw.yaml
deploy/kubernetes/cluster-oilan/41-runner-hermes.yaml
```

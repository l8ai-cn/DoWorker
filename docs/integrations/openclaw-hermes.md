# OpenClaw and Harn Workers

OpenClaw and Harn are available as builtin Worker agents.

## Runtime Contract

| Area | OpenClaw | Harn |
| --- | --- | --- |
| Agent slug | `openclaw` | `harn` |
| Executable | `openclaw` | `harn` |
| Image | `do-worker/runner-openclaw:latest` | `do-worker/runner-harn:latest` |
| Mode | PTY | ACP |
| Launch | `openclaw` | `harn serve acp {sandbox}/harn-agent.harn` |
| Home env | `OPENCLAW_HOME` | `HARN_HOME` |

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

## Local Validation

```bash
bash docker/agent-runtime/build.sh openclaw
bash docker/agent-runtime/build.sh harn
cd deploy/dev
docker compose up -d runner-openclaw runner-harn
```

For local Kubernetes runners, set `RUNNERS_LAUNCHER=k8s`; the generated runner
manifest includes `runner-openclaw` and `runner-harn`.

## Implementation Index

```text
agents table rows are managed through DoSQL-controlled changes
runner/internal/agents/openclaw/
runner/internal/agents/harn/
docker/agent-runtime/Dockerfile
deploy/dev/docker-compose.runners.yml
deploy/kubernetes/cluster-oilan/39-runner-openclaw.yaml
deploy/kubernetes/cluster-oilan/41-runner-harn.yaml
```

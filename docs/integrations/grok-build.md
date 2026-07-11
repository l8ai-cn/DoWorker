# Grok Build Integration

Grok Build is available as the builtin `grok-build` agent. Its executable is
`grok`, and it supports both PTY and ACP pods.

## Runtime Contract

| Area | Value |
| --- | --- |
| Agent slug | `grok-build` |
| Executable | `grok` |
| Image | `do-worker/runner-grok-build:latest` |
| PTY mode | `grok --no-auto-update` |
| ACP mode | `grok --no-auto-update agent stdio` |
| Required secret | `XAI_API_KEY` |

The Runner sends the ACP `initialize` request, then explicitly authenticates
with the `xai.api_key` method. A missing key or an unsupported authentication
method stops the pod initialization with an error.

`GROK_HOME` is isolated per pod at `{sandbox}/grok-home`; the Runner copies
the user's `.grok` directory into that location before launch.

## Local Validation

```bash
bash docker/agent-runtime/build.sh grok-build
cd deploy/dev
docker compose up -d runner-grok-build
```

Set `XAI_API_KEY` through an agent credential or environment bundle before
creating a Grok Build pod.

## Implementation Index

```text
backend/migrations/000191_add_grok_build_agent.{up,down}.sql
runner/internal/agents/grok/
runner/internal/acp/transport_acp_handshake_hook.go
docker/agent-runtime/Dockerfile
deploy/dev/docker-compose.runners.yml
deploy/kubernetes/cluster-oilan/38-runner-grok-build.yaml
```

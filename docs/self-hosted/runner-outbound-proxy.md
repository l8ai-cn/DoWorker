# Runner Outbound Proxy

Use a Runner-only proxy when a Worker must reach an external model provider that
is not directly reachable from the Runner network.

For the development Docker Runner, set these variables in `deploy/dev/.env`
before starting or recreating the Runner services:

```bash
RUNNER_HTTP_PROXY=http://proxy.example:8080
RUNNER_HTTPS_PROXY=http://proxy.example:8443
RUNNER_NO_PROXY=traefik,host.lan,host.docker.internal,localhost,127.0.0.1,::1,postgres,redis,otel-collector
```

`RUNNER_HTTP_PROXY` and `RUNNER_HTTPS_PROXY` are passed only to Runner
containers. They do not change Traefik, Backend, Relay, or browser traffic.
`RUNNER_NO_PROXY` must retain internal service addresses so the control plane,
Relay, and telemetry stay on the local network.

The local Kubernetes Runner manifest generator uses the same variables. Its
default no-proxy set also includes `.svc` and `.cluster.local`.

Do not put provider API keys in these variables. Configure provider credentials
through the Worker model or credential resource flow, then verify the exact
Worker with a harmless prompt and termination cleanup.

Validate the wiring without contacting a provider:

```bash
bash deploy/dev/runner_runtime_contract_test.sh
```

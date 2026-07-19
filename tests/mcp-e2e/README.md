# MCP End-to-End Tests

Black-box test suite that drives the runner's MCP HTTP server (JSON-RPC 2.0)
and asserts the full chain: HTTP `/mcp` → runner gRPC client → backend gRPC
dispatch → service → Postgres. Complements (does not replace) the unit and
integration tests in `runner/internal/mcp/` and `backend/internal/api/grpc/`,
which mock one side of the link.

## Scope

- **In scope**: every MCP tool exposed by `runner/internal/mcp/registerTools()`
- **Out of scope**: real LLM behaviour (we use the `e2e-echo` stub agent),
  mTLS PKI, web/desktop UI

## Local run

```bash
# 1) Start the dev stack — Postgres, host backend/relay, runners, and seeds.
./deploy/dev/dev.sh --backend-only

# 2) Run the suites. Source deploy/dev/.env when using a worktree port offset.
source deploy/dev/.env
cd tests/mcp-e2e
BACKEND_HTTP_PORT="${BACKEND_HTTP_PORT:-10015}" \
MCP_PORT="${RUNNER_MCP_PORT:-10018}" \
RUNNER_2_MCP_PORT="${RUNNER_2_MCP_PORT:-10019}" \
POSTGRES_PORT="${POSTGRES_PORT:-10002}" \
go test ./suites/... -count=1
```

## Layout

| Path | Role |
|---|---|
| `client/mcp.go` | 50-line JSON-RPC 2.0 over HTTP client, `X-Pod-Key` header, double-decode of `result.content[0].text` |
| `client/backend_rest.go` | `/api/v1` client (login, list_runners, create_pod, terminate_pod) |
| `client/db.go` | gorm-free `database/sql` queries for fact assertions (block count, op_log presence, workspace UUID lookup) |
| `fixture/env.go` | Read deploy/dev ports + creds from environment, with sensible defaults |
| `fixture/auth.go` | Process-scoped login cache (one token per `go test` invocation) |
| `fixture/runner.go` | Discover the online dev-runner via REST list |
| `fixture/pod.go` | `NewEchoPod(t)` creates a Pod, waits for runner registration via the debug `/pods` endpoint, registers `t.Cleanup` to terminate |
| `suites/*_test.go` | One file per MCP tool family |

## stub agent: `e2e-echo`

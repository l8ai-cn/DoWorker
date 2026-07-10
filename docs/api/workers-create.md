# Create Worker — Fields & External API

In the product UI, a **Worker** is an isolated AI agent runtime (internally a **Pod**). The create form is a 3-step wizard; the External API accepts the same configuration as JSON.

## Authentication

```http
Authorization: Bearer <organization_api_key>
```

Required scope: `pods:write` (read/list also needs `pods:read`).

Base path:

```text
POST /api/v1/ext/orgs/{org_slug}/workers
```

Legacy alias (same handler): `POST /api/v1/ext/orgs/{org_slug}/pods`

---

## Field reference (UI → API)

### Step 1 — Runtime

| UI field | API field | Required | Notes |
|----------|-----------|----------|-------|
| Worker image | `agent_slug` | **Yes** | e.g. `codex-cli`, `claude-code`, `do-agent`. Discover via `GET .../runners/available` → `available_agents`. |
| Cluster / Runner | `runner_id` | No | Omit to auto-select an online runner that supports `agent_slug`. |
| Git repository | `repository_id` | No | Platform ID from `GET .../repositories`. Clones repo into the sandbox. |
| Branch | `agentfile_layer` → `BRANCH "..."` | No* | *Recommended when `repository_id` is set. Default: repo `default_branch`. |
| Interaction mode | `agentfile_layer` → `MODE pty\|acp` | No | Prefer `automation_level` for permission/autonomy. Default `pty` unless `autonomous` forces ACP. |
| Automation level | `automation_level` | No | `interactive` \| `auto_edit` \| `autonomous` (default). Mapped server-side to each agent's native permission settings; `autonomous` also forces ACP when supported. |
| Duration (short / long) | `perpetual` | No | `true` = long-lived workspace (auto-restart on exit). Default `false`. |
| Initial task | `agentfile_layer` → `PROMPT "..."` | No | First instruction sent to the agent. |
| Alias | `alias` | No | Display name, max 100 chars. |

### Step 2 — Capabilities

| UI field | API field | Required | Notes |
|----------|-----------|----------|-------|
| Knowledge bases | `knowledge_mounts` **or** `agentfile_layer` → `KNOWLEDGE slug [rw], ...` | No | Prefer top-level `knowledge_mounts` for API clarity. |
| Skills | `agentfile_layer` → `SKILLS slug1, slug2` | No | Skills are installed per repository. Requires `repository_id`. |

`knowledge_mounts` shape:

```json
[{ "slug": "my-kb", "mode": "ro" }]
```

`mode`: `ro` (default) or `rw`.

### Step 3 — Agent instructions

| UI field | API field | Required | Notes |
|----------|-----------|----------|-------|
| Generated AgentFile preview | `agentfile_layer` | No | Auto-built from steps 1–2 when using the UI. |
| Custom AgentFile | `agentfile_layer` | No | Full DSL text; overrides generated merge when supplied explicitly. |

### Advanced (collapsed in UI)

| UI field | API field | Required | Notes |
|----------|-----------|----------|-------|
| AI model resource | `model_resource_id` | Yes for model-backed workers | Exact resource selected from AI Resource Center. |
| Runtime env bundles | `agentfile_layer` → `USE_ENV_BUNDLE "name"` (multiple) | No | Merged in declaration order; later keys win. |
| Image plugin config | `agentfile_layer` → `CONFIG key = value` | No | Agent-specific non-secret options. Model credentials are not supplied here. |

### Association & resume

| UI field | API field | Required | Notes |
|----------|-----------|----------|-------|
| Linked ticket | `ticket_slug` | No | Associates Worker with a ticket. |
| Resume from Worker | `source_pod_key` | No | Resume sandbox from a terminated Worker. |
| Resume agent session | `resume_agent_session` | No | With `source_pod_key`; restores agent CLI session when supported. |
| Queue when offline | `queue_if_offline` | No | Default `false`. Set `true` to enqueue when runner is offline/busy. |
| Queue TTL (minutes) | `queue_ttl_minutes` | No | Used with `queue_if_offline`; default 30. |

### Terminal size (optional)

| Field | Required | Notes |
|-------|----------|-------|
| `cols`, `rows` | No | Terminal geometry hint; default 0 (server ignores). |

---

## AgentFile layer DSL

When not sending a pre-built `agentfile_layer`, compose one from structured fields:

```agentfile
MODE acp
USE_ENV_BUNDLE "my-openai-key"
USE_ENV_BUNDLE "dev-preferences"
PROMPT "Fix the failing unit tests in pkg/auth"
CONFIG model = "gpt-4.1"
REPO "acme/backend"
BRANCH "feature/auth-fix"
SKILLS lint-fixer, go-test-helper
KNOWLEDGE platform-docs [rw], api-spec
```

Rules:

- `repository_id` (top-level JSON) is the **platform clone source** (URLs, credentials, runner affinity).
- `REPO` / `BRANCH` in `agentfile_layer` configure the AgentFile merge; keep both in sync with `repository_id` when cloning matters.
- `agentfile_layer` is the SSOT for MODE, PROMPT, CONFIG, bundles, SKILLS, KNOWLEDGE.

---

## Minimal request

```http
POST /api/v1/ext/orgs/dev-org/workers
Authorization: Bearer amk_...
Content-Type: application/json

{
  "agent_slug": "codex-cli"
}
```

Auto-selects a runner and starts a Worker with no repo and no initial prompt.

---

## Full example (matches UI wizard)

```http
POST /api/v1/ext/orgs/dev-org/workers
Authorization: Bearer amk_...
Content-Type: application/json

{
  "agent_slug": "codex-cli",
  "runner_id": 42,
  "repository_id": 7,
  "alias": "auth-fix-worker",
  "perpetual": false,
  "automation_level": "autonomous",
  "knowledge_mounts": [
    { "slug": "platform-docs", "mode": "ro" }
  ],
  "agentfile_layer": "MODE pty\nUSE_ENV_BUNDLE \"openai-prod\"\nPROMPT \"Implement JWT refresh and add tests\"\nREPO \"acme/backend\"\nBRANCH \"feature/jwt-refresh\"\nSKILLS go-test-helper"
}
```

### Response `201 Created`

```json
{
  "pod": {
    "pod_key": "100-standalone-a1b2c3d4",
    "status": "initializing",
    "agent_slug": "codex-cli",
    "alias": "auth-fix-worker",
    "repository_id": 7,
    "runner_id": 42
  }
}
```

Optional advisory (quota near limit):

```json
{
  "pod": { "...": "..." },
  "warning": "Concurrent pod quota nearly exceeded"
}
```

---

## Discovery endpoints (External API)

Before creating a Worker, clients typically call:

| Method | Endpoint | Scope | Purpose |
|--------|----------|-------|---------|
| GET | `/runners/available` | `runners:read` | Online runners + `available_agents` |
| GET | `/repositories` | `repos:read` | Repository IDs and slugs |
| GET | `/repositories/{id}/branches` | `repos:read` | Branch names |

---

## Lifecycle (External API)

| Action | Method | Endpoint |
|--------|--------|----------|
| List | GET | `/workers` |
| Get | GET | `/workers/{pod_key}` |
| Create | POST | `/workers` |
| Send prompt | POST | `/workers/{pod_key}/prompt` |
| Terminate | POST | `/workers/{pod_key}/terminate` |

---

## Error codes (create)

| HTTP | Code | Cause |
|------|------|-------|
| 400 | `MISSING_AGENT_SLUG` | `agent_slug` empty (non-resume) |
| 400 | `VALIDATION_FAILED` | Invalid `agentfile_layer` or alias |
| 402 | `CONCURRENT_POD_QUOTA_EXCEEDED` | Plan quota |
| 403 | `SOURCE_POD_ACCESS_DENIED` | Resume source in another org |
| 404 | `SOURCE_POD_NOT_FOUND` | Invalid `source_pod_key` |
| 503 | `NO_AVAILABLE_RUNNER` | No runner supports agent (and no queue) |
| 502 | `RUNNER_DISPATCH_FAILED` | Runner offline / unreachable |

---

## Internal (JWT) clients

The web app uses **Connect-RPC** `PodService.CreatePod` with the same fields (`proto/pod/v1/pod.proto`). External integrations should prefer the REST External API above.

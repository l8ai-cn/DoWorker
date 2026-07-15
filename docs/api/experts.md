# Experts API

Experts are reusable worker configuration templates. Each expert stores agent, runner, repository, prompt, skills, knowledge mounts, and env bundles. Running an expert creates a new worker (pod) with that configuration.

## Scopes

| Scope | Access |
|-------|--------|
| `experts:read` | List and get experts |
| `experts:write` | Create, update, delete, run experts |
| `pods:read` / `pods:write` | Accepted as fallback for read/write |

## Endpoints

Base path: `/api/v1/ext/orgs/{org_slug}/experts`

### List experts

```
GET /experts?limit=50&offset=0
```

### Get expert

```
GET /experts/{slug}
```

### Create expert

```
POST /experts
```

```json
{
  "name": "Code review assistant",
  "slug": "code-review-assistant",
  "agent_slug": "codex",
  "runner_id": 1,
  "repository_id": 42,
  "branch_name": "main",
  "prompt": "Review pull requests for security issues.",
  "interaction_mode": "pty",
  "automation_level": "autonomous",
  "perpetual": false,
  "used_env_bundles": ["openai-default"],
  "skill_slugs": ["pdf-tool"],
  "knowledge_mounts": [{ "slug": "team-docs", "mode": "ro" }]
}
```

### Update expert

```
PATCH /experts/{slug}
```

Partial update — only include fields to change.

### Delete expert

```
DELETE /experts/{slug}
```

### Run expert

```
POST /experts/{slug}/run
```

```json
{
  "alias": "review-run-1",
  "prompt_override": "Focus on SQL injection this time.",
  "runner_id": 2,
  "cols": 120,
  "rows": 40
}
```

Response `201`:

```json
{
  "pod": { "pod_key": "...", "status": "initializing" }
}
```

## Publish from worker (session auth only)

```
POST /api/v1/orgs/{org_slug}/pods/{pod_key}/publish-expert
```

Copies runtime fields from the worker record. Pass skills, knowledge mounts, and env bundles in the body when they are not persisted on the pod.

```json
{
  "name": "My expert",
  "slug": "my-expert",
  "skill_slugs": ["pdf-tool"],
  "knowledge_mounts": [{ "slug": "team-docs" }]
}
```

## Expert marketplace

Marketplace releases snapshot the expert WorkerSpec and every referenced Skill
package. Installing or upgrading never resolves mutable source rows.

Session-authenticated organization endpoints:

```text
POST /api/v1/orgs/{org_slug}/experts/{expert_slug}/market-submissions
GET  /api/v1/orgs/{org_slug}/marketplace/submissions
POST /api/v1/orgs/{org_slug}/marketplace/releases/{release_id}/withdraw
GET  /api/v1/orgs/{org_slug}/experts/{expert_slug}/market-upgrade
POST /api/v1/orgs/{org_slug}/experts/{expert_slug}/market-upgrade
POST /api/v1/orgs/{org_slug}/marketplace/experts/{application_slug}/install
```

Submission metadata includes `slug`, `summary`, `description`, `category`,
`icon`, `tags`, and `outcomes`. Only organization members may submit their
organization's experts. A release remains unavailable to installers until a
system administrator approves it in the Admin Console.

Installed experts retain their source application and release IDs. Upgrades are
explicit and atomically replace the installed expert with the latest published
release snapshot.

## Operator video catalog

The backend command below idempotently creates six platform-owned video Skills
and four software-delivery Skills, creates the three video source experts,
submits their releases, and approves them:

```bash
backend bootstrap-marketplace \
  --organization <publisher-org-slug> \
  --publisher <publisher-email> \
  --reviewer <system-admin-email> \
  --model-resource-id <model-resource-id> \
  --runtime-image-id <video-studio-runtime-image-id>
```

The publisher must be an active member of the organization. The reviewer must
be an active system administrator. Existing catalog content must match the
embedded operator catalog exactly; drift stops the command with a conflict
instead of overwriting marketplace data.

Run this provisioning command before enabling platform-owned Marketplace
experts. Their immutable runtime snapshots reference these active platform
Skills and installation rejects missing or drifted dependencies.

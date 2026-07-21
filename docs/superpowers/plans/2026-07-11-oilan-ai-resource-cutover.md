# Oilan AI Resource Cutover Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Repair the Oilan database migration gap, convert the existing encrypted OpenAI Codex configuration into an organization model resource, and create a running Codex Worker through the browser.

**Architecture:** The deployed backend already contains the AI-resource service but its migration source accidentally lost `000188_add_minimax_cli_agent`. Production is already clean at version 188 and has the exact effect of its one-statement migration. Restore only the original 188 SQL pair in source; do not restore 189 because current migration 195 owns the overlapping Pod-config schema. Build and verify a backend image that embeds the restored source and the conversion CLI before applying 190–197, run that CLI inside an isolated Kubernetes Job using the server secret environment without printing credentials, then apply cutover migrations 198 and 199. Verify the Connect API and browser wizard before creating the Worker.

**Tech Stack:** PostgreSQL, golang-migrate embedded in `/app/server`, Kubernetes Jobs, Doops gateway, Go legacy migration utility, Next.js browser UI.

---

### Task 1: Record and verify the production preconditions

**Files:**

- Create: `docs/superpowers/plans/2026-07-11-oilan-ai-resource-cutover.md`
- Test: read-only PostgreSQL and Kubernetes commands through `doops`

- [x] **Step 1: Verify migration state and live service image**

Run:

```bash
doops -session worker-ai-resource-cutover exec --target gw-oilan-node --cmd \
  "kubectl exec -n agentcloud deploy/postgres -- \
  psql -U agentcloud -d agentcloud -Atc 'SELECT version, dirty FROM schema_migrations;'"
doops -session worker-ai-resource-cutover exec --target gw-oilan-node --cmd \
  "kubectl get deploy/backend -n agentcloud -o jsonpath='{.spec.template.spec.containers[0].image}'"
```

Verified: schema version is `188|f` (`dirty=false`); the backend image is `repo.aiedulab.cn:8443/agentcloud/backend@sha256:811208b0d91f2e1eb97ec0606e1447b03dff94adcddd6f68a345b7b36f8ab611`.

Incident note: an initial recovery Job invoked `migrate force 188` after `f` was misread as dirty. It did not run migration SQL or alter application rows, and the resulting state remains `188|f`. No further force command is permitted in this cutover.

- [x] **Step 2: Verify staged migration safety**

Run:

```sql
SELECT count(*) FROM agents WHERE slug = 'grok-build';
SELECT count(*) FROM worker_spec_snapshots;
SELECT count(*) FROM ai_models;
SELECT count(*) FROM env_bundles WHERE kind = 'credential' AND is_active = TRUE;
SELECT count(*) FROM virtual_api_keys;
```

Verified: one legacy `OpenAI (Codex)` AI model, no active credential EnvBundles, and two virtual API keys. This query omitted credentials.

- [x] **Step 3: Record the browser failure**

Open `/dev-org/settings?scope=organization&tab=ai-resources` in the local Web app connected to `https://dowork.l8ai.cn`.

Verified: the browser reports failures from `ListOrganizationConnections` and `ListOrganizationEffectiveResources` while the server returns HTTP 500. This proves the repair targets the missing schema rather than a local proxy failure.

### Task 2: Restore the 188 migration source and apply schema migrations without entering the cutover

**Files:**

- Modify: `backend/migrations/000188_add_minimax_cli_agent.{up,down}.sql`
- Create: `backend/migrations/minimax_cli_agent_test.go`
- Create: Kubernetes Job manifest in the Doops session workspace only
- Test: static migration contract, `schema_migrations`, table existence, and job completion

- [x] **Step 1: Prove that production matches the historical 188 migration**

The recovered source is historical commit `2e3793c5b06ee95f64938e9b586c24162a3eb057`, with up blob `1e3964a7bfc4c444e0843ae741bf6ecc2af22af9` and down blob `fb8fad1c504100154267fc324df8ef89136001fe`. The up body contains only one `INSERT`. Production has exactly one `minimax-cli` row and every inserted field, including the AgentFile content, matches that historical insert.

- [x] **Step 2: Restore and test only the required migration source**

The source test must prove both:

```text
ReadDown(188) succeeds
Next(188) == 190
```

Do not restore `000189_pod_config_lifecycle`: it conflicts with current `000195_pod_config_revisions`.

The preliminary Job `ai-resource-schema-migrate-190-197` used the old image and failed before running SQL with `no migration found for version 188: read down for version 188`. Its failure confirms that a clean database version still requires both directions of that source migration.

- [ ] **Step 3: Build and publish a backend image containing restored 188 and the conversion CLI**

The image must be built from `git archive ac9f32fb54a91fa6c7cff49d3cf045511cac4a9f` extracted into a temporary directory, never from this shared working tree. It must contain `/app/server` and `/app/migrate-ai-resources`. Before the Docker build, run the embedded-source contract tests in that extracted archive and verify the sequence `188 → 190 → 194 → 195 → 196 → 197 → 198`; record the pushed image digest before it is used for production schema changes.

- [ ] **Step 4: Create a one-shot schema Job using the restored-image digest**

Create a Job named `ai-resource-schema-migrate-190-197-r2` with:

```yaml
command: ["/app/server", "migrate", "up", "5"]
envFrom:
  - configMapRef:
      name: agentcloud-config
  - secretRef:
      name: agentcloud-secrets
```

Use the exact digest from Step 3, not the previously deployed image that omits 188. The five required migrations are `190`, `194`, `195`, `196`, and `197`; golang-migrate counts applied migration files, not numeric gaps. `000191_add_grok_build_agent` belongs to a separate uncommitted stream and must not enter this recovery commit or image.

- [ ] **Step 5: Wait and inspect the Job**

Run:

```bash
kubectl -n agentcloud wait --for=condition=complete job/ai-resource-schema-migrate-190-197-r2 --timeout=300s
kubectl -n agentcloud logs job/ai-resource-schema-migrate-190-197-r2
```

Expected: Job completes and logs `migrate ok`.

- [ ] **Step 6: Verify schema state**

Run:

```sql
SELECT version, dirty FROM schema_migrations;
SELECT to_regclass('public.provider_connections');
SELECT to_regclass('public.model_resources');
SELECT to_regclass('public.worker_spec_snapshots');
```

Expected: version `197`, `dirty=false`, and all three tables exist.

### Task 3: Convert the legacy Codex configuration and complete the cutover

**Files:**

- Create: Kubernetes Job manifest in the Doops session workspace only
- Test: legacy conversion report, migration-map integrity, final schema migration

- [ ] **Step 1: Run the committed conversion utility in an isolated Job**

Use the exact image from Task 2 Step 3. The Job must:

```sh
/app/migrate-ai-resources --apply --created-by=2
```

Pass `DB_*` and `JWT_SECRET` only from `agentcloud-config` and `agentcloud-secrets`. The CLI reuses the backend DB configuration when `DATABASE_URL` is absent. The legacy model belongs to organization `1`, whose owner is user `2`; use that verified user as `created_by`. Do not print the connection string, key, or decrypted credentials.

- [ ] **Step 2: Verify conversion output and model-resource mapping**

Run:

```sql
SELECT source_kind, source_id, provider_connection_id, model_resource_id, status
FROM ai_resource_migration_map
ORDER BY source_kind, source_id;

SELECT connection.owner_scope, connection.owner_id, connection.provider_key,
       resource.model_id, resource.is_enabled, resource.status
FROM provider_connections connection
JOIN model_resources resource ON resource.provider_connection_id = connection.id;
```

Expected: the single legacy AI model maps to one enabled `openai` resource for `gpt-5.5`, with encrypted credentials preserved server-side and never returned by SQL output.

- [ ] **Step 3: Apply migrations 198 and 199 with the restored-image digest**

Create a one-shot Job named `ai-resource-finalize-migrate` using the exact digest from Task 2 Step 3:

```yaml
command: ["/app/server", "migrate", "up", "2"]
```

Wait for completion, then verify:

```sql
SELECT version, dirty FROM schema_migrations;
SELECT model_resource_id IS NOT NULL FROM virtual_api_keys;
```

Expected: version `199`, `dirty=false`, and every virtual key has a model-resource reference.

- [ ] **Step 4: Update the backend Deployment to the restored-image digest**

Before browser verification, update the backend Deployment to the same exact digest and wait for its rollout to complete. This guarantees the live service and future embedded migrations use the restored source.

### Task 4: Browser verification and Codex Worker creation

**Files:**

- Modify: local ignored `clients/web/.env.local`
- Test: real browser UI on `http://127.0.0.1:10007`

- [ ] **Step 1: Verify the local Web proxy reaches the server**

Run:

```bash
curl -fsS http://127.0.0.1:10007/health
```

Expected:

```json
{"service":"agent-cloud-api","status":"ok"}
```

- [ ] **Step 2: Verify the organization AI resource page in Chrome**

Navigate to:

```text
http://127.0.0.1:10007/dev-org/settings?scope=organization&tab=ai-resources
```

Expected: the page lists the migrated OpenAI Codex resource and no AI-resource Connect request returns 500.

- [ ] **Step 3: Create the Worker through the four-step browser wizard**

Use:

```text
http://127.0.0.1:10007/dev-org/workspace
```

Select the exact migrated `gpt-5.5` model resource, Codex CLI Worker type, Codex runner pool, an immutable runtime image, pooled deployment, and the Standard resource profile. Complete the Worker-specific configuration and submit only after the preflight step is successful.

- [ ] **Step 4: Verify the running Worker**

Expected: the workspace shows the created Worker as running on a Codex runner; browser console has no new errors; the server returns no AI-resource 500 responses.

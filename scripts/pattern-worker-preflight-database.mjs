import { execFileSync } from "node:child_process";
import { summarizeWorkerSpec } from "./pattern-worker-preflight-artifact.mjs";
import {
  openAICompatibleProviders,
  requiredPatternSkills,
} from "./pattern-worker-preflight-config.mjs";

export async function checkDatabaseDependencies(config, result) {
  const snapshot = readDatabaseSnapshot(config);
  const workerSpecSnapshot = summarizeWorkerSpec(
    snapshot.dependencyArtifacts,
    snapshot.organizationId,
    snapshot.actorUserId,
    snapshot.providerModel,
    snapshot.lovartCredential,
    config.orgSlug,
  );
  result.database = {
    organizationId: snapshot.organizationId ?? null,
    actorUserResolved: Boolean(snapshot.actorUserId),
    skillPackages: snapshot.skills,
    lovartCredential: snapshot.lovartCredential,
    providerModel: snapshot.providerModel,
    runner: snapshot.runner,
    workerSpecSnapshot,
  };

  if (!snapshot.organizationId) result.failures.push(`organization ${config.orgSlug} is missing`);
  if (!snapshot.actorUserId) result.failures.push("preflight user is not a dev-org member");
  for (const slug of requiredPatternSkills) {
    const skill = snapshot.skills.find((entry) => entry.slug === slug);
    if (!skill?.package_ready) result.failures.push(`skill package ${slug} is missing or incomplete`);
  }
  if (!snapshot.lovartCredential) result.failures.push("Lovart credential bundle is missing for this user/org");
  if (!snapshot.providerModel) result.failures.push("Pattern openai-compatible chat model resource is missing");
  if (!snapshot.runner?.online) result.failures.push("dedicated runner is not online with an active tunnel");
  if (!workerSpecSnapshot.ready) {
    result.failures.push(workerSpecSnapshot.reason ??
      "immutable Pattern WorkerSpec dependency artifact is missing required bindings");
  }
}

function readDatabaseSnapshot(config) {
  const sql = `
with org as (select id from organizations where slug = ${sqlString(config.orgSlug)} limit 1),
actor as (
  select u.id from users u
  join organization_members om on om.user_id = u.id
  join org on org.id = om.organization_id
  where u.email = ${sqlString(config.username ?? "")}
     or u.username = ${sqlString(config.username ?? "")}
  limit 1
)
select jsonb_build_object(
  'organizationId', (select id from org),
  'actorUserId', (select id from actor),
  'skills', coalesce((select jsonb_agg(jsonb_build_object(
    'slug', s.slug, 'agent_filter', s.agent_filter, 'content_sha', s.content_sha,
    'package_ready', s.storage_key <> '' and s.package_size > 0
      and s.agent_filter ? 'pattern-designer') order by s.slug)
    from skills s join org on s.organization_id = org.id
    where s.slug in (${sqlStringList(requiredPatternSkills)})), '[]'::jsonb),
  'lovartCredential', (select jsonb_build_object(
    'id', e.id, 'owner_scope', e.owner_scope, 'owner_id', e.owner_id,
    'name', e.name, 'agent_slug', e.agent_slug,
    'configured_fields', (
      select coalesce(jsonb_agg(key order by key), '[]'::jsonb)
      from jsonb_object_keys(e.data) as key))
    from env_bundles e
    where e.name = 'lovart' and e.agent_slug = 'pattern-designer'
      and e.is_active and e.data ? 'LOVART_ACCESS_KEY' and e.data ? 'LOVART_SECRET_KEY'
      and ((e.owner_scope = 'org' and e.owner_id = (select id from org))
        or (e.owner_scope = 'user' and e.owner_id = (select id from actor)))
    order by case when e.owner_scope = 'user' then 0 else 1 end limit 1),
  'providerModel', (select jsonb_build_object(
    'connection_id', pc.id, 'resource_id', mr.id, 'provider_key', pc.provider_key,
    'protocol_adapter', 'openai-compatible',
    'connection_revision', pc.revision, 'resource_revision', mr.revision,
    'model_id', mr.model_id, 'modalities', mr.modalities, 'capabilities', mr.capabilities)
    from provider_connections pc
    join model_resources mr on mr.provider_connection_id = pc.id
    join org on pc.owner_scope = 'org' and pc.owner_id = org.id
    where pc.is_enabled and mr.is_enabled and pc.status = 'valid' and mr.status = 'valid'
      and pc.provider_key in (${sqlStringList(openAICompatibleProviders)})
      and mr.modalities ? 'chat' and mr.capabilities ? 'text-generation'
    order by pc.id, mr.id limit 1),
  'runner', (select jsonb_build_object(
    'node_id', r.node_id, 'status', r.status, 'is_enabled', r.is_enabled,
    'available_agents', r.available_agents, 'tunnel_state', r.tunnel_state,
    'online', r.is_enabled and r.status = 'online' and r.tunnel_state = 'connected'
      and r.available_agents ? 'pattern-designer')
    from runners r join org on r.organization_id = org.id
    where r.node_id = ${sqlString(config.runnerNodeId ?? "")} limit 1),
  'dependencyArtifacts', coalesce((select jsonb_agg(jsonb_build_object(
    'snapshot_id', a.worker_spec_snapshot_id,
    'artifact_digest', a.artifact_digest,
    'artifact_json', a.artifact_json) order by a.worker_spec_snapshot_id desc)
    from worker_spec_dependency_artifacts a
    join worker_spec_snapshots s on s.id = a.worker_spec_snapshot_id
      and s.organization_id = a.organization_id
    join org on org.id = a.organization_id), '[]'::jsonb)
)::text;`;
  return JSON.parse(psql(config, sql));
}

function psql(config, sql) {
  return execFileSync("docker", ["exec", "-i", config.pgContainer, "psql",
    "-U", config.pgUser, "-d", config.pgDatabase, "-At", "-c", sql], {
    encoding: "utf8",
  }).trim();
}

function sqlStringList(values) { return values.map(sqlString).join(","); }

function sqlString(value) { return `'${String(value).replaceAll("'", "''")}'`; }

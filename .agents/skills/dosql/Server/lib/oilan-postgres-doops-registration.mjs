import { readFile } from "node:fs/promises";

const RESOURCE_URL = new URL("../../config/resources.json", import.meta.url);
const INVENTORY_URL = new URL("../../config/assets.json", import.meta.url);

export const OILAN_POSTGRES = Object.freeze({
  databaseAssetId: "db_agentsmesh_prod_postgres",
  projectId: "agentsmesh",
  environmentId: "prod",
  engine: "postgresql",
  namespace: "agentsmesh",
  serviceName: "postgres",
  doopsTarget: "gw-oilan-node",
  secretRef: "secret://agentsmesh/agentsmesh-secrets#DB_PASSWORD",
});

const QUERIES = Object.freeze({
  "asset-probe": [
    "select current_database(),",
    "current_setting('server_version_num'),",
    "exists (",
    "select 1 from information_schema.tables",
    "where table_schema = 'public' and table_name = 'schema_migrations'",
    ");",
  ].join(" "),
  "migration-version": [
    "select version, dirty",
    "from public.schema_migrations",
    "limit 1;",
  ].join(" "),
});

export async function loadOilanPostgresRegistration() {
  const [resourcesText, inventoryText] = await Promise.all([
    readFile(RESOURCE_URL, "utf8"),
    readFile(INVENTORY_URL, "utf8"),
  ]);
  return validateOilanPostgresRegistration(
    JSON.parse(resourcesText),
    JSON.parse(inventoryText),
  );
}

export function validateOilanPostgresRegistration(resources, inventory) {
  const resource = (resources?.resources ?? []).find(
    (item) => item.id === OILAN_POSTGRES.databaseAssetId,
  );
  const asset = (inventory?.assets ?? []).find(
    (item) => item.databaseAssetId === OILAN_POSTGRES.databaseAssetId,
  );
  requireEqual(resource?.kind, "PostgreSQL", "resource.kind");
  requireEqual(resource?.environmentId, OILAN_POSTGRES.environmentId, "resource.environmentId");
  requireEqual(resource?.doopsTarget, OILAN_POSTGRES.doopsTarget, "resource.doopsTarget");
  requireEqual(resource?.secretRef, OILAN_POSTGRES.secretRef, "resource.secretRef");
  requireEqual(resource?.connectionRef, "k8s://doops-oilan/oilan-node/agentsmesh/service/postgres:5432#db=agentsmesh", "resource.connectionRef");
  requireEqual(asset?.projectId, OILAN_POSTGRES.projectId, "asset.projectId");
  requireEqual(asset?.environmentId, OILAN_POSTGRES.environmentId, "asset.environmentId");
  requireEqual(asset?.engine, OILAN_POSTGRES.engine, "asset.engine");
  requireEqual(asset?.namespace, OILAN_POSTGRES.namespace, "asset.namespace");
  requireEqual(asset?.serviceName, OILAN_POSTGRES.serviceName, "asset.serviceName");
  requireEqual(asset?.source?.targetName, OILAN_POSTGRES.doopsTarget, "asset.source.targetName");
  return assertOilanPostgresRegistration({
    ...OILAN_POSTGRES,
    registrationStatus: requiredText(resource.status, "resource.status"),
    versionText: requiredText(asset.versionText, "asset.versionText"),
  });
}

export function assertOilanPostgresRegistration(registration) {
  for (const [key, value] of Object.entries(OILAN_POSTGRES)) {
    requireEqual(registration?.[key], value, `registration.${key}`);
  }
  requiredText(registration?.registrationStatus, "registration.registrationStatus");
  requiredText(registration?.versionText, "registration.versionText");
  return registration;
}

export function resolveOilanPostgresQuery(queryName) {
  const name = requiredText(queryName, "queryName");
  const sql = QUERIES[name];
  if (!sql) {
    throw new Error(`unsupported Oilan PostgreSQL read-only query: ${name}`);
  }
  return { name, sql };
}

export function buildOilanPostgresRemoteCommand(query) {
  const encodedQuery = Buffer.from(query.sql, "utf8").toString("base64");
  return [
    "set -euo pipefail",
    "namespace=agentsmesh",
    "service=postgres",
    "secret=agentsmesh-secrets",
    'kubectl -n "$namespace" get service "$service" >/dev/null',
    'kubectl -n "$namespace" get secret "$secret" >/dev/null',
    `pod="$(kubectl -n "$namespace" get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}')"`,
    'test -n "$pod"',
    `printf '%s' '${encodedQuery}' | base64 -d | kubectl -n "$namespace" exec -i "$pod" -- sh -ceu '`,
    'test -n "${POSTGRES_USER:-}"',
    'test -n "${POSTGRES_DB:-}"',
    'test -n "${POSTGRES_PASSWORD:-}"',
    'PGPASSWORD="$POSTGRES_PASSWORD" psql --no-psqlrc --set ON_ERROR_STOP=1 --no-align --tuples-only --username "$POSTGRES_USER" --dbname "$POSTGRES_DB"',
    "'",
  ].join("\n");
}

function requireEqual(actual, expected, fieldName) {
  if (actual !== expected) {
    throw new Error(`${fieldName} must equal the fixed Oilan PostgreSQL registration`);
  }
}

function requiredText(value, fieldName) {
  const text = value === undefined || value === null ? "" : String(value).trim();
  if (!text) throw new Error(`${fieldName} is required`);
  return text;
}

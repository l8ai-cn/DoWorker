import { readFile } from "node:fs/promises";

const RESOURCE_URL = new URL("../../config/resources.json", import.meta.url);
const INVENTORY_URL = new URL("../../config/assets.json", import.meta.url);

export const OILAN_POSTGRES = Object.freeze({
  databaseAssetId: "db_agentcloud_prod_postgres",
  projectId: "agentcloud",
  environmentId: "prod",
  engine: "postgresql",
  namespace: "agentcloud",
  serviceName: "postgres",
  databaseName: "agentcloud",
  doopsTarget: "gw-oilan-node",
  secretRef: "secret://agentcloud/agentcloud-secrets#DB_PASSWORD",
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
  requireEqual(resource?.connectionRef, "k8s://doops-oilan/oilan-node/agentcloud/service/postgres:5432#db=agentcloud", "resource.connectionRef");
  requireEqual(asset?.projectId, OILAN_POSTGRES.projectId, "asset.projectId");
  requireEqual(asset?.environmentId, OILAN_POSTGRES.environmentId, "asset.environmentId");
  requireEqual(asset?.engine, OILAN_POSTGRES.engine, "asset.engine");
  requireEqual(asset?.namespace, OILAN_POSTGRES.namespace, "asset.namespace");
  requireEqual(asset?.serviceName, OILAN_POSTGRES.serviceName, "asset.serviceName");
  requireEqual(asset?.databaseName, OILAN_POSTGRES.databaseName, "asset.databaseName");
  requireEqual(asset?.source?.targetName, OILAN_POSTGRES.doopsTarget, "asset.source.targetName");
  requireEqual(resource?.status, "verified-read-only", "resource.status");
  requireEqual(asset?.registration?.status, "verified-read-only", "asset.registration.status");
  requireEqual(resource?.serverVersionNum, asset?.serverVersionNum, "resource.serverVersionNum");
  requireEqual(resource?.registration, asset?.registration, "resource.registration");
  requireEqual(resource?.migrationState, asset?.migrationState, "resource.migrationState");
  return assertOilanPostgresRegistration({
    ...OILAN_POSTGRES,
    registrationStatus: resource.status,
    versionText: requiredText(asset.versionText, "asset.versionText"),
    serverVersionNum: requiredServerVersion(asset.serverVersionNum),
    registration: validateVerification(asset.registration, "registration"),
    migrationState: validateMigrationState(asset.migrationState),
  });
}

export function assertOilanPostgresRegistration(registration) {
  for (const [key, value] of Object.entries(OILAN_POSTGRES)) {
    requireEqual(registration?.[key], value, `registration.${key}`);
  }
  requireEqual(registration?.registrationStatus, "verified-read-only", "registration.registrationStatus");
  requiredText(registration?.versionText, "registration.versionText");
  requiredServerVersion(registration?.serverVersionNum);
  validateVerification(registration?.registration, "registration");
  validateMigrationState(registration?.migrationState);
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
    "namespace=agentcloud",
    "service=postgres",
    "secret=agentcloud-secrets",
    'kubectl -n "$namespace" get service "$service" >/dev/null',
    'kubectl -n "$namespace" get secret "$secret" >/dev/null',
    `pod="$(kubectl -n "$namespace" get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}')"`,
    'test -n "$pod"',
    `printf '%s' '${encodedQuery}' | base64 -d | kubectl -n "$namespace" exec -i "$pod" -- sh -ceu '`,
    'test -n "${POSTGRES_USER:-}"',
    'test -n "${POSTGRES_DB:-}"',
    'test -n "${POSTGRES_PASSWORD:-}"',
    'PGOPTIONS="-c default_transaction_read_only=on -c statement_timeout=15000" PGPASSWORD="$POSTGRES_PASSWORD" psql --no-psqlrc --set ON_ERROR_STOP=1 --no-align --tuples-only --field-separator "|" --username "$POSTGRES_USER" --dbname "$POSTGRES_DB"',
    "'",
  ].join("\n");
}

function requireEqual(actual, expected, fieldName) {
  if (stableJson(actual) !== stableJson(expected)) {
    throw new Error(`${fieldName} must equal the fixed Oilan PostgreSQL registration`);
  }
}

function validateMigrationState(value) {
  if (!Number.isSafeInteger(value?.version) || value.version < 0) {
    throw new Error("migrationState.version must be a non-negative integer");
  }
  if (typeof value.dirty !== "boolean") {
    throw new Error("migrationState.dirty must be boolean");
  }
  validateVerification(value, "migrationState");
  return value;
}

function validateVerification(value, fieldName) {
  requiredText(value?.verifiedAt, `${fieldName}.verifiedAt`);
  const operationId = requiredSlug(value?.operationId, `${fieldName}.operationId`);
  requiredSlug(value?.session, `${fieldName}.session`);
  requireEqual(
    value?.evidenceRef,
    `.dosql/readonly-evidence/${operationId}.json`,
    `${fieldName}.evidenceRef`,
  );
  if (!/^sha256:[a-f0-9]{64}$/.test(requiredText(value?.evidenceFingerprint, `${fieldName}.evidenceFingerprint`))) {
    throw new Error(`${fieldName}.evidenceFingerprint must be a sha256 digest`);
  }
  validateGatewayAudit(value?.gatewayAudit, fieldName);
  return value;
}

function validateGatewayAudit(value, fieldName) {
  if (!Number.isSafeInteger(value?.eventId) || value.eventId < 1) {
    throw new Error(`${fieldName}.gatewayAudit.eventId must be a positive integer`);
  }
  requireEqual(value?.cluster, "doops-oilan", `${fieldName}.gatewayAudit.cluster`);
  requireEqual(value?.instance, "oilan-node", `${fieldName}.gatewayAudit.instance`);
  requiredText(value?.startedAt, `${fieldName}.gatewayAudit.startedAt`);
  requiredText(value?.endedAt, `${fieldName}.gatewayAudit.endedAt`);
}

function requiredServerVersion(value) {
  const text = requiredText(value, "serverVersionNum");
  if (!/^[1-9][0-9]{4,5}$/.test(text)) {
    throw new Error("serverVersionNum must be a PostgreSQL server version number");
  }
  return text;
}

function requiredSlug(value, fieldName) {
  const text = requiredText(value, fieldName);
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(text) || text.length < 2 || text.length > 100) {
    throw new Error(`${fieldName} must be a safe identifier`);
  }
  return text;
}

function stableJson(value) {
  if (Array.isArray(value)) return `[${value.map(stableJson).join(",")}]`;
  if (value && typeof value === "object") {
    return `{${Object.keys(value).sort().map((key) => `${JSON.stringify(key)}:${stableJson(value[key])}`).join(",")}}`;
  }
  return JSON.stringify(value);
}

function requiredText(value, fieldName) {
  const text = value === undefined || value === null ? "" : String(value).trim();
  if (!text) throw new Error(`${fieldName} is required`);
  return text;
}

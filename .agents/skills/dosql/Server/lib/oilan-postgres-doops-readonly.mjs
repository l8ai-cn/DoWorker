import { spawnSync } from "node:child_process";
import { homedir } from "node:os";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

import {
  appendOilanPostgresAuditEvent,
  prepareOilanPostgresAudit,
  sha256,
  stableJson,
  verifyOilanPostgresReadOnlyEvidence,
  writeOilanPostgresEvidence,
} from "./oilan-postgres-doops-evidence.mjs";
import {
  assertOilanPostgresRegistration,
  buildOilanPostgresRemoteCommand,
  loadOilanPostgresRegistration,
  resolveOilanPostgresQuery,
} from "./oilan-postgres-doops-registration.mjs";
import { parseOilanPostgresDoopsResult } from "./oilan-postgres-query-result.mjs";

const REPO_ROOT = resolve(fileURLToPath(new URL("../../../../..", import.meta.url)));
const DOOPS_BIN = resolve(homedir(), ".local/bin/doops");
const DOOPS_CONFIG = resolve(homedir(), ".agent/skills/doops/config.json");
const ALLOWED_INPUT_FIELDS = new Set(["operationId", "session", "queryName"]);
export async function executeOilanPostgresReadOnly(input, dependencies = {}) {
  rejectUnexpectedInput(input);
  const operationId = requiredIdentifier(input.operationId, "operationId");
  const session = requiredSession(input.session);
  const query = resolveOilanPostgresQuery(input.queryName);
  const registration = assertOilanPostgresRegistration(
    dependencies.registration ?? await loadOilanPostgresRegistration(),
  );
  const auditRoot = dependencies.auditRoot ?? resolve(REPO_ROOT, ".dosql");
  const now = dependencies.now ?? (() => new Date());
  const execute = dependencies.execute ?? executeDoops;
  const paths = await prepareOilanPostgresAudit(auditRoot, operationId);
  const base = baseEvent({ operationId, session, query, registration });

  await appendOilanPostgresAuditEvent(paths.journalPath, {
    ...base,
    eventType: "dosql.readonly-query.planned",
    status: "planned",
    createdAt: now().toISOString(),
  });
  await appendOilanPostgresAuditEvent(paths.journalPath, {
    ...base,
    eventType: "dosql.readonly-query.running",
    status: "running",
    createdAt: now().toISOString(),
  });

  try {
    const result = await execute({
      command: DOOPS_BIN,
      args: [
        "-session",
        session,
        "exec",
        "--target",
        registration.doopsTarget,
        "--cmd",
        buildOilanPostgresRemoteCommand(query),
      ],
    });
    if (result?.error || result?.status !== 0) {
      throw new Error(`DoOps exited with status ${String(result?.status ?? "unavailable")}`);
    }
    const parsed = parseOilanPostgresDoopsResult(query.name, result.stdout);
    const evidence = createEvidence({
      base,
      query,
      registration,
      parsed,
      capturedAt: now().toISOString(),
      journalRef: paths.journalRef,
    });
    await writeOilanPostgresEvidence(paths.evidencePath, evidence);
    await appendOilanPostgresAuditEvent(paths.journalPath, {
      ...base,
      eventType: "dosql.readonly-query.succeeded",
      status: "succeeded",
      evidenceRef: paths.evidenceRef,
      evidenceFingerprint: evidence.evidenceFingerprint,
      createdAt: now().toISOString(),
    });
    await appendOilanPostgresAuditEvent(paths.journalPath, {
      ...base,
      eventType: "dosql.readonly-query.verified",
      status: "verified",
      evidenceRef: paths.evidenceRef,
      evidenceFingerprint: evidence.evidenceFingerprint,
      createdAt: now().toISOString(),
    });
    return evidence;
  } catch (error) {
    const evidence = createFailedEvidence({
      base,
      query,
      registration,
      capturedAt: now().toISOString(),
      journalRef: paths.journalRef,
    });
    await writeOilanPostgresEvidence(paths.evidencePath, evidence);
    await appendOilanPostgresAuditEvent(paths.journalPath, {
      ...base,
      eventType: "dosql.readonly-query.failed",
      status: "failed",
      evidenceRef: paths.evidenceRef,
      evidenceFingerprint: evidence.evidenceFingerprint,
      errorFingerprint: sha256(error instanceof Error ? error.message : String(error)),
      createdAt: now().toISOString(),
    });
    throw new Error("Oilan PostgreSQL read-only query failed");
  }
}

export { verifyOilanPostgresReadOnlyEvidence };

function rejectUnexpectedInput(input) {
  if (!input || typeof input !== "object" || Array.isArray(input)) {
    throw new Error("read-only query input must be an object");
  }
  for (const key of Object.keys(input)) {
    if (!ALLOWED_INPUT_FIELDS.has(key)) {
      throw new Error(`unsupported read-only query input: ${key}`);
    }
  }
}

function baseEvent({ operationId, session, query, registration }) {
  return {
    operationId,
    projectId: registration.projectId,
    environmentId: registration.environmentId,
    databaseAssetId: registration.databaseAssetId,
    engine: registration.engine,
    doopsTarget: registration.doopsTarget,
    session,
    queryName: query.name,
    queryFingerprint: sha256(query.sql),
  };
}

function createEvidence({ base, registration, parsed, capturedAt, journalRef }) {
  const payload = {
    schema: "dosql.oilan-postgres-readonly-evidence.v1",
    status: "verified",
    capturedAt,
    journalRef,
    ...base,
    registrationStatus: registration.registrationStatus,
    doopsRoute: parsed.doopsRoute,
    result: parsed.result,
    resultFingerprint: sha256(stableJson(parsed.result)),
  };
  return { ...payload, evidenceFingerprint: sha256(stableJson(payload)) };
}

function createFailedEvidence({ base, registration, capturedAt, journalRef }) {
  const payload = {
    schema: "dosql.oilan-postgres-readonly-evidence.v1",
    status: "failed",
    capturedAt,
    journalRef,
    ...base,
    registrationStatus: registration.registrationStatus,
  };
  return { ...payload, evidenceFingerprint: sha256(stableJson(payload)) };
}

function executeDoops(request) {
  return spawnSync(request.command, request.args, {
    encoding: "utf8",
    env: { ...process.env, DOOPS_CONFIG },
  });
}

function requiredIdentifier(value, fieldName) {
  const text = requiredText(value, fieldName);
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(text) || text.length < 2 || text.length > 100) {
    throw new Error(`${fieldName} must be a safe identifier`);
  }
  return text;
}

function requiredSession(value) {
  const text = requiredText(value, "session");
  if (!/^[a-z][a-z0-9-]{7,100}$/.test(text)) {
    throw new Error("session must be a unique DoOps session identifier");
  }
  return text;
}

function requiredText(value, fieldName) {
  const text = value === undefined || value === null ? "" : String(value).trim();
  if (!text) throw new Error(`${fieldName} is required`);
  return text;
}

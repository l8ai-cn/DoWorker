#!/usr/bin/env node
import { createHash } from "node:crypto";
import { existsSync, readFileSync, realpathSync } from "node:fs";
import { dirname, relative, resolve, sep } from "node:path";

const args = parseArgs(process.argv.slice(2));
const expected = {
  target: required(args.target, "target"),
  mode: required(args.mode, "mode"),
  session: required(args.session, "session"),
  changeId: required(args.changeId, "change-id"),
  operationId: required(args.operationId, "operation-id"),
  version: Number(required(args.expectedVersion, "expected-version")),
};
requireEqual(expected.target, "db_agentsmesh_prod_postgres", "target");

const sources = resolveAuditSources(args);
const journal = readJournal(sources.journal);
const evidenceText = readFileSync(sources.evidence, "utf8");
const evidence = JSON.parse(evidenceText);
const evidenceFingerprint = `sha256:${sha256(evidenceText)}`;
verifyEvidence(evidence, evidenceFingerprint, expected);
verifyLifecycle(journal, evidenceFingerprint, sources, expected);
console.log(`==> verified DoSql journal evidence operation=${expected.operationId} version=${expected.version}`);

function resolveAuditSources(value) {
  if (value.testJournal || value.testEvidence) {
    if (process.env.DOSQL_RELEASE_GATE_TEST_MODE !== "1") {
      throw new Error("test DoSql evidence paths require DOSQL_RELEASE_GATE_TEST_MODE=1");
    }
    return {
      journal: required(value.testJournal, "test-journal"),
      evidence: required(value.testEvidence, "test-evidence"),
    };
  }
  const root = realpathSync(required(value.canonicalRoot, "canonical-root"));
  return {
    journal: resolveCanonicalPath(value.journal, root, "journal"),
    evidence: resolveCanonicalPath(value.evidence, root, "evidence"),
  };
}

function resolveCanonicalPath(path, root, field) {
  const requested = resolve(required(path, field));
  if (!existsSync(requested)) {
    throw new Error(`canonical DoSql ${field} is missing: ${requested}`);
  }
  const resolved = realpathSync(requested);
  const nestedPath = relative(root, resolved);
  if (!nestedPath || nestedPath.startsWith(`..${sep}`) || nestedPath === "..") {
    throw new Error(`${field} must be inside the canonical DoSql audit root`);
  }
  return resolved;
}

function readJournal(path) {
  const events = readFileSync(path, "utf8").split("\n").filter(Boolean).map(JSON.parse);
  let previousEventHash = "";
  for (const event of events) {
    if (event.previousEventHash !== previousEventHash) {
      throw new Error("DoSql journal hash chain is broken");
    }
    const claimed = required(event.eventHash, "journal.eventHash");
    const payload = { ...event };
    delete payload.eventHash;
    const actual = `sha256:${sha256(stableJson(payload))}`;
    if (claimed !== actual) throw new Error("DoSql journal event hash is invalid");
    previousEventHash = claimed;
  }
  return events;
}

function verifyEvidence(evidence, fingerprint, expectedValue) {
  requireEqual(evidence.schema, "dosql.database-operation-evidence.v1", "evidence.schema");
  requireEqual(evidence.status, "verified", "evidence.status");
  requireEqual(evidence.operationId, expectedValue.operationId, "evidence.operationId");
  requireEqual(evidence.changeRequestId, expectedValue.changeId, "evidence.changeRequestId");
  requireEqual(evidence.databaseAssetId, expectedValue.target, "evidence.databaseAssetId");
  requireEqual(evidence.environmentId, expectedValue.mode, "evidence.environmentId");
  requireEqual(evidence.projectId, "agentsmesh", "evidence.projectId");
  requireEqual(evidence.release?.cluster, "doops-oilan", "evidence.release.cluster");
  requireEqual(evidence.release?.doopsTarget, "gw-oilan-node", "evidence.release.doopsTarget");
  requireEqual(evidence.release?.namespace, "agentsmesh", "evidence.release.namespace");
  requireEqual(evidence.release?.session, expectedValue.session, "evidence.release.session");
  requireEqual(Number(evidence.release?.migrationVersion), expectedValue.version, "evidence.release.migrationVersion");
  requireEqual(evidence.executionResult?.status, "succeeded", "evidence.executionResult.status");
  requireEqual(evidence.verificationResult?.status, "succeeded", "evidence.verificationResult.status");
  requireSha256(evidence.release?.versionSource?.statementFingerprint, "evidence.release.versionSource.statementFingerprint");
  requireSha256(evidence.release?.versionSource?.resultFingerprint, "evidence.release.versionSource.resultFingerprint");
  requireSha256(fingerprint, "evidence artifact fingerprint");
}

function verifyLifecycle(events, evidenceFingerprint, sources, expectedValue) {
  const operationEvents = events.filter((event) =>
    event.operationId === expectedValue.operationId &&
    event.changeRequestId === expectedValue.changeId
  );
  if (operationEvents.some((event) =>
    event.status === "failed" || event.status === "verification_failed"
  )) {
    throw new Error("DoSql journal operation contains a failed lifecycle event");
  }
  const statuses = ["planned", "running", "succeeded", "verified"];
  let cursor = -1;
  for (const status of statuses) {
    const index = operationEvents.findIndex((event, candidateIndex) =>
      candidateIndex > cursor && event.status === status
    );
    if (index === -1) throw new Error(`DoSql journal is missing ${status} lifecycle event`);
    verifyReleaseEvent(operationEvents[index], expectedValue);
    cursor = index;
  }
  const verified = operationEvents[cursor];
  requireEqual(resolveEvidenceRef(verified.evidenceRef, sources.journal), realpathSync(sources.evidence), "journal.evidenceRef");
  requireEqual(verified.evidenceArtifactFingerprint, evidenceFingerprint, "journal.evidenceArtifactFingerprint");
}

function verifyReleaseEvent(event, expectedValue) {
  requireEqual(event.databaseAssetId, expectedValue.target, "journal.databaseAssetId");
  requireEqual(event.environmentId, expectedValue.mode, "journal.environmentId");
  requireEqual(event.projectId, "agentsmesh", "journal.projectId");
  requireEqual(event.release?.cluster, "doops-oilan", "journal.release.cluster");
  requireEqual(event.release?.doopsTarget, "gw-oilan-node", "journal.release.doopsTarget");
  requireEqual(event.release?.namespace, "agentsmesh", "journal.release.namespace");
  requireEqual(event.release?.session, expectedValue.session, "journal.release.session");
  requireEqual(Number(event.release?.migrationVersion), expectedValue.version, "journal.release.migrationVersion");
}

function resolveEvidenceRef(ref, journalPath) {
  const text = required(ref, "journal.evidenceRef");
  return realpathSync(text.startsWith("/") ? text : resolve(dirname(journalPath), text));
}

function parseArgs(argv) {
  const parsed = {};
  for (let index = 0; index < argv.length; index += 2) {
    const key = argv[index];
    if (!key?.startsWith("--")) throw new Error(`unexpected argument: ${key}`);
    parsed[toCamel(key.slice(2))] = argv[index + 1];
  }
  return parsed;
}

function toCamel(value) {
  return value.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase());
}

function required(value, field) {
  const text = value === undefined || value === null ? "" : String(value).trim();
  if (!text) throw new Error(`${field} is required`);
  return text;
}

function requireEqual(actual, expectedValue, field) {
  if (actual !== expectedValue) throw new Error(`${field} must be ${expectedValue}`);
}

function requireSha256(value, field) {
  if (!/^sha256:[a-f0-9]{64}$/.test(required(value, field))) {
    throw new Error(`${field} must be a sha256 digest`);
  }
}

function stableJson(value) {
  return JSON.stringify(sortObject(value));
}

function sortObject(value) {
  if (Array.isArray(value)) return value.map(sortObject);
  if (!value || typeof value !== "object") return value;
  return Object.fromEntries(
    Object.entries(value)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, entry]) => [key, sortObject(entry)]),
  );
}

function sha256(value) {
  return createHash("sha256").update(String(value)).digest("hex");
}

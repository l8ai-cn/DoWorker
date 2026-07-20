#!/usr/bin/env node
import { createHash } from "node:crypto";
import { mkdirSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";

const args = parseArgs(process.argv.slice(2));
const root = resolve(required(args.root, "root"));
const fields = {
  target: required(args.target, "target"),
  mode: required(args.mode, "mode"),
  session: required(args.session, "session"),
  changeId: required(args.changeId, "change-id"),
  operationId: required(args.operationId, "operation-id"),
  version: Number(required(args.version, "version")),
};
const fixtureCase = args.case ?? "valid";
mkdirSync(root, { recursive: true });

const evidencePath = join(root, "evidence.json");
const evidence = {
  schema: "dosql.database-operation-evidence.v1",
  status: "verified",
  operationId: fields.operationId,
  changeRequestId: fields.changeId,
  mode: "execute",
  projectId: "agentsmesh",
  databaseAssetId: fields.target,
  environmentId: fields.mode,
  executionResult: { status: "succeeded" },
  verificationResult: { status: "succeeded" },
  release: {
    cluster: "doops-oilan",
    doopsTarget: "gw-oilan-node",
    namespace: "agentsmesh",
    session: fields.session,
    migrationVersion: fields.version,
    versionSource: {
      statementFingerprint: `sha256:${"a".repeat(64)}`,
      resultFingerprint: `sha256:${"b".repeat(64)}`,
    },
  },
};
if (fixtureCase === "target-mismatch") evidence.databaseAssetId = "other-postgres";
if (fixtureCase === "mode-mismatch") evidence.environmentId = "other-mode";
if (fixtureCase === "session-mismatch") evidence.release.session = "other-session";
if (fixtureCase === "change-mismatch") evidence.changeRequestId = "other-change";
if (fixtureCase === "stale-version") evidence.release.migrationVersion = 1;
const evidenceText = `${JSON.stringify(evidence, null, 2)}\n`;
writeFileSync(evidencePath, evidenceText);
const fingerprint = fixtureCase === "fingerprint-mismatch"
  ? `sha256:${"c".repeat(64)}`
  : `sha256:${sha256(evidenceText)}`;

const statuses = fixtureCase === "missing-running"
  ? ["planned", "succeeded", "verified"]
  : ["planned", "running", "succeeded", "verified"];
const events = [];
for (const [index, status] of statuses.entries()) {
  const event = {
    operationId: fields.operationId,
    changeRequestId: evidence.changeRequestId,
    projectId: "agentsmesh",
    databaseAssetId: evidence.databaseAssetId,
    environmentId: evidence.environmentId,
    release: {
      cluster: "doops-oilan",
      doopsTarget: "gw-oilan-node",
      namespace: "agentsmesh",
      session: evidence.release.session,
      migrationVersion: evidence.release.migrationVersion,
    },
    eventType: `database.operation.${status}`,
    status,
    createdAt: `2026-07-20T00:00:0${index}.000Z`,
    previousEventHash: events.at(-1)?.eventHash ?? "",
  };
  if (status === "verified") {
    event.evidenceRef = evidencePath;
    event.evidenceArtifactFingerprint = fingerprint;
  }
  event.eventHash = `sha256:${sha256(stableJson(event))}`;
  events.push(event);
}
if (fixtureCase === "broken-chain") events[1].previousEventHash = "sha256:broken";
writeFileSync(join(root, "journal.jsonl"), `${events.map(JSON.stringify).join("\n")}\n`);

function parseArgs(argv) {
  const parsed = {};
  for (let index = 0; index < argv.length; index += 2) {
    const key = argv[index];
    if (!key?.startsWith("--")) throw new Error(`unexpected argument: ${key}`);
    parsed[key.slice(2).replace(/-([a-z])/g, (_, letter) => letter.toUpperCase())] = argv[index + 1];
  }
  return parsed;
}

function required(value, field) {
  const text = value === undefined || value === null ? "" : String(value).trim();
  if (!text) throw new Error(`${field} is required`);
  return text;
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

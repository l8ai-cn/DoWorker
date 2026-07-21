import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import { readFile } from "node:fs/promises";
import { homedir } from "node:os";
import { resolve } from "node:path";

import {
  buildOilanPostgresRemoteCommand,
  resolveOilanPostgresQuery,
} from "./oilan-postgres-doops-registration.mjs";
import { parseOilanDoopsRows } from "./oilan-postgres-query-result.mjs";

const DOOPS_BIN = resolve(homedir(), ".local/bin/doops");
const DOOPS_CONFIG = resolve(homedir(), ".agent/skills/doops/config.json");
const DOOPS_AUTH = resolve(homedir(), ".agent/skills/doops/auth.json");

export async function verifyOilanPostgresGatewayAudit(registration, dependencies = {}) {
  const queryAudit = dependencies.queryAudit ?? queryCanonicalGatewayAudit;
  const queryWorkspaceAudit = dependencies.queryWorkspaceAudit ?? queryCanonicalWorkspaceAudit;
  const assetProbe = await verifyReference({
    registration,
    reference: registration.registration,
    queryName: "asset-probe",
    expectedTail: `${registration.databaseName}|${registration.serverVersionNum}|t\n`,
    queryAudit,
  });
  const migrationState = await verifyReference({
    registration,
    reference: registration.migrationState,
    queryName: "migration-version",
    expectedTail: `${registration.migrationState.version}|${registration.migrationState.dirty ? "t" : "f"}\n`,
    queryAudit,
  });
  const workspaceAudits = [
    expectedWorkspaceAudit(registration.registration, "asset-probe"),
    expectedWorkspaceAudit(registration.migrationState, "migration-version"),
  ];
  const actualWorkspaceAudits = await queryWorkspaceAudit(workspaceAudits);
  for (const expected of workspaceAudits) {
    const matches = actualWorkspaceAudits.filter((item) => item.session === expected.session);
    if (matches.length !== 1 || matches[0].digest !== expected.digest) {
      throw new Error(`DoOps workspace audit digest does not match session ${expected.session}`);
    }
  }
  return {
    databaseAssetId: registration.databaseAssetId,
    registrationStatus: registration.registrationStatus,
    assetProbe,
    migrationState,
  };
}

function expectedWorkspaceAudit(reference, queryName) {
  const command = buildOilanPostgresRemoteCommand(resolveOilanPostgresQuery(queryName));
  const completedAt = reference.gatewayAudit.endedAt.replace("T", " ").replace(/Z$/, "");
  const content = `# [${completedAt}] exit=0\n${command}\n`;
  return {
    session: reference.session,
    digest: `sha256:${createHash("sha256").update(content).digest("hex")}`,
  };
}

async function verifyReference({ registration, reference, queryName, expectedTail, queryAudit }) {
  const events = await queryAudit({
    cluster: reference.gatewayAudit.cluster,
    instance: reference.gatewayAudit.instance,
    session: reference.session,
  });
  const matches = events.filter((event) => Number(event.id) === reference.gatewayAudit.eventId);
  if (matches.length !== 1) {
    throw new Error(`DoOps Gateway audit event ${reference.gatewayAudit.eventId} is not unique`);
  }
  const event = matches[0];
  requireEqual(event.cluster, reference.gatewayAudit.cluster, "gatewayAudit.cluster");
  requireEqual(event.instance, reference.gatewayAudit.instance, "gatewayAudit.instance");
  requireEqual(event.action, "exec", "gatewayAudit.action");
  requireEqual(event.session, reference.session, "gatewayAudit.session");
  requireEqual(event.status, "success", "gatewayAudit.status");
  requireEqual(event.started_at, reference.gatewayAudit.startedAt, "gatewayAudit.startedAt");
  requireEqual(event.ended_at, reference.gatewayAudit.endedAt, "gatewayAudit.endedAt");
  requireEqual(event.tail, expectedTail, "gatewayAudit.tail");
  const command = buildOilanPostgresRemoteCommand(resolveOilanPostgresQuery(queryName));
  requireEqual(event.command_summary, command.slice(-512), "gatewayAudit.commandSummary");
  return {
    eventId: Number(event.id),
    cluster: event.cluster,
    instance: event.instance,
    session: event.session,
    startedAt: event.started_at,
    endedAt: event.ended_at,
    result: queryName === "asset-probe"
      ? {
          databaseName: registration.databaseName,
          serverVersionNum: registration.serverVersionNum,
          schemaMigrationsPresent: true,
        }
      : {
          version: registration.migrationState.version,
          dirty: registration.migrationState.dirty,
        },
  };
}

async function queryCanonicalGatewayAudit(filter) {
  const [configText, authText] = await Promise.all([
    readFile(DOOPS_CONFIG, "utf8"),
    readFile(DOOPS_AUTH, "utf8"),
  ]);
  const config = JSON.parse(configText);
  const auth = JSON.parse(authText);
  const target = (config.servers ?? []).find((item) => item.name === "gw-oilan-node");
  requireEqual(target?.gateway, "https://doops.l8ai.cn", "doops.gateway");
  requireEqual(target?.cluster, filter.cluster, "doops.cluster");
  requireEqual(target?.instance, filter.instance, "doops.instance");
  const token = auth.tokens?.[target.gateway];
  if (!token) throw new Error("canonical DoOps Gateway login is required");

  const url = new URL("/v1/audit", target.gateway);
  url.searchParams.set("cluster", filter.cluster);
  url.searchParams.set("instance", filter.instance);
  url.searchParams.set("session", filter.session);
  url.searchParams.set("action", "exec");
  url.searchParams.set("limit", "20");
  const response = await fetch(url, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!response.ok) {
    throw new Error(`DoOps Gateway audit query failed with status ${response.status}`);
  }
  const body = await response.json();
  if (!Array.isArray(body.events)) {
    throw new Error("DoOps Gateway audit response is invalid");
  }
  return body.events;
}

function queryCanonicalWorkspaceAudit(expected) {
  const sessions = expected.map((item) => item.session);
  const session = `agentsmesh-pg-registration-gate-${process.pid}-${Date.now()}`;
  const command = [
    "set -euo pipefail",
    ...sessions.map((value) => [
      `path="/root/ws/${value}/.doops-audit-log"`,
      'test -f "$path"',
      `printf '%s|' '${value}'`,
      'sha256sum "$path" | awk \'{print $1}\'',
    ].join("\n")),
  ].join("\n");
  const result = spawnSync(DOOPS_BIN, [
    "-session",
    session,
    "exec",
    "--target",
    "gw-oilan-node",
    "--cmd",
    command,
  ], {
    encoding: "utf8",
    env: { ...process.env, DOOPS_CONFIG },
  });
  if (result.error || result.status !== 0) {
    throw new Error("DoOps workspace audit verification failed");
  }
  return parseOilanDoopsRows(result.stdout).map((row) => {
    const [auditSession, digest] = row.split("|");
    if (!auditSession || !/^[a-f0-9]{64}$/.test(digest ?? "")) {
      throw new Error("DoOps workspace audit response is invalid");
    }
    return { session: auditSession, digest: `sha256:${digest}` };
  });
}

function requireEqual(actual, expected, fieldName) {
  if (actual !== expected) {
    throw new Error(`${fieldName} does not match the canonical DoOps Gateway audit`);
  }
}

import assert from "node:assert/strict";
import test from "node:test";

import { verifyOilanPostgresGatewayAudit } from "../Server/lib/oilan-postgres-gateway-audit.mjs";
import {
  buildOilanPostgresRemoteCommand,
  loadOilanPostgresRegistration,
  resolveOilanPostgresQuery,
} from "../Server/lib/oilan-postgres-doops-registration.mjs";

test("verifies registration against authoritative DoOps Gateway audit events", async () => {
  const registration = await loadOilanPostgresRegistration();
  const events = auditEvents(registration);
  const result = await verifyOilanPostgresGatewayAudit(registration, {
    queryAudit: ({ session }) => events.filter((event) => event.session === session),
    queryWorkspaceAudit: (expected) => expected,
  });
  assert.equal(result.assetProbe.eventId, 423834);
  assert.deepEqual(result.migrationState.result, {
    version: registration.migrationState.version,
    dirty: registration.migrationState.dirty,
  });
});

test("rejects a Gateway event whose command was not the fixed adapter command", async () => {
  const registration = await loadOilanPostgresRegistration();
  const events = auditEvents(registration);
  events[0].command_summary = "printf forged";
  await assert.rejects(
    verifyOilanPostgresGatewayAudit(registration, {
      queryAudit: ({ session }) => events.filter((event) => event.session === session),
      queryWorkspaceAudit: (expected) => expected,
    }),
    /gatewayAudit.commandSummary/,
  );
});

test("rejects a changed full workspace audit command", async () => {
  const registration = await loadOilanPostgresRegistration();
  const events = auditEvents(registration);
  await assert.rejects(
    verifyOilanPostgresGatewayAudit(registration, {
      queryAudit: ({ session }) => events.filter((event) => event.session === session),
      queryWorkspaceAudit: (expected) => expected.map((item, index) => (
        index === 0 ? { ...item, digest: `sha256:${"0".repeat(64)}` } : item
      )),
    }),
    /workspace audit digest/,
  );
});

function auditEvents(registration) {
  return [
    event({
      reference: registration.registration,
      queryName: "asset-probe",
      tail: "agentcloud|160014|t\n",
    }),
    event({
      reference: registration.migrationState,
      queryName: "migration-version",
      tail: `${registration.migrationState.version}|${registration.migrationState.dirty ? "t" : "f"}\n`,
    }),
  ];
}

function event({ reference, queryName, tail }) {
  const command = buildOilanPostgresRemoteCommand(resolveOilanPostgresQuery(queryName));
  return {
    id: reference.gatewayAudit.eventId,
    cluster: reference.gatewayAudit.cluster,
    instance: reference.gatewayAudit.instance,
    action: "exec",
    session: reference.session,
    status: "success",
    command_summary: command.slice(-512),
    tail,
    started_at: reference.gatewayAudit.startedAt,
    ended_at: reference.gatewayAudit.endedAt,
  };
}

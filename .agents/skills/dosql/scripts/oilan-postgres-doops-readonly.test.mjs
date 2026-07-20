import assert from "node:assert/strict";
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import test from "node:test";

import {
  executeOilanPostgresReadOnly,
  verifyOilanPostgresReadOnlyEvidence,
} from "../Server/lib/oilan-postgres-doops-readonly.mjs";
import { loadOilanPostgresRegistration } from "../Server/lib/oilan-postgres-doops-registration.mjs";
import { sha256, stableJson } from "../Server/lib/oilan-postgres-doops-evidence.mjs";

const CLI = new URL("./oilan-postgres-doops-readonly.mjs", import.meta.url).pathname;
const INPUT = {
  operationId: "dbop_oilan_probe_001",
  session: "oilan-read-20260720-001",
  queryName: "asset-probe",
};

test("runs the fixed Oilan PostgreSQL probe through DoOps and writes redacted evidence", async () => {
  const auditRoot = await mkdtemp(join(tmpdir(), "dosql-oilan-readonly-"));
  try {
    const registration = await loadOilanPostgresRegistration();
    let request;
    const evidence = await executeOilanPostgresReadOnly(INPUT, {
      auditRoot,
      registration,
      now: clock(),
      execute(value) {
        request = value;
        return { status: 0, stdout: "postgres://hidden-password@postgres/agentsmesh\nprivate-row\n" };
      },
    });

    assert.equal(request.command, "doops");
    assert.deepEqual(request.args.slice(0, 6), [
      "-session",
      INPUT.session,
      "exec",
      "--target",
      "gw-oilan-node",
      "--cmd",
    ]);
    assert.match(request.args[6], /namespace=agentsmesh/);
    assert.match(request.args[6], /service=postgres/);
    assert.match(request.args[6], /secret=agentsmesh-secrets/);
    assert.equal(request.args[6].includes("postgresql://"), false);
    assert.equal(JSON.stringify(evidence).includes("hidden-password"), false);
    assert.equal(JSON.stringify(evidence).includes("private-row"), false);
    assert.equal(evidence.status, "verified");
    assert.equal(verifyOilanPostgresReadOnlyEvidence(evidence), true);

    const stored = JSON.parse(await readFile(join(auditRoot, "readonly-evidence", `${INPUT.operationId}.json`), "utf8"));
    assert.deepEqual(stored, evidence);
    const events = (await readFile(join(auditRoot, "readonly-journal", `${INPUT.operationId}.jsonl`), "utf8"))
      .trim()
      .split("\n")
      .map(JSON.parse);
    assert.deepEqual(events.map((event) => event.status), ["planned", "running", "succeeded", "verified"]);
    assertHashChain(events);
  } finally {
    await rm(auditRoot, { recursive: true, force: true });
  }
});

test("fails closed for unavailable DoOps and only persists redacted failure evidence", async () => {
  const auditRoot = await mkdtemp(join(tmpdir(), "dosql-oilan-readonly-"));
  try {
    const registration = await loadOilanPostgresRegistration();
    await assert.rejects(
      executeOilanPostgresReadOnly(INPUT, {
        auditRoot,
        registration,
        now: clock(),
        execute: () => ({ status: null, error: new Error("target offline"), stderr: "postgres://secret" }),
      }),
      /Oilan PostgreSQL read-only query failed/,
    );
    const evidence = JSON.parse(await readFile(join(auditRoot, "readonly-evidence", `${INPUT.operationId}.json`), "utf8"));
    assert.equal(evidence.status, "failed");
    assert.equal(JSON.stringify(evidence).includes("secret"), false);
    assert.equal(verifyOilanPostgresReadOnlyEvidence(evidence), true);
  } finally {
    await rm(auditRoot, { recursive: true, force: true });
  }
});

test("rejects caller SQL, URI and non-fixed registrations before invoking DoOps", async () => {
  const auditRoot = await mkdtemp(join(tmpdir(), "dosql-oilan-readonly-"));
  const registration = await loadOilanPostgresRegistration();
  let invoked = false;
  try {
    await assert.rejects(
      executeOilanPostgresReadOnly({ ...INPUT, statement: "select 1" }, {
        auditRoot,
        registration,
        execute: () => {
          invoked = true;
          return { status: 0, stdout: "" };
        },
      }),
      /unsupported read-only query input: statement/,
    );
    await assert.rejects(
      executeOilanPostgresReadOnly({ ...INPUT, connectionUri: "postgres://caller-supplied" }, {
        auditRoot,
        registration,
        execute: () => ({ status: 0, stdout: "" }),
      }),
      /unsupported read-only query input: connectionUri/,
    );
    await assert.rejects(
      executeOilanPostgresReadOnly(INPUT, {
        auditRoot,
        registration: { ...registration, doopsTarget: "other-target" },
        execute: () => ({ status: 0, stdout: "" }),
      }),
      /fixed Oilan PostgreSQL registration/,
    );
    assert.equal(invoked, false);
  } finally {
    await rm(auditRoot, { recursive: true, force: true });
  }
});

test("evidence verifier rejects a fingerprint mismatch", async () => {
  const auditRoot = await mkdtemp(join(tmpdir(), "dosql-oilan-readonly-"));
  try {
    const evidence = await executeOilanPostgresReadOnly(INPUT, {
      auditRoot,
      registration: await loadOilanPostgresRegistration(),
      now: clock(),
      execute: () => ({ status: 0, stdout: "version=231\ndirty=false\n" }),
    });
    assert.throws(
      () => verifyOilanPostgresReadOnlyEvidence({ ...evidence, resultFingerprint: sha256("tampered") }),
      /evidence fingerprint is invalid/,
    );
  } finally {
    await rm(auditRoot, { recursive: true, force: true });
  }
});

test("CLI returns a failed response instead of accepting caller-supplied SQL", async () => {
  const dir = await mkdtemp(join(tmpdir(), "dosql-oilan-readonly-cli-"));
  try {
    const inputPath = join(dir, "input.json");
    const outputPath = join(dir, "output.json");
    await writeFile(inputPath, JSON.stringify({
      ...INPUT,
      queryName: "migration-version",
      statement: "select 1",
    }));
    const result = spawnSync(process.execPath, [CLI, "query", "--input", inputPath, "--output", outputPath], {
      encoding: "utf8",
    });
    const output = JSON.parse(await readFile(outputPath, "utf8"));
    assert.equal(result.status, 1);
    assert.equal(output.status, "failed");
    assert.match(output.error.message, /unsupported read-only query input/);
  } finally {
    await rm(dir, { recursive: true, force: true });
  }
});

test("CLI keeps probe and query command scopes fixed", async () => {
  const dir = await mkdtemp(join(tmpdir(), "dosql-oilan-readonly-cli-"));
  try {
    const inputPath = join(dir, "input.json");
    const outputPath = join(dir, "output.json");
    await writeFile(inputPath, JSON.stringify({ ...INPUT, queryName: "asset-probe" }));
    const result = spawnSync(process.execPath, [CLI, "query", "--input", inputPath, "--output", outputPath], {
      encoding: "utf8",
    });
    const output = JSON.parse(await readFile(outputPath, "utf8"));
    assert.equal(result.status, 1);
    assert.match(output.error.message, /queryName must be migration-version/);

    await writeFile(inputPath, JSON.stringify({ ...INPUT, queryName: "migration-version" }));
    const probeResult = spawnSync(process.execPath, [CLI, "probe", "--input", inputPath, "--output", outputPath], {
      encoding: "utf8",
    });
    const probeOutput = JSON.parse(await readFile(outputPath, "utf8"));
    assert.equal(probeResult.status, 1);
    assert.match(probeOutput.error.message, /probe does not accept queryName/);
  } finally {
    await rm(dir, { recursive: true, force: true });
  }
});

function assertHashChain(events) {
  let previous = "";
  for (const event of events) {
    assert.equal(event.previousEventHash, previous);
    const payload = { ...event };
    delete payload.eventHash;
    assert.equal(event.eventHash, sha256(stableJson(payload)));
    previous = event.eventHash;
  }
}

function clock() {
  let second = 0;
  return () => new Date(Date.UTC(2026, 6, 20, 0, 0, second++));
}

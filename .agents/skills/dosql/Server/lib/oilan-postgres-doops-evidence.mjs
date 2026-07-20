import { createHash } from "node:crypto";
import { appendFile, chmod, lstat, mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";

export async function prepareOilanPostgresAudit(auditRoot, operationId) {
  const root = resolve(auditRoot);
  await mkdir(root, { recursive: true, mode: 0o700 });
  const metadata = await lstat(root);
  if (!metadata.isDirectory() || metadata.isSymbolicLink()) {
    throw new Error("DoSql audit root must be a real directory");
  }
  await chmod(root, 0o700);
  const journalPath = resolve(root, "readonly-journal", `${operationId}.jsonl`);
  const evidencePath = resolve(root, "readonly-evidence", `${operationId}.json`);
  return {
    journalPath,
    evidencePath,
    journalRef: `readonly-journal/${operationId}.jsonl`,
    evidenceRef: `readonly-evidence/${operationId}.json`,
  };
}

export async function appendOilanPostgresAuditEvent(path, event) {
  await mkdir(dirname(path), { recursive: true, mode: 0o700 });
  const previousEventHash = await lastEventHash(path);
  const payload = { ...event, previousEventHash };
  const record = { ...payload, eventHash: sha256(stableJson(payload)) };
  await appendFile(path, `${JSON.stringify(record)}\n`, "utf8");
  return record;
}

export async function writeOilanPostgresEvidence(path, evidence) {
  await mkdir(dirname(path), { recursive: true, mode: 0o700 });
  await writeFile(path, `${JSON.stringify(evidence, null, 2)}\n`, {
    encoding: "utf8",
    mode: 0o600,
  });
  await chmod(path, 0o600);
}

export function verifyOilanPostgresReadOnlyEvidence(evidence) {
  if (evidence?.schema !== "dosql.oilan-postgres-readonly-evidence.v1") {
    throw new Error("evidence schema is invalid");
  }
  if (!["verified", "failed"].includes(evidence.status)) {
    throw new Error("evidence status is invalid");
  }
  const claimed = requiredDigest(evidence.evidenceFingerprint, "evidenceFingerprint");
  const payload = { ...evidence };
  delete payload.evidenceFingerprint;
  if (claimed !== sha256(stableJson(payload))) {
    throw new Error("evidence fingerprint is invalid");
  }
  return true;
}

export function sha256(value) {
  return `sha256:${createHash("sha256").update(String(value)).digest("hex")}`;
}

export function stableJson(value) {
  if (Array.isArray(value)) return `[${value.map(stableJson).join(",")}]`;
  if (value && typeof value === "object") {
    return `{${Object.keys(value).sort().map((key) => `${JSON.stringify(key)}:${stableJson(value[key])}`).join(",")}}`;
  }
  return JSON.stringify(value);
}

async function lastEventHash(path) {
  try {
    const lines = (await readFile(path, "utf8")).trim().split("\n").filter(Boolean);
    return lines.length === 0 ? "" : requiredDigest(JSON.parse(lines.at(-1)).eventHash, "eventHash");
  } catch (error) {
    if (error?.code === "ENOENT") return "";
    throw error;
  }
}

function requiredDigest(value, fieldName) {
  const text = value === undefined || value === null ? "" : String(value).trim();
  if (!/^sha256:[a-f0-9]{64}$/.test(text)) {
    throw new Error(`${fieldName} must be a sha256 fingerprint`);
  }
  return text;
}

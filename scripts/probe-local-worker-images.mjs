import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const definitionsRoot = path.join(root, "config/worker-types");
const evidenceRoot = path.join(
  root,
  "tools/loops/worker-onboarding/catalog-loop/evidence/local-image-probes",
);
const catalog = readJson(path.join(definitionsRoot, "catalog.json"));
const selectedWorker = process.argv[2];
const observedAt = process.env.RUNTIME_OBSERVED_AT ?? new Date().toISOString();

fs.mkdirSync(evidenceRoot, { recursive: true });

for (const entry of catalog.worker_types) {
  if (selectedWorker && entry.slug !== selectedWorker) continue;
  probeWorker(entry);
}

function probeWorker(entry) {
  const sourceDir = path.join(root, path.dirname(entry.definition_path));
  const definition = readJson(path.join(sourceDir, "definition.json"));
  const imageReference = `do-worker/runner-${definition.image.runtime}:latest`;
  const probe = definition.image.version_probe;
  const image = docker(["image", "inspect", "-f", "{{.Id}}", imageReference]);

  if (image.status !== 0) {
    writeEvidence(entry, {
      image_reference: imageReference,
      probe_command: probe,
      status: "image_missing",
      exit_code: image.status,
      output: outputOf(image),
    });
    return;
  }

  const run = docker([
    "run",
    "--rm",
    "--pull=never",
    "--platform",
    "linux/amd64",
    "--entrypoint",
    definition.executable,
    imageReference,
    ...probe.slice(1),
  ]);
  writeEvidence(entry, {
    image_reference: imageReference,
    image_id: image.stdout.trim(),
    probe_command: probe,
    status: run.status === 0 ? "passed" : "probe_failed",
    exit_code: run.status,
    output: outputOf(run),
  });
}

function writeEvidence(entry, result) {
  const document = {
    schema_version: 1,
    worker_slug: entry.slug,
    definition_hash: entry.definition_hash,
    platform: "linux/amd64",
    observed_at: observedAt,
    ...result,
  };
  fs.writeFileSync(
    path.join(evidenceRoot, `${entry.slug}.json`),
    JSON.stringify(document, null, 2) + "\n",
  );
  console.log(`${entry.slug}: ${document.status}`);
}

function docker(args) {
  return spawnSync("docker", args, { encoding: "utf8" });
}

function outputOf(result) {
  return [result.stdout, result.stderr].filter(Boolean).join("").trim();
}

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

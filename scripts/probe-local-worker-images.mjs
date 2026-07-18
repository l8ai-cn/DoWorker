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

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}

export function localWorkerImageReference(composeProjectName, runtime) {
  return `${composeProjectName}-runner-${runtime}:latest`;
}

function main() {
  const composeProjectName = readDevComposeProjectName();
  fs.mkdirSync(evidenceRoot, { recursive: true });
  for (const entry of catalog.worker_types) {
    probeWorker(entry, composeProjectName);
  }
}

function probeWorker(entry, composeProjectName) {
  const sourceDir = path.join(root, path.dirname(entry.definition_path));
  const definition = readJson(path.join(sourceDir, "definition.json"));
  const imageReference = localWorkerImageReference(
    composeProjectName,
    definition.image.runtime,
  );
  const probe = definition.image.version_probe;
  const image = docker(["image", "inspect", "-f", "{{.Id}}", imageReference]);

  if (image.status !== 0) {
    writeEvidence(entry, {
      image_reference: imageReference,
      probe_command: probe,
      network: "none",
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
    "--network",
    "none",
    "--entrypoint",
    definition.executable,
    imageReference,
    ...probe.slice(1),
  ]);
  writeEvidence(entry, {
    image_reference: imageReference,
    image_id: image.stdout.trim(),
    probe_command: probe,
    network: "none",
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
    observed_at: new Date().toISOString(),
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

function readDevComposeProjectName() {
  const environment = fs.readFileSync(path.join(root, "deploy/dev/.env"), "utf8");
  const match = environment.match(/^COMPOSE_PROJECT_NAME=([a-z0-9][a-z0-9_-]*)$/m);
  if (!match) {
    throw new Error("deploy/dev/.env must define COMPOSE_PROJECT_NAME");
  }
  return match[1];
}

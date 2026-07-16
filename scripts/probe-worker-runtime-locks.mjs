import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const lock = readJson(
  path.join(root, "backend/internal/domain/workerruntime/runtime_catalog.lock.json"),
);
const evidenceRoot = path.join(
  root,
  "tools/loops/worker-onboarding/catalog-loop/evidence/runtime-lock-probes",
);
const selectedWorker = process.argv[2];
const runtimePlatform = process.env.RUNTIME_PLATFORM ?? "linux/amd64";
const aggregatePath = path.join(evidenceRoot, "..", "runtime-lock-probes.json");
const probes = new Map(
  selectedWorker && fs.existsSync(aggregatePath)
    ? readJson(aggregatePath).probes.map((probe) => [probe.worker_slug, probe])
    : [],
);

fs.mkdirSync(evidenceRoot, { recursive: true });

for (const image of lock.images) {
  if (
    selectedWorker &&
    !image.worker_type_slugs.includes(selectedWorker)
  ) {
    continue;
  }
  const result = spawnSync("docker", [
    "pull",
    "--platform",
    runtimePlatform,
    image.reference,
  ], {
    encoding: "utf8",
  });
  const output = [result.stdout, result.stderr].filter(Boolean).join("").trim();
  const status =
    result.status === 0
      ? "available"
      : output.includes("not found")
        ? "not_found"
        : "unavailable";

  for (const slug of image.worker_type_slugs) {
    writeEvidence(slug, image, status, result.status, output);
  }
}

function writeEvidence(slug, image, status, exitCode, output) {
  const document = {
    schema_version: 1,
    worker_slug: slug,
    runtime_catalog_revision: lock.revision,
    image_reference: image.reference,
    image_digest: image.digest,
    status,
    exit_code: exitCode,
    output,
    observed_at: new Date().toISOString(),
  };
  probes.set(slug, document);
  fs.writeFileSync(
    path.join(evidenceRoot, `${slug}.json`),
    JSON.stringify(document, null, 2) + "\n",
  );
  console.log(`${slug}: ${status}`);
}

fs.writeFileSync(
  aggregatePath,
  JSON.stringify(
    {
      schema_version: 1,
      runtime_catalog_revision: lock.revision,
      probes: [...probes.values()].sort((left, right) =>
        left.worker_slug.localeCompare(right.worker_slug),
      ),
    },
    null,
    2,
  ) + "\n",
);

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

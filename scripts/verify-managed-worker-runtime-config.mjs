import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const clusterRoot = path.join(root, "deploy/kubernetes/cluster-oilan");
const definitions = readJson(path.join(root, "config/worker-types/catalog.json"));
const lock = readJson(
  path.join(root, "backend/internal/domain/workerruntime/runtime_catalog.lock.json"),
);
const formalSlugs = new Set(definitions.worker_types.map((worker) => worker.slug));
const released = new Map(
  lock.images
    .filter((image) => image.enabled)
    .flatMap((image) => image.worker_type_slugs.map((slug) => [slug, image.reference])),
);
const errors = [];

for (const file of fs.readdirSync(clusterRoot).filter((name) => name.endsWith(".yaml"))) {
  validateFile(path.join(clusterRoot, file), file);
}

if (errors.length > 0) {
  throw new Error(`managed Worker runtime configuration is invalid:\n${errors.join("\n")}`);
}

console.log("managed Worker runtime configuration verified");

function validateFile(filePath, label) {
  const content = fs.readFileSync(filePath, "utf8");
  for (const reference of content.matchAll(/^\s*image:\s*([^\s#]+)\s*$/gm)) {
    validateReference(label, reference[1]);
  }
  const mapping = content.match(
    /name:\s*COORDINATOR_RUNNER_IMAGES\s*\n\s*value:\s*"([^"]*)"/m,
  );
  if (!mapping) return;
  for (const item of mapping[1].split(",")) {
    const [slug, reference] = item.split("=");
    if (!slug || !reference) {
      errors.push(`${label}: invalid COORDINATOR_RUNNER_IMAGES entry ${JSON.stringify(item)}`);
      continue;
    }
    validateReference(`${label}:${slug}`, reference, slug);
  }
}

function validateReference(label, reference, mappedSlug) {
  const match = reference.match(/\/runner-([a-z0-9-]+)(?:@|:)/);
  const slug = mappedSlug ?? match?.[1];
  if (slug && formalSlugs.has(slug)) {
    const expected = released.get(slug);
    if (!expected) {
      errors.push(`${label}: formal Worker ${slug} has no released runtime image`);
      return;
    }
    if (reference !== expected) {
      errors.push(`${label}: ${slug} does not match the runtime catalog`);
      return;
    }
  }
  if (reference.includes("/runner-") && !/@sha256:[a-f0-9]{64}$/.test(reference)) {
    errors.push(`${label}: runner image must use an immutable sha256 digest`);
  }
}

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

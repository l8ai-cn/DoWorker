import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  buildDrift,
  buildInventory,
  mapByWorkerSlug,
} from "./worker-loop-inventory-builders.mjs";
import {
  readJson,
} from "./worker-loop-json.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const loopRoot = path.join(
  root,
  "tools/loops/worker-onboarding/catalog-loop",
);
const loopRelativeRoot = "tools/loops/worker-onboarding/catalog-loop";
const definitionsRoot = path.join(root, "config/worker-types");
const checkOnly = process.argv.includes("--check");

const definitionCatalog = readJson(path.join(definitionsRoot, "catalog.json"));
const runtimeCatalog = readJson(
  path.join(root, "backend/internal/domain/workerruntime/runtime_catalog.lock.json"),
);
const evidenceMatrix = readJson(
  path.join(loopRoot, "catalog/worker-evidence-matrix.json"),
);
const lockProbes = mapByWorkerSlug(
  readJson(path.join(loopRoot, "evidence/runtime-lock-probes.json")).probes,
);

const context = {
  definitionCatalog,
  runtimeCatalog,
  evidenceMatrix,
  lockProbes,
  readJson,
  root,
  loopRoot,
  loopRelativeRoot,
};
const outputs = {
  inventory: buildInventory(context),
  drift: buildDrift(context),
};

writeOrCheck("catalog/inventory.json", outputs.inventory);
writeOrCheck("catalog/drift.json", outputs.drift);

function writeOrCheck(relativePath, value) {
  const destination = path.join(loopRoot, relativePath);
  const output = JSON.stringify(value, null, 2) + "\n";
  if (checkOnly) {
    if (!fs.existsSync(destination) || fs.readFileSync(destination, "utf8") !== output) {
      throw new Error(`Worker loop artifact is stale: ${relativePath}`);
    }
    return;
  }
  fs.writeFileSync(destination, output);
}

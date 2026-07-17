import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const definitionsRoot = path.join(root, "config/worker-types");
const runsRoot = path.join(
  root,
  "tools/loops/worker-onboarding/catalog-loop/runs",
);
const templateScriptsRoot = path.join(
  root,
  "tools/loops/worker-onboarding/worker-loop-template/scripts",
);
const checkOnly = process.argv.includes("--check");

const catalog = readJson(path.join(definitionsRoot, "catalog.json"));
const schema = fs.readFileSync(
  path.join(definitionsRoot, "schema/definition.schema.json"),
  "utf8",
);

for (const entry of catalog.worker_types) {
  syncWorker(entry);
}

function syncWorker(entry) {
  const sourceDir = path.join(root, path.dirname(entry.definition_path));
  const runRoot = path.join(runsRoot, entry.slug);
  const definition = readJson(path.join(sourceDir, "definition.json"));
  const definitionText = fs.readFileSync(
    path.join(sourceDir, "definition.json"),
    "utf8",
  );
  const agentFile = fs.readFileSync(path.join(sourceDir, "AgentFile"), "utf8");

  if (!fs.existsSync(runRoot)) {
    throw new Error(`Worker run is not initialized: ${entry.slug}`);
  }
  if (definition.slug !== entry.slug) {
    throw new Error(`Definition slug does not match catalog: ${entry.slug}`);
  }

  syncTemplateScripts(runRoot);
  writeOrCheck(path.join(runRoot, "artifacts/definition.json"), definitionText);
  writeOrCheck(path.join(runRoot, "artifacts/AgentFile"), agentFile);
  writeOrCheck(
    path.join(runRoot, "artifacts/schemas/definition.schema.json"),
    schema,
  );
  writeOrCheck(
    path.join(runRoot, "artifacts/credential-bindings.json"),
    json({
      schema_version: 1,
      worker_slug: entry.slug,
      bindings: definition.credential_bindings,
    }),
  );
  writeOrCheck(
    path.join(runRoot, "artifacts/config-documents.json"),
    json({
      schema_version: 1,
      worker_slug: entry.slug,
      documents: definition.config_documents,
    }),
  );
}

function syncTemplateScripts(runRoot) {
  for (const name of [
    "worker-context.sh",
    "verify-contract.sh",
    "test-definition-bundle-hash.sh",
  ]) {
    writeOrCheck(
      path.join(runRoot, "scripts", name),
      fs.readFileSync(path.join(templateScriptsRoot, name), "utf8"),
    );
  }
}

function writeOrCheck(destination, content) {
  if (checkOnly) {
    if (!fs.existsSync(destination) || fs.readFileSync(destination, "utf8") !== content) {
      throw new Error(`Worker run contract is stale: ${path.relative(root, destination)}`);
    }
    return;
  }
  fs.mkdirSync(path.dirname(destination), { recursive: true });
  fs.writeFileSync(destination, content);
}

function json(value) {
  return JSON.stringify(value, null, 2) + "\n";
}

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

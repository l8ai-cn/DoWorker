import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  hasAvailableRuntime,
  mapRuntimeCatalogEvidence,
} from "./worker-runtime-catalog-evidence.mjs";
import { loadPublicWorkerDefinitions } from "./public-worker-definitions.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const outputPath = path.join(
  root,
  "clients/web/src/generated/worker-runtime-catalog.json",
);
const checkOnly = process.argv.includes("--check");

const definitionsRoot = path.join(root, "config/worker-types");
const definitionCatalog = readJson(path.join(definitionsRoot, "catalog.json"));
const runtimeCatalog = readJson(
  path.join(root, "backend/internal/domain/workerruntime/runtime_catalog.lock.json"),
);
const evidenceMatrix = readJson(
  path.join(
    root,
    "tools/loops/worker-onboarding/catalog-loop/catalog/worker-evidence-matrix.json",
  ),
);
const lockProbes = mapByWorkerSlug(
  readJson(
    path.join(
      root,
      "tools/loops/worker-onboarding/catalog-loop/evidence/runtime-lock-probes.json",
    ),
  ).probes,
);

const workers = buildWorkers();
const output = JSON.stringify(
  {
    schemaVersion: 1,
    runtimeCatalogRevision: runtimeCatalog.revision,
    workers,
  },
  null,
  2,
) + "\n";

if (checkOnly) {
  if (!fs.existsSync(outputPath) || fs.readFileSync(outputPath, "utf8") !== output) {
    throw new Error(
      "Worker documentation catalog is stale. Run: pnpm run worker-docs:sync",
    );
  }
} else {
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, output);
}

function buildWorkers() {
  const evidenceBySlug = new Map(
    evidenceMatrix.workers.map((worker) => [worker.slug, worker]),
  );
  const lockedImages = runtimeCatalog.images;
  const publicDefinitions = loadPublicWorkerDefinitions({
    definitionCatalog,
    readJson,
    root,
  });
  const definitionSlugs = publicDefinitions.map(({ entry }) => entry.slug);

  assertSameSlugs(definitionSlugs, [...evidenceBySlug.keys()], "evidence matrix");

  return publicDefinitions.map(({ entry, definition }) => {
    const evidence = evidenceBySlug.get(entry.slug);
    const runtimeImage = lockedImages.find((image) =>
      image.worker_type_slugs.includes(entry.slug),
    );
    const runtimeCatalogEvidence = mapRuntimeCatalogEvidence(
      runtimeCatalog,
      lockProbes,
      entry.slug,
    );
    const agentFile = fs.readFileSync(
      path.join(definitionsRoot, entry.slug, "AgentFile"),
      "utf8",
    );

    return {
      slug: entry.slug,
      name: workerDisplayName(entry.slug),
      executable: definition.executable,
      adapterId: definition.adapter_id,
      interactionModes: definition.interaction_modes,
      modelRequirement: definition.model_requirement,
      ...(definition.tool_model_requirements?.length
        ? { toolModelRequirements: definition.tool_model_requirements }
        : {}),
      credentialBindings: definition.credential_bindings.map((binding) => ({
        sourceKind: binding.source.kind,
        sourceRef: binding.source.ref,
        environmentVariable: binding.target.name,
      })),
      configFields: parseConfigFields(agentFile, entry.slug),
      configDocuments: definition.config_documents,
      runtimeImage: runtimeImage
        ? {
            name: runtimeImage.name,
            reference: runtimeImage.reference,
            availability: runtimeCatalogEvidence.status,
          }
        : null,
      validationStatus: validationStatus(evidence, runtimeCatalogEvidence),
    };
  });
}

function parseConfigFields(agentFile, slug) {
  return agentFile
    .split("\n")
    .filter((line) => line.startsWith("CONFIG "))
    .map((line) => {
      const match = line.match(/^CONFIG\s+([a-z0-9_]+)\s+(.+?)\s*=\s*(.+)$/i);
      if (!match) {
        throw new Error(`Cannot parse CONFIG declaration for ${slug}: ${line}`);
      }
      const [, name, rawKind, defaultValue] = match;
      const options = rawKind.startsWith("SELECT(")
        ? [...rawKind.matchAll(/"([^"]*)"/g)].map((option) => option[1])
        : [];
      return { name, kind: configKind(rawKind), options, defaultValue };
    });
}

function configKind(rawKind) {
  const kinds = { BOOL: "boolean", STRING: "string", NUMBER: "number", SECRET: "secret" };
  if (rawKind.startsWith("SELECT(")) return "select";
  if (kinds[rawKind]) return kinds[rawKind];
  throw new Error(`Unsupported Worker configuration field kind: ${rawKind}`);
}

function validationStatus(evidence, runtimeCatalogEvidence) {
  if (!hasAvailableRuntime(runtimeCatalogEvidence)) {
    if (evidence.support_status === "verified_local_dev") {
      return "local_evidence_release_blocked";
    }
    if (runtimeCatalogEvidence.status === "invalid_published_digest") {
      return "invalid_published_runtime";
    }
    return "runtime_image_unavailable";
  }
  if (evidence.support_status === "verified_local_dev") {
    return "runtime_ready_unverified";
  }
  if (evidence.browser === "missing_model_resource_guard_verified") {
    return "requires_model_resource";
  }
  return "runtime_ready_unverified";
}

function workerDisplayName(slug) {
  const names = {
    "codex-cli": "Codex CLI",
    "cursor-cli": "Cursor CLI",
    "do-agent": "Do Agent",
    "gemini-cli": "Gemini CLI",
    "grok-build": "Grok Build",
    "minimax-cli": "MiniMax CLI",
    openclaw: "OpenClaw",
    opencode: "OpenCode",
  };
  return names[slug] ?? slug.split("-").map(capitalize).join(" ");
}

function capitalize(value) {
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function assertSameSlugs(left, right, source) {
  const expected = [...left].sort().join(",");
  const actual = [...right].sort().join(",");
  if (expected !== actual) {
    throw new Error(`Worker definitions and ${source} have different slugs`);
  }
}

function mapByWorkerSlug(probes) {
  const result = new Map();
  for (const probe of probes) {
    if (result.has(probe.worker_slug)) {
      throw new Error(`runtime lock probes repeat Worker slug: ${probe.worker_slug}`);
    }
    result.set(probe.worker_slug, probe);
  }
  return result;
}

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

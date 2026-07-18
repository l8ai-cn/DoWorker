import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../../../..");
const loopRoot = path.join(
  root,
  "tools/loops/worker-onboarding/definition-driven-contract",
);
const matrixPath = path.join(loopRoot, "evidence/current-worker-evidence-matrix.json");
const defaultOutput = path.join(
  loopRoot,
  "evidence/worker-integration-inventory-2026-07-16.md",
);

if (process.argv[1] === fileURLToPath(import.meta.url)) main();

function main() {
  const matrix = readJson(matrixPath);
  const output = outputPath(process.argv.slice(2));
  const content = renderInventory(matrix);
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, content);
  console.log(`wrote Worker integration inventory: ${output}`);
}

function renderInventory(matrix) {
  requireMatrix(matrix);
  const workers = [...matrix.workers].sort((left, right) =>
    left.slug.localeCompare(right.slug),
  );
  const runtimeReady = workers.filter(
    (worker) => worker.runtime.evidence_state === "runtime_evidence_only",
  );
  const documents = workers.filter((worker) => worker.config_documents.length > 0);
  const blocked = workers.filter((worker) => worker.runtime.evidence_state === "blocked");
  const lifecyclePassed = workers.filter(
    (worker) => worker.integration_gates.lifecycle === "passed",
  );
  const lifecycleFailed = workers.filter(
    (worker) => worker.integration_gates.lifecycle === "failed",
  );
  const lines = [
    "# Worker Integration Inventory",
    "",
    "Source: `evidence/current-worker-evidence-matrix.json`.",
    "",
    `- Formal Worker types: ${workers.length}`,
    `- Runtime evidence only: ${runtimeReady.length}`,
    `- Runtime blocked: ${blocked.length}`,
    `- Definition-owned config documents: ${documents.length}`,
    `- Lifecycle passed: ${lifecyclePassed.length}`,
    `- Lifecycle failed: ${lifecycleFailed.length}`,
    "- Formal support: none; local lifecycle proof is not a full release claim.",
    "",
    "| Worker | Runtime and transport | Model or credential source | Config document | Current release blockers |",
    "| --- | --- | --- | --- | --- |",
    ...workers.map(renderWorkerRow),
    "",
    "## Gate Semantics",
    "",
    "- `runtime evidence only`: a local image probe, selectable create option, and",
    "  online Runner report exist. It is not a successful Worker lifecycle.",
    "- `credential reference`: only proves the API projects a reference field;",
    "  it does not prove key injection or provider authentication.",
    "- `config document`: Do Agent, OpenClaw, and Seedance require an explicit",
    "  `{document_id, config_bundle_id}` binding before materialization can be",
    "  tested. Anonymous config bundles are insufficient.",
    "- `lifecycle`: `failed` records an actual attempt that did not satisfy every",
    "  release phase; it is not a support claim. `not_run` has no lifecycle evidence.",
    "",
  ];
  return lines.join("\n");
}

function renderWorkerRow(worker) {
  const runtime = worker.runtime.evidence_state === "runtime_evidence_only"
    ? `${worker.executable}; ${worker.interaction_modes.join("/")}; runtime evidence only`
    : `${worker.executable}; ${worker.interaction_modes.join("/")}; ${worker.runtime.create_option_blocking_reason}`;
  return [
    escapeCell(worker.slug),
    escapeCell(runtime),
    escapeCell(sourceSummary(worker)),
    escapeCell(documentSummary(worker)),
    escapeCell(blockerSummary(worker)),
  ].join(" | ").replace(/^/, "| ").concat(" |");
}

function sourceSummary(worker) {
  if (worker.credential_bindings.length === 0) return "none";
  return worker.credential_bindings.map((binding) => {
    const required = binding.source_kind === "credential_bundle"
      ? binding.required === true ? " (required)" : binding.required === false ? " (optional)" : ""
      : "";
    return `${binding.source_kind}:${binding.source_ref}->${binding.environment_variable}${required}`;
  }).join("<br>");
}

function documentSummary(worker) {
  if (worker.config_documents.length === 0) return "none";
  return worker.config_documents.map((document) =>
    `${document.id}:${document.format}->${document.target_path}`,
  ).join("<br>");
}

function blockerSummary(worker) {
  const gates = worker.integration_gates;
  const blockers = [
    gates.credential,
    gates.config_document_binding,
    gates.rust_core,
    gates.web,
    gates.lifecycle,
  ].filter((value) => value && value !== "not_required");
  return blockers.join("<br>");
}

function escapeCell(value) {
  return String(value).replaceAll("|", "\\|");
}

function requireMatrix(matrix) {
  if (!matrix || !Array.isArray(matrix.workers) || matrix.workers.length === 0) {
    throw new Error("current Worker evidence matrix is missing workers");
  }
  for (const worker of matrix.workers) {
    if (!worker.slug || !worker.runtime || !worker.integration_gates) {
      throw new Error("current Worker evidence matrix has an invalid worker row");
    }
  }
}

function outputPath(argumentsList) {
  if (argumentsList.length === 0) return defaultOutput;
  if (argumentsList.length === 2 && argumentsList[0] === "--output") {
    return path.resolve(argumentsList[1]);
  }
  throw new Error(
    "usage: node render-worker-integration-inventory.mjs [--output <path>]",
  );
}

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

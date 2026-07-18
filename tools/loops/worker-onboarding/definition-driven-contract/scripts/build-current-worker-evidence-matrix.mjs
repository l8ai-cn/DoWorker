import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../../../..");
const loopRoot = path.join(
  root,
  "tools/loops/worker-onboarding/definition-driven-contract",
);
const defaultOutput = path.join(loopRoot, "evidence/current-worker-evidence-matrix.json");
const lifecycleEvidencePath = path.join(
  loopRoot,
  "evidence/real-worker-release.json",
);

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}

function main() {
  const output = outputPath(process.argv.slice(2));
  const catalog = readJson(path.join(root, "config/worker-types/catalog.json"));
  const options = readJson(path.join(loopRoot, "evidence/current-live-create-options.json"));
  const runnerEvidence = readJson(path.join(loopRoot, "evidence/current-runner-agents.json"));
  const optionBySlug = new Map(options.worker_types.map((value) => [value.slug, value]));
  const runnersBySlug = currentRunnersByWorker(runnerEvidence.runners);
  const lifecycleBySlug = new Map(
    readLifecycleEvidence().worker_runs.map((run) => [run.worker_type, run]),
  );
  const workers = catalog.worker_types.map((entry) => {
    const definition = readJson(path.join(root, entry.definition_path));
    const probe = readJson(path.join(
      root,
      "tools/loops/worker-onboarding/catalog-loop/evidence/local-image-probes",
      `${entry.slug}.json`,
    ));
    const option = optionBySlug.get(entry.slug);
    if (!option) throw new Error(`missing live create option for ${entry.slug}`);
    const runners = runnersBySlug.get(entry.slug) ?? [];
    return workerEvidence(entry, definition, probe, option, runners, lifecycleBySlug.get(entry.slug));
  }).sort(compareSlug);
  const document = {
    schema_version: 1,
    formal_support_status: "none",
    options_revision: options.revision,
    worker_count: workers.length,
    workers,
  };
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, JSON.stringify(document, null, 2) + "\n");
  console.log(`wrote current Worker evidence matrix: ${output}`);
}

function workerEvidence(entry, definition, probe, option, runners, lifecycleRun) {
  const runtimeReady = probe.status === "passed" && option.selectable && runners.length > 0;
  const credentialProjection = credentialProjectionState(definition, option);
  const lifecycle = lifecycleState(lifecycleRun);
  const lifecyclePassed = lifecycle === "passed";
  const browserFlow = browserFlowState(lifecycleRun);
  return {
    slug: entry.slug,
    definition_hash: entry.definition_hash,
    executable: definition.executable,
    adapter_id: definition.adapter_id,
    interaction_modes: definition.interaction_modes,
    credential_bindings: definition.credential_bindings.map((binding) =>
      credentialBindingEvidence(binding, option),
    ),
    config_documents: definition.config_documents,
    runtime: {
      image_probe_status: probe.status,
      image_reference: probe.image_reference,
      create_option_selectable: option.selectable,
      create_option_blocking_reason: option.blocking_reason,
      online_runners: runners,
      evidence_state: runtimeReady ? "runtime_evidence_only" : "blocked",
    },
    integration_gates: {
      adapter: definition.interaction_modes.includes("acp")
        ? "runner_registry_contract_passed"
        : "pty_transport_not_applicable",
      credential: lifecyclePassed && definition.credential_bindings.length > 0
        ? "lifecycle_passed"
        : "not_verified",
      credential_reference_api: credentialProjection,
      config_document_binding: lifecyclePassed && definition.config_documents.length > 0
        ? "lifecycle_passed"
        : definition.config_documents.length > 0
        ? "public_contract_pending"
        : "not_required",
      config_document_api: definition.config_documents.length > 0
        ? "not_projected_by_current_wire_contract"
        : "not_required",
      rust_core: browserFlow,
      web: browserFlow,
      lifecycle,
    },
    formal_support_status: lifecyclePassed
      ? "local_lifecycle_verified"
      : "not_verified",
  };
}

function lifecycleState(run) {
  if (!run) return "not_run";
  const passed = [
    run.preflight,
    run.pod_lifecycle,
    run.transport,
    run.browser,
    run.model_execution,
    run.cleanup,
  ].every((value) => value === "passed");
  return passed ? "passed" : "failed";
}

function browserFlowState(run) {
  if (!run) return "browser_not_verified";
  return run.browser === "passed" ? "browser_flow_passed" : "browser_flow_failed";
}

function credentialBindingEvidence(binding, option) {
  const evidence = {
    id: binding.id,
    source_kind: binding.source.kind,
    source_ref: binding.source.ref,
    environment_variable: binding.target.name,
  };
  if (binding.source.kind !== "credential_bundle") return evidence;
  return {
    ...evidence,
    required: option.config_schema?.fields?.[binding.target.name]?.required === true,
  };
}

function credentialProjectionState(definition, option) {
  const fields = option.config_schema?.fields;
  if (fields === null || typeof fields !== "object" || Array.isArray(fields)) {
    throw new Error(`live option ${definition.slug} has no valid config schema fields`);
  }
  const bundleTargets = definition.credential_bindings
    .filter((binding) => binding.source.kind === "credential_bundle")
    .map((binding) => binding.target.name);
  const modelTargets = definition.credential_bindings
    .filter((binding) => binding.source.kind === "model_resource")
    .map((binding) => binding.target.name);
  for (const target of bundleTargets) {
    if (fields[target]?.kind !== "secret") {
      throw new Error(
        `live option ${definition.slug} does not project credential bundle target ${target} as secret`,
      );
    }
  }
  for (const target of modelTargets) {
    if (target in fields) {
      throw new Error(
        `live option ${definition.slug} exposes model resource target ${target} as a user field`,
      );
    }
  }
  if (bundleTargets.length > 0) return "definition_projection_passed";
  if (modelTargets.length > 0) return "model_resource_targets_hidden";
  return "not_required";
}

function currentRunnersByWorker(runners) {
  const results = new Map();
  for (const runner of runners) {
    if (runner.status !== "online") continue;
    for (const slug of runner.available_agents) {
      const version = runner.agent_versions.find((item) => item.slug === slug);
      const item = {
        node_id: runner.node_id,
        version: version?.version ?? "",
        path: version?.path ?? "",
      };
      results.set(slug, [...(results.get(slug) ?? []), item]);
    }
  }
  return results;
}

function outputPath(argumentsList) {
  if (argumentsList.length === 0) return defaultOutput;
  if (argumentsList.length === 2 && argumentsList[0] === "--output") {
    return path.resolve(argumentsList[1]);
  }
  throw new Error("usage: node build-current-worker-evidence-matrix.mjs [--output <path>]");
}

function compareSlug(left, right) {
  return left.slug.localeCompare(right.slug);
}

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

function readLifecycleEvidence() {
  if (!fs.existsSync(lifecycleEvidencePath)) return { worker_runs: [] };
  const evidence = readJson(lifecycleEvidencePath);
  if (!Array.isArray(evidence.worker_runs)) {
    throw new Error("real Worker lifecycle evidence has no worker_runs array");
  }
  return evidence;
}

import path from "node:path";
import {
  hasAvailableRuntime,
  mapRuntimeCatalogEvidence,
} from "./worker-runtime-catalog-evidence.mjs";
import { assertSameSlugs, mapBySlug } from "./worker-loop-json.mjs";

export function buildInventory(context) {
  const {
    definitionCatalog,
    runtimeCatalog,
    evidenceMatrix,
    lockProbes,
    readJson,
    root,
    loopRoot,
    loopRelativeRoot,
  } = context;
  const evidenceBySlug = mapBySlug(evidenceMatrix.workers, "evidence matrix");
  assertSameSlugs(definitionCatalog.worker_types, evidenceMatrix.workers);

  return {
    schema_version: 1,
    workers: definitionCatalog.worker_types.map((entry) => {
      const definition = readJson(path.join(root, entry.definition_path));
      const evidence = evidenceBySlug.get(entry.slug);
      const runtimeEvidencePath = runtimeEvidenceRelativePath(entry.slug);
      const runtimeEvidence = readJson(path.join(loopRoot, runtimeEvidencePath));
      const runtimeCatalogEvidence = mapRuntimeCatalogEvidence(
        runtimeCatalog,
        lockProbes,
        entry.slug,
      );
      return {
        slug: entry.slug,
        source_refs: [
          entry.definition_path,
          `config/worker-types/${entry.slug}/AgentFile`,
          "backend/internal/domain/workerruntime/runtime_catalog.lock.json",
          loopArtifactRepoPath(loopRelativeRoot, "catalog/worker-evidence-matrix.json"),
          loopArtifactRepoPath(loopRelativeRoot, runtimeEvidencePath),
          loopArtifactRepoPath(loopRelativeRoot, "evidence/runtime-lock-probes.json"),
        ],
        layers: [
          "definition",
          "runtime-catalog",
          "runtime-evidence",
          "runner-evidence",
          "product-evidence",
        ],
        definition_hash: entry.definition_hash,
        executable: definition.executable,
        adapter_id: definition.adapter_id,
        interaction_modes: definition.interaction_modes,
        model_requirement: definition.model_requirement,
        credential_bindings: definition.credential_bindings,
        config_documents: definition.config_documents,
        runtime_catalog: runtimeCatalogEvidence,
        runtime_evidence: runtimeEvidence.verdict,
        runner: evidence.runner,
        product_path: evidence.product_path,
        browser: evidence.browser,
        support_status: effectiveSupportStatus(evidence, runtimeCatalogEvidence),
      };
    }),
  };
}

export function buildDrift(context) {
  const { definitionCatalog, runtimeCatalog, evidenceMatrix, lockProbes, readJson, loopRoot } =
    context;
  const runtimeEvidenceBySlug = new Map(
    definitionCatalog.worker_types.map((entry) => [
      entry.slug,
      readJson(path.join(loopRoot, runtimeEvidenceRelativePath(entry.slug))),
    ]),
  );

  return {
    schema_version: 1,
    mismatches: evidenceMatrix.workers
      .filter((worker) => {
        const runtime = mapRuntimeCatalogEvidence(
          runtimeCatalog,
          lockProbes,
          worker.slug,
        );
        return effectiveSupportStatus(worker, runtime) !== "verified_local_dev";
      })
      .map((worker) => {
        const runtimeEvidence = runtimeEvidenceBySlug.get(worker.slug);
        const runtimeCatalogEvidence = mapRuntimeCatalogEvidence(
          runtimeCatalog,
          lockProbes,
          worker.slug,
        );
        return {
          slug: worker.slug,
          layer: hasAvailableRuntime(runtimeCatalogEvidence)
            ? "product-path"
            : "runtime-catalog",
          observed: observedBlocker(worker, runtimeEvidence, runtimeCatalogEvidence),
          target:
            "Definition, immutable runtime, Runner, product-path, and browser evidence must all pass before support.",
          status: "blocked",
          blocker_code: blockerCode(worker, runtimeEvidence, runtimeCatalogEvidence),
          evidence_refs: [
            "catalog/worker-evidence-matrix.json",
            runtimeEvidenceRelativePath(worker.slug),
            "evidence/runtime-lock-probes.json",
          ],
        };
      }),
  };
}

export function mapByWorkerSlug(probes) {
  const result = new Map();
  for (const probe of probes) {
    if (result.has(probe.worker_slug)) {
      throw new Error(`runtime lock probes repeat Worker slug: ${probe.worker_slug}`);
    }
    result.set(probe.worker_slug, probe);
  }
  return result;
}

function observedBlocker(worker, runtimeEvidence, runtimeCatalogEvidence) {
  if (
    hasAvailableRuntime(runtimeCatalogEvidence) &&
    worker.product_path === "blocked_missing_model_resource_verified"
  ) {
    return "A compatible model resource is missing; the Worker wizard blocks creation.";
  }
  if (runtimeEvidence.verdict === "blocked_external_apt_repository") {
    return runtimeEvidence.observed_failure;
  }
  if (runtimeEvidence.verdict === "blocked_missing_real_artifact") {
    return runtimeEvidence.observed_failure;
  }
  if (runtimeCatalogEvidence.status === "invalid_published_digest") {
    return "The published immutable runtime digest could not be pulled from the configured registry.";
  }
  if (!hasAvailableRuntime(runtimeCatalogEvidence)) {
    return "A local runtime smoke result exists, but no published immutable runtime image digest is locked.";
  }
  return "The immutable image is locked, but Runner, product-path, or browser evidence remains incomplete.";
}

function blockerCode(worker, runtimeEvidence, runtimeCatalogEvidence) {
  if (
    hasAvailableRuntime(runtimeCatalogEvidence) &&
    worker.product_path === "blocked_missing_model_resource_verified"
  ) {
    return "missing-model-resource";
  }
  if (runtimeEvidence.verdict === "blocked_external_apt_repository") {
    return "upstream-apt-repository";
  }
  if (runtimeEvidence.verdict === "blocked_missing_real_artifact") {
    return "missing-real-artifact";
  }
  if (runtimeCatalogEvidence.status === "invalid_published_digest") {
    return "invalid-published-runtime-image";
  }
  if (hasAvailableRuntime(runtimeCatalogEvidence)) {
    return "product-path-unverified";
  }
  return "missing-immutable-runtime-image";
}

function effectiveSupportStatus(worker, runtimeCatalogEvidence) {
  return worker.support_status === "verified_local_dev" &&
    !hasAvailableRuntime(runtimeCatalogEvidence)
    ? "not_supported"
    : worker.support_status;
}

function runtimeEvidenceRelativePath(slug) {
  return `evidence/runtime-builds/${slug}.json`;
}

function loopArtifactRepoPath(loopRelativeRoot, relativePath) {
  return `${loopRelativeRoot}/${relativePath}`;
}

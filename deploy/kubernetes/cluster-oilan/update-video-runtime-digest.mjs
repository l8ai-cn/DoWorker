#!/usr/bin/env node

import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
  withinDigestUpdateLock,
  writeRecoverably,
} from "./video-runtime-digest-transaction.mjs";

const IMAGE = "repo.aiedulab.cn:8443/agentcloud/runner-video-studio";
const DIGEST_PATTERN = "sha256:[a-f0-9]{64}";
const DIGEST_RE = new RegExp(`^${DIGEST_PATTERN}$`);
const SCRIPT_DIR = dirname(fileURLToPath(import.meta.url));
const DEFAULT_ROOT = resolve(SCRIPT_DIR, "../../..");
const OBSERVED_AT = process.env.RUNTIME_OBSERVED_AT;

function requireSingleMatch(content, pattern, label) {
  const matches = [...content.matchAll(pattern)];
  if (matches.length !== 1) {
    throw new Error(`${label} must contain exactly one video-studio reference`);
  }
  return matches[0];
}

function loadUpdates(root, digest) {
  if (!DIGEST_RE.test(digest)) {
    throw new Error(`invalid video-studio digest: ${digest}`);
  }

  const paths = {
    lock: join(root, "backend/internal/domain/workerruntime/runtime_catalog.lock.json"),
    backend: join(root, "deploy/kubernetes/cluster-oilan/30-backend.yaml"),
    prepull: join(root, "deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml"),
    build: join(root, "tools/loops/worker-onboarding/catalog-loop/evidence/runtime-builds/video-studio.json"),
    matrix: join(root, "tools/loops/worker-onboarding/catalog-loop/catalog/worker-evidence-matrix.json"),
  };
  const contents = Object.fromEntries(
    Object.entries(paths).map(([key, path]) => [key, readFileSync(path, "utf8")]),
  );

  const catalog = JSON.parse(contents.lock);
  const build = JSON.parse(contents.build);
  const matrix = JSON.parse(contents.matrix);
  const entries = catalog.images?.filter((image) => image.slug === "video-studio-stable") ?? [];
  if (entries.length !== 1) {
    throw new Error("runtime catalog must contain exactly one video-studio-stable image");
  }
  const entry = entries[0];
  const expectedReference = `${IMAGE}@${entry.digest}`;
  if (!DIGEST_RE.test(entry.digest) || entry.reference !== expectedReference) {
    throw new Error("runtime catalog video-studio reference and digest are inconsistent");
  }
  const lockMatch = requireSingleMatch(
    contents.lock,
    /\{[^{}]*"slug": "video-studio-stable"[^{}]*\}/gs,
    "runtime catalog",
  );

  const backendMatch = requireSingleMatch(
    contents.backend,
    new RegExp(`video-studio=${IMAGE}@(${DIGEST_PATTERN})`, "g"),
    "30-backend.yaml",
  );
  const prepullMatch = requireSingleMatch(
    contents.prepull,
    new RegExp(`image: ${IMAGE}@(${DIGEST_PATTERN})`, "g"),
    "60-prepull-daemonset.yaml",
  );
  const currentDigests = new Set([
    entry.digest,
    backendMatch[1],
    prepullMatch[1],
  ]);
  if (currentDigests.size !== 1) {
    throw new Error("video-studio digest references are not synchronized before update");
  }
  if (build.worker_slug !== "video-studio") {
    throw new Error("video-studio runtime evidence has the wrong worker slug");
  }
  const workers = matrix.workers?.filter((worker) => worker.slug === "video-studio") ?? [];
  if (workers.length > 1) {
    throw new Error("worker evidence matrix must not contain duplicate video-studio rows");
  }
  const worker =
    workers[0] ??
    {
      slug: "video-studio",
      support_status: "not_supported",
      definition: "projection_verified_current",
      database_registration: "projection_verified_current",
      runner: "published_runtime_probe_verified_current",
      preflight: "not_run",
      product_path: "not_run",
      browser: "not_run",
      evidence_ref: "evidence/runtime-builds/video-studio.json",
    };
  if (workers.length === 0) {
    matrix.workers ??= [];
    matrix.workers.push(worker);
  }

  const updatedLockEntry = lockMatch[0]
    .replace(`"reference": "${entry.reference}"`, `"reference": "${IMAGE}@${digest}"`)
    .replace(`"digest": "${entry.digest}"`, `"digest": "${digest}"`);
  if (digest !== entry.digest && updatedLockEntry === lockMatch[0]) {
    throw new Error("runtime catalog video-studio entry could not be updated");
  }
  build.validated_at = OBSERVED_AT;
  build.image = `${IMAGE}@${digest}`;
  build.image_id = digest;
  build.checks = {
    local_image_inspect: "passed",
    version_probe: ["video-studio-codex", "--version"],
    video_runtime_contract: "passed",
  };
  build.verdict = "published_runtime_smoke_verified";
  delete build.observed_failure;
  build.limits = [
    "No Runner registration, backend creation request, model credential materialization, or browser flow was executed.",
  ];
  Object.assign(worker, {
    runtime_catalog: "locked_available",
    live_create_option: "published_runtime_available",
    image_target: "published_runtime_smoke_verified",
  });
  return [
    { path: paths.lock, content: contents.lock.replace(lockMatch[0], updatedLockEntry) },
    {
      path: paths.backend,
      content: contents.backend.replace(backendMatch[0], `video-studio=${IMAGE}@${digest}`),
    },
    {
      path: paths.prepull,
      content: contents.prepull.replace(prepullMatch[0], `image: ${IMAGE}@${digest}`),
    },
    { path: paths.build, content: `${JSON.stringify(build, null, 2)}\n` },
    { path: paths.matrix, content: `${JSON.stringify(matrix, null, 2)}\n` },
  ];
}

const digest = process.argv[2];
const root = resolve(process.argv[3] ?? DEFAULT_ROOT);
if (!OBSERVED_AT) {
  throw new Error("RUNTIME_OBSERVED_AT is required");
}
withinDigestUpdateLock(root, () => {
  const updates = loadUpdates(root, digest);
  writeRecoverably(root, updates);
});
console.log(`updated video-studio runtime references to ${digest}`);

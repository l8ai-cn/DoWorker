#!/usr/bin/env node

import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
  withinDigestUpdateLock,
  writeRecoverably,
} from "./video-runtime-digest-transaction.mjs";

const IMAGE = "repo.aiedulab.cn:8443/agentsmesh/runner-video-studio";
const DIGEST_PATTERN = "sha256:[a-f0-9]{64}";
const DIGEST_RE = new RegExp(`^${DIGEST_PATTERN}$`);
const SCRIPT_DIR = dirname(fileURLToPath(import.meta.url));
const DEFAULT_ROOT = resolve(SCRIPT_DIR, "../../..");

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
  };
  const contents = Object.fromEntries(
    Object.entries(paths).map(([key, path]) => [key, readFileSync(path, "utf8")]),
  );

  const catalog = JSON.parse(contents.lock);
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

  const updatedLockEntry = lockMatch[0]
    .replace(`"reference": "${entry.reference}"`, `"reference": "${IMAGE}@${digest}"`)
    .replace(`"digest": "${entry.digest}"`, `"digest": "${digest}"`);
  if (digest !== entry.digest && updatedLockEntry === lockMatch[0]) {
    throw new Error("runtime catalog video-studio entry could not be updated");
  }
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
  ];
}

const digest = process.argv[2];
const root = resolve(process.argv[3] ?? DEFAULT_ROOT);
withinDigestUpdateLock(root, () => {
  const updates = loadUpdates(root, digest);
  writeRecoverably(root, updates);
});
console.log(`updated video-studio runtime references to ${digest}`);

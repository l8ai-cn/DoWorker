#!/usr/bin/env node

import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
  withinDigestUpdateLock,
  writeRecoverably,
} from "./video-runtime-digest-transaction.mjs";

const IMAGE = "repo.aiedulab.cn:8443/agentcloud/runner-do-agent";
const DIGEST_RE = /^sha256:[a-f0-9]{64}$/;
const COMMIT_RE = /^[a-f0-9]{40}$/;
const SCRIPT_DIR = dirname(fileURLToPath(import.meta.url));
const DEFAULT_ROOT = resolve(SCRIPT_DIR, "../../..");

function updateFiles(root, digest, releaseCommit, releaseTag) {
  if (!DIGEST_RE.test(digest)) throw new Error(`invalid do-agent digest: ${digest}`);
  if (!COMMIT_RE.test(releaseCommit)) throw new Error("invalid release source commit");
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(releaseTag)) {
    throw new Error("invalid do-agent release tag");
  }

  const paths = {
    release: join(root, "docker/agent-runtime/do-agent-release.json"),
    catalog: join(root, "backend/internal/domain/workerruntime/runtime_catalog.lock.json"),
    backend: join(root, "deploy/kubernetes/cluster-oilan/30-backend.yaml"),
    doAgentBuild: join(root, "tools/loops/worker-onboarding/catalog-loop/evidence/runtime-builds/do-agent.json"),
    seedanceBuild: join(root, "tools/loops/worker-onboarding/catalog-loop/evidence/runtime-builds/seedance-expert.json"),
  };
  const release = readJson(paths.release);
  const catalog = readJson(paths.catalog);
  const backend = readFileSync(paths.backend, "utf8");
  const builds = [readJson(paths.doAgentBuild), readJson(paths.seedanceBuild)];
  const entry = catalog.images?.filter((image) => image.slug === "do-agent-stable") ?? [];
  if (entry.length !== 1) throw new Error("runtime catalog must contain one do-agent-stable image");
  const currentDigest = entry[0].digest;
  const currentReference = `${IMAGE}@${currentDigest}`;
  if (!DIGEST_RE.test(currentDigest) || entry[0].reference !== currentReference) {
    throw new Error("runtime catalog do-agent reference is inconsistent");
  }
  if (release.image.repository !== IMAGE || release.image.digest !== currentDigest) {
    throw new Error("do-agent release manifest and runtime catalog disagree");
  }
  const matches = [...backend.matchAll(new RegExp(`${IMAGE}@${currentDigest}`, "g"))];
  if (matches.length !== 2) throw new Error("backend must map do-agent and seedance-expert");

  release.image.tag = releaseTag;
  release.image.digest = digest;
  catalog.revision = `runtime-catalog-2026-07-16-${releaseCommit.slice(0, 12)}`;
  entry[0].reference = `${IMAGE}@${digest}`;
  entry[0].digest = digest;
  for (const build of builds) {
    build.validated_at = process.env.RUNTIME_OBSERVED_AT;
    build.image = `${IMAGE}@${digest}`;
    build.image_id = digest;
  }
  return [
    jsonUpdate(paths.release, release),
    jsonUpdate(paths.catalog, catalog),
    {
      path: paths.backend,
      content: backend.replaceAll(currentReference, `${IMAGE}@${digest}`),
    },
    jsonUpdate(paths.doAgentBuild, builds[0]),
    jsonUpdate(paths.seedanceBuild, builds[1]),
  ];
}

function readJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function jsonUpdate(path, value) {
  return { path, content: `${JSON.stringify(value, null, 2)}\n` };
}

const digest = process.argv[2];
const releaseCommit = process.argv[3];
const releaseTag = process.argv[4];
const root = resolve(process.argv[5] ?? DEFAULT_ROOT);
if (!process.env.RUNTIME_OBSERVED_AT) {
  throw new Error("RUNTIME_OBSERVED_AT is required");
}
withinDigestUpdateLock(root, () => {
  writeRecoverably(
    root,
    updateFiles(root, digest, releaseCommit, releaseTag),
    "do-agent-runtime",
  );
}, "do-agent-runtime");
console.log(`updated do-agent runtime references to ${digest}`);

import assert from "node:assert/strict";
import test from "node:test";
import { localWorkerImageReference } from "./probe-local-worker-images.mjs";

test("uses the active Compose project for local Worker images", () => {
  assert.equal(
    localWorkerImageReference("agentsmesh-dev", "codex-cli"),
    "agentsmesh-dev-runner-codex-cli:latest",
  );
});

test("preserves shared runtime image names", () => {
  assert.equal(
    localWorkerImageReference("agentsmesh-dev", "do-agent"),
    "agentsmesh-dev-runner-do-agent:latest",
  );
});

import assert from "node:assert/strict";
import test from "node:test";
import { localWorkerImageReference } from "./probe-local-worker-images.mjs";

test("uses the active Compose project for local Worker images", () => {
  assert.equal(
    localWorkerImageReference("agentcloud-dev", "codex-cli"),
    "agentcloud-dev-runner-codex-cli:latest",
  );
});

test("preserves shared runtime image names", () => {
  assert.equal(
    localWorkerImageReference("agentcloud-dev", "do-agent"),
    "agentcloud-dev-runner-do-agent:latest",
  );
});

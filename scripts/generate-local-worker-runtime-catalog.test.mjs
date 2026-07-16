import assert from "node:assert/strict";
import test from "node:test";
import { buildLocalRuntimeCatalog } from "./generate-local-worker-runtime-catalog.mjs";

const codexDigest = "sha256:e66f3e1990dd7828a9ee8dfc3685a155df55e3ff243a39eaaf6971925c7bee35";
const geminiDigest = "sha256:c24d6da11c46954cd617b21ff33581f9821dc07a8506378c7ac7e305c4ad7cab";
const minimaxDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111";
const openclawDigest = "sha256:2222222222222222222222222222222222222222222222222222222222222222";
const doAgentDigest = "sha256:3333333333333333333333333333333333333333333333333333333333333333";

test("builds an explicit local catalog from every verified local runtime", () => {
  const catalog = buildLocalRuntimeCatalog({
    runtimeImages: [
      ["codex-cli", "agentsmesh-main-runner-codex-cli:latest"],
      ["gemini-cli", "agentsmesh-main-runner-gemini-cli:latest"],
      ["minimax-cli", "agentsmesh-main-runner-minimax-cli:latest"],
      ["openclaw", "agentsmesh-main-runner-openclaw:latest"],
      ["do-agent", "agentsmesh-main-runner-do-agent:latest"],
    ],
    inspectImage: (image) => ({
      "agentsmesh-main-runner-codex-cli:latest": codexDigest,
      "agentsmesh-main-runner-gemini-cli:latest": geminiDigest,
      "agentsmesh-main-runner-minimax-cli:latest": minimaxDigest,
      "agentsmesh-main-runner-openclaw:latest": openclawDigest,
      "agentsmesh-main-runner-do-agent:latest": doAgentDigest,
    })[image],
  });

  assert.equal(catalog.images.length, 5);
  assert.deepEqual(
    catalog.images.map((image) => image.worker_type_slugs),
    [
      ["codex-cli"],
      ["gemini-cli"],
      ["minimax-cli"],
      ["openclaw"],
      ["do-agent", "seedance-expert"],
    ],
  );
  assert.deepEqual(
    catalog.images.map((image) => image.reference),
    [
      `docker-daemon://agentsmesh-main-runner-codex-cli:latest@${codexDigest}`,
      `docker-daemon://agentsmesh-main-runner-gemini-cli:latest@${geminiDigest}`,
      `docker-daemon://agentsmesh-main-runner-minimax-cli:latest@${minimaxDigest}`,
      `docker-daemon://agentsmesh-main-runner-openclaw:latest@${openclawDigest}`,
      `docker-daemon://agentsmesh-main-runner-do-agent:latest@${doAgentDigest}`,
    ],
  );
  assert.ok(catalog.images.every((image) => image.enabled));
});

test("returns no catalog when no requested local runtime has an immutable image ID", () => {
  const catalog = buildLocalRuntimeCatalog({
    runtimeImages: [["codex-cli", "agentsmesh-main-runner-codex-cli:latest"]],
    inspectImage: () => undefined,
  });

  assert.equal(catalog, undefined);
});

import assert from "node:assert/strict";
import test from "node:test";
import { buildLocalRuntimeCatalog } from "./generate-local-worker-runtime-catalog.mjs";

const codexDigest = "sha256:e66f3e1990dd7828a9ee8dfc3685a155df55e3ff243a39eaaf6971925c7bee35";
const videoDigest = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd";
const geminiDigest = "sha256:c24d6da11c46954cd617b21ff33581f9821dc07a8506378c7ac7e305c4ad7cab";
const minimaxDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111";
const openclawDigest = "sha256:2222222222222222222222222222222222222222222222222222222222222222";
const doAgentDigest = "sha256:3333333333333333333333333333333333333333333333333333333333333333";
const e2eEchoDigest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";
const claudeDigest = "sha256:4444444444444444444444444444444444444444444444444444444444444444";
const cursorDigest = "sha256:5555555555555555555555555555555555555555555555555555555555555555";
const loopalDigest = "sha256:6666666666666666666666666666666666666666666666666666666666666666";
const grokDigest = "sha256:7777777777777777777777777777777777777777777777777777777777777777";
const hermesDigest = "sha256:8888888888888888888888888888888888888888888888888888888888888888";
const aiderDigest = "sha256:9999999999999999999999999999999999999999999999999999999999999999";
const opencodeDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";

test("builds an explicit local catalog from every verified local runtime", () => {
  const catalog = buildLocalRuntimeCatalog({
    runtimeImages: [
      ["codex-cli", "agentcloud-main-runner-codex-cli:latest"],
      ["gemini-cli", "agentcloud-main-runner-gemini-cli:latest"],
      ["minimax-cli", "agentcloud-main-runner-minimax-cli:latest"],
      ["openclaw", "agentcloud-main-runner-openclaw:latest"],
      ["do-agent", "agentcloud-main-runner-do-agent:latest"],
      ["e2e-echo", "agentcloud-main-runner-e2e-echo:latest"],
    ],
    inspectImage: (image) => ({
      "agentcloud-main-runner-codex-cli:latest": codexDigest,
      "agentcloud-main-runner-gemini-cli:latest": geminiDigest,
      "agentcloud-main-runner-minimax-cli:latest": minimaxDigest,
      "agentcloud-main-runner-openclaw:latest": openclawDigest,
      "agentcloud-main-runner-do-agent:latest": doAgentDigest,
      "agentcloud-main-runner-e2e-echo:latest": e2eEchoDigest,
    })[image],
  });

  assert.equal(catalog.images.length, 6);
  assert.deepEqual(
    catalog.images.map((image) => image.worker_type_slugs),
    [
      ["codex-cli", "pattern-designer"],
      ["gemini-cli"],
      ["minimax-cli"],
      ["openclaw"],
      ["do-agent", "seedance-expert"],
      ["e2e-echo"],
    ],
  );
  assert.deepEqual(
    catalog.images.map((image) => image.reference),
    [
      `docker-daemon://agentcloud-main-runner-codex-cli:latest@${codexDigest}`,
      `docker-daemon://agentcloud-main-runner-gemini-cli:latest@${geminiDigest}`,
      `docker-daemon://agentcloud-main-runner-minimax-cli:latest@${minimaxDigest}`,
      `docker-daemon://agentcloud-main-runner-openclaw:latest@${openclawDigest}`,
      `docker-daemon://agentcloud-main-runner-do-agent:latest@${doAgentDigest}`,
      `docker-daemon://agentcloud-main-runner-e2e-echo:latest@${e2eEchoDigest}`,
    ],
  );
  assert.ok(catalog.images.every((image) => image.enabled));
});

test("supports every formal Worker runtime in the local catalog", () => {
  const runtimeImages = [
    "aider", "claude-code", "codex-cli", "cursor-cli", "do-agent", "gemini-cli",
    "e2e-echo", "grok-build", "hermes", "loopal", "minimax-cli", "openclaw",
    "opencode", "video-studio",
  ].map((slug) => [slug, `agentcloud-main-runner-${slug}:latest`]);
  const digests = {
    "agentcloud-main-runner-aider:latest": aiderDigest,
    "agentcloud-main-runner-claude-code:latest": claudeDigest,
    "agentcloud-main-runner-codex-cli:latest": codexDigest,
    "agentcloud-main-runner-video-studio:latest": videoDigest,
    "agentcloud-main-runner-cursor-cli:latest": cursorDigest,
    "agentcloud-main-runner-do-agent:latest": doAgentDigest,
    "agentcloud-main-runner-e2e-echo:latest": e2eEchoDigest,
    "agentcloud-main-runner-gemini-cli:latest": geminiDigest,
    "agentcloud-main-runner-grok-build:latest": grokDigest,
    "agentcloud-main-runner-hermes:latest": hermesDigest,
    "agentcloud-main-runner-loopal:latest": loopalDigest,
    "agentcloud-main-runner-minimax-cli:latest": minimaxDigest,
    "agentcloud-main-runner-openclaw:latest": openclawDigest,
    "agentcloud-main-runner-opencode:latest": opencodeDigest,
  };

  const catalog = buildLocalRuntimeCatalog({
    runtimeImages,
    inspectImage: (image) => digests[image],
  });

  assert.deepEqual(
    catalog.images.flatMap((image) => image.worker_type_slugs).sort(),
    [
      "aider", "claude-code", "codex-cli", "cursor-cli", "do-agent", "e2e-echo",
      "gemini-cli", "grok-build", "hermes", "loopal", "minimax-cli", "openclaw",
      "opencode", "pattern-designer", "seedance-expert", "video-studio",
    ],
  );
  assert.match(catalog.revision, /^local-dev-[a-f0-9]{64}$/);
  assert.ok(catalog.revision.length <= 128);
});

test("returns no catalog when no requested local runtime has an immutable image ID", () => {
  const catalog = buildLocalRuntimeCatalog({
    runtimeImages: [["codex-cli", "agentcloud-main-runner-codex-cli:latest"]],
    inspectImage: () => undefined,
  });

  assert.equal(catalog, undefined);
});

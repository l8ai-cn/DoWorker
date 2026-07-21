import path from "node:path";

export const requiredPatternSkills = [
  "pattern-generate",
  "canvas-compose",
  "pattern-seam-review",
  "lovart-api",
];

export const openAICompatibleProviders = [
  "openai",
  "openrouter",
  "dashscope",
  "doubao",
  "deepseek",
  "zhipu",
  "moonshot",
  "xai",
  "mistral",
  "custom-openai-compatible",
];

export function createPatternPreflightConfig() {
  const checkedAt = new Date().toISOString().slice(0, 10);
  return {
    checkedAt,
    webUrl: process.env.PATTERN_WORKER_WEB_URL ?? "http://localhost:12407",
    backendHealthUrl: process.env.PATTERN_WORKER_BACKEND_HEALTH_URL
      ?? "http://localhost:12415/health",
    relayHealthUrl: process.env.PATTERN_WORKER_RELAY_HEALTH_URL
      ?? "http://localhost:12417/health",
    orgSlug: process.env.PATTERN_WORKER_ORG_SLUG ?? "dev-org",
    username: process.env.PATTERN_WORKER_USERNAME,
    password: process.env.PATTERN_WORKER_PASSWORD,
    runnerNodeId: process.env.PATTERN_WORKER_RUNNER_NODE_ID,
    chromiumExecutable: process.env.PATTERN_WORKER_CHROMIUM_EXECUTABLE,
    pgContainer: process.env.PATTERN_WORKER_POSTGRES_CONTAINER
      ?? "agentcloud-codex-loop-blockly-mvp-postgres-1",
    pgUser: process.env.PATTERN_WORKER_POSTGRES_USER ?? "agentcloud",
    pgDatabase: process.env.PATTERN_WORKER_POSTGRES_DATABASE ?? "agentcloud",
    evidenceDir: process.env.PATTERN_WORKER_EVIDENCE_DIR
      ?? path.join(process.cwd(), `deploy/dev/runtime/pattern-worker-evidence-${checkedAt}`),
  };
}

export function createPatternPreflightResult(config) {
  return {
    checked_at: config.checkedAt,
    scenario: "pattern-worker-preflight",
    verdict: "fail_closed",
    config: {
      webUrl: config.webUrl,
      backendHealthUrl: config.backendHealthUrl,
      relayHealthUrl: config.relayHealthUrl,
      orgSlug: config.orgSlug,
      username: config.username ? "<provided>" : "<missing>",
      runnerNodeId: config.runnerNodeId ?? "<missing>",
      chromiumExecutable: config.chromiumExecutable ?? "<playwright-default>",
      evidenceDir: config.evidenceDir,
    },
    probes: {},
    database: {},
    browser: {},
    failures: [],
  };
}

export async function probeHttp(url) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 3000);
  try {
    const response = await fetch(url, { signal: controller.signal });
    return { ok: response.ok, status: response.status };
  } catch (error) {
    return { ok: false, error: error instanceof Error ? error.message : String(error) };
  } finally {
    clearTimeout(timeout);
  }
}

export function requireProbe(result, name, probe) {
  if (!probe.ok) result.failures.push(`${name} health is unavailable`);
}

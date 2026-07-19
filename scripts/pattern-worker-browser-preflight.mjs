import fs from "node:fs";
import path from "node:path";
import {
  createPatternPreflightConfig,
  createPatternPreflightResult,
  probeHttp,
  requireProbe,
} from "./pattern-worker-preflight-config.mjs";
import { checkDatabaseDependencies } from "./pattern-worker-preflight-database.mjs";
import { runBrowserPreflight } from "./pattern-worker-preflight-browser.mjs";

const config = createPatternPreflightConfig();
const result = createPatternPreflightResult(config);

fs.mkdirSync(config.evidenceDir, { recursive: true });

try {
  for (const [name, value] of [
    ["PATTERN_WORKER_USERNAME", config.username],
    ["PATTERN_WORKER_PASSWORD", config.password],
    ["PATTERN_WORKER_RUNNER_NODE_ID", config.runnerNodeId],
  ]) {
    if (!value) result.failures.push(`${name} is required`);
  }

  result.probes.backend = await probeHttp(config.backendHealthUrl);
  result.probes.relay = await probeHttp(config.relayHealthUrl);
  result.probes.web = await probeHttp(config.webUrl);
  requireProbe(result, "backend", result.probes.backend);
  requireProbe(result, "relay", result.probes.relay);
  requireProbe(result, "web frontend", result.probes.web);

  await checkDatabaseDependencies(config, result);
  if (result.failures.length === 0) await runBrowserPreflight(config, result);
  if (result.failures.length === 0) result.verdict = "pass";
} catch (error) {
  result.failures.push(error instanceof Error ? error.message : String(error));
} finally {
  const evidence = path.join(config.evidenceDir, "pattern-worker-preflight.json");
  fs.writeFileSync(evidence, JSON.stringify(result, null, 2) + "\n");
  process.exitCode = result.verdict === "pass" ? 0 : 1;
  console.log(JSON.stringify({
    verdict: result.verdict,
    evidence,
    failures: result.failures,
  }, null, 2));
}

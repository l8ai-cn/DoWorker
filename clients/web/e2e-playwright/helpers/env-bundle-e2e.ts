import { execSync } from "node:child_process";
import { TEST_ORG_SLUG, getComposeProject } from "./env";
import { pollUntil } from "./retry";
import type { ApiFixture } from "../fixtures/api.fixture";
import { pickE2EEchoRunner } from "./e2e-echo-runner";

// Dev compose has TWO runner services (runner-1 + runner-2 — see
// deploy/dev/docker-compose.yml). The pod scheduler picks one via
// least-loaded affinity, so we can't hard-code which container holds the
// dump file. Discover candidates at runtime via `docker ps`; reads happen
// against all of them and the first non-empty wins.
//
// docker-compose name suffix differs per service: `runner-1` has the
// short form, `runner-2` has `runner-2-1` (compose appends an instance
// index when the service name itself contains a digit). Filter by prefix
// and trust docker to list them all.
function listRunnerContainers(): string[] {
  const prefix = `${getComposeProject()}-runner`;
  try {
    const out = execSync(
      `docker ps --filter "name=${prefix}" --format "{{.Names}}"`,
      { encoding: "utf-8" },
    ).trim();
    return out.length ? out.split("\n") : [];
  } catch {
    return [];
  }
}
/**
 * Create an e2e-echo pod through the internal test-agent contract.
 *
 * e2e-echo is intentionally excluded from the public WorkerSpec catalog, so
 * browser worker creation cannot be used to validate Runner environment setup.
 */
export async function createPodAndWaitRunning(args: {
  api: ApiFixture;
  agentSlug: string;
  runtimeBundleNames?: string[];
  statusTimeoutMs?: number;
}): Promise<string> {
  const {
    api,
    agentSlug,
    runtimeBundleNames = [],
    statusTimeoutMs = 30_000,
  } = args;
  const cc = await api.connect();
  const { items: runners } = await cc.runner.listAvailableRunners({
    orgSlug: TEST_ORG_SLUG,
  }) as { items?: Array<{ id: bigint; nodeId?: string }> };
  const created = await cc.pod.createPod({
    orgSlug: TEST_ORG_SLUG,
    runnerId: pickE2EEchoRunner(runners).id,
    agentSlug,
    agentfileLayer: runtimeBundleLayer(runtimeBundleNames),
    automationLevel: "interactive",
  }) as { pod?: { podKey?: string } };
  const podKey = created.pod?.podKey;
  if (!podKey) {
    throw new Error(`createPod missing podKey: ${JSON.stringify(created)}`);
  }

  await pollUntil(
    async () => {
      const pod = await cc.pod.getPod({
        orgSlug: TEST_ORG_SLUG,
        podKey,
      }) as { status?: string };
      return pod.status === "running";
    },
    {
      maxAttempts: Math.ceil(statusTimeoutMs / 1000),
      intervalMs: 1_000,
      label: `pod-${podKey}-running`,
    },
  );
  return podKey;
}

function runtimeBundleLayer(names: string[]): string {
  // e2e-echo writes the child-environment dump only through its PTY launch path.
  return ["MODE pty", ...names.map((name) => `USE_ENV_BUNDLE ${JSON.stringify(name)}`), ""].join("\n");
}

/**
 * Read the env dump file that the e2e-echo agent writes on startup
 * (`/tmp/e2e-echo-env-dump-<pid>`). Polls every runner container until
 * one returns non-empty content or the timeout fires.
 *
 * 60s timeout (was 30s): the full chain is runner.gRPC stream →
 * create_pod RPC → PTY spawn → bash → `echo ready; env > /tmp/dump`,
 * which on a cold self-hosted runner with docker.io pulls + mTLS
 * cert exchange routinely takes 30-45s. PR #410's per-shard backend
 * isolation removed the cross-shard `terminateAllPods` race; what
 * remains is genuine cold-start latency, not a race.
 */
export async function readEnvDumpFromRunner(timeoutMs = 60_000): Promise<string> {
  const deadline = Date.now() + timeoutMs;
  let lastErr: string | undefined;
  const containers = listRunnerContainers();
  if (containers.length === 0) {
    throw new Error(
      `no runner containers found matching ${getComposeProject()}-runner — is the dev environment up?`,
    );
  }
  while (true) {
    for (const container of containers) {
      try {
        const out = execSync(
          `docker exec ${container} sh -c 'cat /tmp/e2e-echo-env-dump-* 2>/dev/null || true'`,
          { encoding: "utf-8" },
        ).trim();
        if (out.length > 0) return out;
      } catch (err) {
        lastErr = (err as Error).message;
      }
    }
    if (Date.now() >= deadline) break;
    await new Promise((resolve) => setTimeout(resolve, 500));
  }
  throw new Error(
    `env dump file did not appear in any of [${containers.join(", ")}] within ${timeoutMs}ms` +
      (lastErr ? ` (last error: ${lastErr})` : ""),
  );
}

/** Wipe any stale dump files from prior runs across every runner container. */
export function clearRunnerDumps(): void {
  for (const container of listRunnerContainers()) {
    try {
      execSync(
        `docker exec ${container} sh -c 'rm -f /tmp/e2e-echo-env-dump-* 2>/dev/null || true'`,
        { encoding: "utf-8" },
      );
    } catch {
      // Container may not be up yet — best effort.
    }
  }
}

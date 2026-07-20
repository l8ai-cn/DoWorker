import { randomUUID } from "node:crypto";
import { E2E_ECHO_AGENT_SLUG } from "./e2e-echo-runner";
import {
  createE2ECleanupSession,
  e2ECleanupError,
  getE2ECleanupPod,
  terminateE2ECleanupPod,
  type E2EPodIdentity,
} from "./e2e-pod-cleanup-transport";
import { TEST_ORG_SLUG } from "./env";

const E2E_RUN_MARKER = `e2e:${randomUUID().slice(0, 12)}`;
const E2E_POD_ALIAS_PATTERN = /^\[e2e:[0-9a-f]{12}\] /;
const TERMINABLE_STATUSES = ["queued", "initializing", "running", "paused", "disconnected"];
const STALE_POD_PAGE_SIZE = 100;
const registeredPods = new Map<string, RegisteredE2EPod>();

interface RegisteredE2EPod {
  orgSlug: string;
  alias: string;
}

export function createE2EPodAlias(alias?: string): string {
  const value = alias?.trim() || "E2E Echo Worker";
  return value.startsWith(`[${E2E_RUN_MARKER}]`)
    ? value
    : `[${E2E_RUN_MARKER}] ${value}`;
}

export function registerE2ECreatedPod(
  podKey: string,
  alias: string,
  orgSlug: string = TEST_ORG_SLUG,
): void {
  if (!podKey.trim()) throw new Error("E2E pod registration requires a pod key");
  if (orgSlug !== TEST_ORG_SLUG) {
    throw new Error(`E2E pod registration is limited to ${TEST_ORG_SLUG}`);
  }
  if (!alias.startsWith(`[${E2E_RUN_MARKER}]`)) {
    throw new Error("E2E pod registration requires the current run marker");
  }
  registeredPods.set(podKey, { orgSlug, alias });
}

export function unregisterE2ECreatedPod(podKey: string): void {
  registeredPods.delete(podKey);
}

export async function terminateRegisteredE2EPods(): Promise<number> {
  const pending = [...registeredPods.entries()];
  if (pending.length === 0) return 0;
  const session = await createE2ECleanupSession("registered");
  let count = 0;
  for (const [podKey, registered] of pending) {
    const pod = await getE2ECleanupPod(session, registered.orgSlug, podKey, "registered");
    if (!matchesRegisteredE2EPod(pod, podKey, registered)) {
      throw e2ECleanupError(
        `refused to terminate registered pod ${podKey}: identity does not match the E2E record`,
      );
    }
    if (!isTerminable(pod)) {
      if (isFinished(pod)) {
        registeredPods.delete(podKey);
        continue;
      }
      throw e2ECleanupError(`registered pod ${podKey} is not in a terminable state`);
    }
    await terminateE2ECleanupPod(session, registered.orgSlug, podKey, "registered");
    registeredPods.delete(podKey);
    count++;
  }
  return count;
}

// CI shards use isolated backends. Shared-dev cleanup is limited to active,
// marker-tagged e2e-echo Pods so unrelated workloads are never selected.
export async function terminateStaleMarkedE2EPods(): Promise<number> {
  const session = await createE2ECleanupSession("stale");
  let count = 0;
  for (const candidate of await listActiveStalePods(session)) {
    if (!isTerminableMarkedE2EPod(candidate) || !candidate.podKey) continue;
    const pod = await getE2ECleanupPod(session, TEST_ORG_SLUG, candidate.podKey, "stale");
    if (!isMarkedE2EPod(pod) || pod.podKey !== candidate.podKey) {
      throw e2ECleanupError(
        `refused to terminate stale pod ${candidate.podKey}: identity changed after listing`,
      );
    }
    if (!isTerminable(pod)) continue;
    await terminateE2ECleanupPod(session, TEST_ORG_SLUG, candidate.podKey, "stale");
    count++;
  }
  return count;
}

async function listActiveStalePods(
  session: Awaited<ReturnType<typeof createE2ECleanupSession>>,
): Promise<E2EPodIdentity[]> {
  const pods: E2EPodIdentity[] = [];
  let offset = 0;
  while (true) {
    const page = await session.listPods(
      TERMINABLE_STATUSES.join(","),
      STALE_POD_PAGE_SIZE,
      offset,
    );
    const items = page.items ?? [];
    const total = page.total === undefined && offset === 0 && items.length === 0
      ? 0
      : Number(page.total);
    if (
      !Number.isSafeInteger(total) ||
      total < 0 ||
      (offset === 0 ? page.offset !== undefined && page.offset !== 0 : page.offset !== offset)
    ) {
      throw e2ECleanupError("stale E2E pod list returned an invalid page");
    }
    pods.push(...items);
    if (offset + items.length >= total) return pods;
    if (items.length === 0) throw e2ECleanupError("stale E2E pod list did not advance");
    offset += items.length;
  }
}

function matchesRegisteredE2EPod(
  pod: E2EPodIdentity,
  podKey: string,
  registered: RegisteredE2EPod,
): boolean {
  return pod.podKey === podKey &&
    pod.alias === registered.alias &&
    pod.agentSlug === E2E_ECHO_AGENT_SLUG;
}

function isTerminableMarkedE2EPod(pod: E2EPodIdentity): boolean {
  return isMarkedE2EPod(pod) && isTerminable(pod);
}

function isMarkedE2EPod(pod: E2EPodIdentity): boolean {
  return pod.agentSlug === E2E_ECHO_AGENT_SLUG &&
    E2E_POD_ALIAS_PATTERN.test(pod.alias ?? "");
}

function isTerminable(pod: E2EPodIdentity): boolean {
  return TERMINABLE_STATUSES.includes(pod.status ?? "");
}

function isFinished(pod: E2EPodIdentity): boolean {
  return ["completed", "terminated", "orphaned", "error"].includes(pod.status ?? "");
}

export function resetRegisteredE2EPodsForTest(): void {
  registeredPods.clear();
}

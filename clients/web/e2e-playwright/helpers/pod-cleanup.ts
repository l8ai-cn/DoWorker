import { randomUUID } from "node:crypto";
import { E2E_ECHO_AGENT_SLUG } from "./e2e-echo-runner";
import { getApiBaseUrl, TEST_ORG_SLUG, TEST_USER } from "./env";

const CONNECT_HEADERS = {
  "Content-Type": "application/json",
  "Connect-Protocol-Version": "1",
};
const E2E_RUN_MARKER = `e2e:${randomUUID().slice(0, 12)}`;
const E2E_POD_ALIAS_PATTERN = /^\[e2e:[0-9a-f]{12}\] /;
const registeredPods = new Map<string, RegisteredE2EPod>();

interface RegisteredE2EPod {
  orgSlug: string;
  alias: string;
}

interface PodIdentity {
  podKey?: string;
  alias?: string;
  agentSlug?: string;
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

export async function terminateRegisteredE2EPods(): Promise<number> {
  const pending = [...registeredPods.entries()];
  if (pending.length === 0) return 0;
  const { baseUrl, headers } = await cleanupSession("registered");
  let count = 0;
  for (const [podKey, registered] of pending) {
    const pod = await getPod(baseUrl, headers, registered.orgSlug, podKey, "registered");
    if (!matchesRegisteredE2EPod(pod, podKey, registered)) {
      throw cleanupError(
        `refused to terminate registered pod ${podKey}: identity does not match the E2E record`,
      );
    }
    await terminatePod(baseUrl, headers, registered.orgSlug, podKey, "registered");
    registeredPods.delete(podKey);
    count++;
  }
  return count;
}

// CI shards use isolated backends. This fallback only reclaims stale Pods in
// a shared dev org when both the e2e-echo agent and the strict run marker match.
export async function terminateStaleMarkedE2EPods(): Promise<number> {
  const { baseUrl, headers } = await cleanupSession("stale");
  const listed = await requestJson<{ items?: PodIdentity[] }>(
    `${baseUrl}/proto.pod.v1.PodService/ListPods`,
    {
      method: "POST",
      headers,
      body: JSON.stringify({ orgSlug: TEST_ORG_SLUG }),
    },
    "list stale E2E pods",
  );
  let count = 0;
  for (const candidate of listed.items ?? []) {
    if (!isMarkedE2EPod(candidate) || !candidate.podKey) continue;
    const pod = await getPod(baseUrl, headers, TEST_ORG_SLUG, candidate.podKey, "stale");
    if (!isMarkedE2EPod(pod) || pod.podKey !== candidate.podKey) {
      throw cleanupError(
        `refused to terminate stale pod ${candidate.podKey}: identity changed after listing`,
      );
    }
    await terminatePod(baseUrl, headers, TEST_ORG_SLUG, candidate.podKey, "stale");
    count++;
  }
  return count;
}

async function cleanupSession(
  scope: "registered" | "stale",
): Promise<{ baseUrl: string; headers: Record<string, string> }> {
  const baseUrl = getApiBaseUrl();
  const login = await requestJson<{ token?: string }>(
    `${baseUrl}/proto.auth.v1.AuthService/Login`,
    {
      method: "POST",
      headers: CONNECT_HEADERS,
      body: JSON.stringify({ username: TEST_USER.username, password: TEST_USER.password }),
    },
    `${scope} cleanup login`,
  );
  if (!login.token) throw cleanupError(`${scope} cleanup login returned no token`);
  return {
    baseUrl,
    headers: { ...CONNECT_HEADERS, Authorization: `Bearer ${login.token}` },
  };
}

async function getPod(
  baseUrl: string,
  headers: Record<string, string>,
  orgSlug: string,
  podKey: string,
  scope: "registered" | "stale",
): Promise<PodIdentity> {
  return requestJson<PodIdentity>(
    `${baseUrl}/proto.pod.v1.PodService/GetPod`,
    {
      method: "POST",
      headers,
      body: JSON.stringify({ orgSlug, podKey }),
    },
    `read ${scope} E2E pod ${podKey}`,
  );
}

async function terminatePod(
  baseUrl: string,
  headers: Record<string, string>,
  orgSlug: string,
  podKey: string,
  scope: "registered" | "stale",
): Promise<void> {
  await requestJson(
    `${baseUrl}/proto.pod.v1.PodService/TerminatePod`,
    {
      method: "POST",
      headers: { ...headers, "X-E2E-Caller": cleanupCaller(scope) },
      body: JSON.stringify({ orgSlug, podKey }),
    },
    `terminate ${scope} E2E pod ${podKey}`,
  );
}

async function requestJson<T>(
  url: string,
  init: RequestInit,
  action: string,
): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, init);
  } catch (cause) {
    throw cleanupError(`${action} request failed: ${errorMessage(cause)}`);
  }
  if (!response.ok) throw cleanupError(`${action} returned HTTP ${response.status}`);
  try {
    return await response.json() as T;
  } catch (cause) {
    throw cleanupError(`${action} returned invalid JSON: ${errorMessage(cause)}`);
  }
}

function matchesRegisteredE2EPod(
  pod: PodIdentity | undefined,
  podKey: string,
  registered: RegisteredE2EPod,
): boolean {
  return pod?.podKey === podKey &&
    pod.alias === registered.alias &&
    pod.agentSlug === E2E_ECHO_AGENT_SLUG;
}

function isMarkedE2EPod(pod: PodIdentity): boolean {
  return pod.agentSlug === E2E_ECHO_AGENT_SLUG &&
    E2E_POD_ALIAS_PATTERN.test(pod.alias ?? "");
}

function cleanupError(message: string): Error {
  return new Error(`E2E pod cleanup: ${message}`);
}

function cleanupCaller(scope: "registered" | "stale"): string {
  return scope === "registered"
    ? "terminateRegisteredE2EPods"
    : "terminateStaleMarkedE2EPods";
}

function errorMessage(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

export function resetRegisteredE2EPodsForTest(): void {
  registeredPods.clear();
}

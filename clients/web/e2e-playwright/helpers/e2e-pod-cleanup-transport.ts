import { getApiBaseUrl, TEST_ORG_SLUG, TEST_USER } from "./env";

const CONNECT_HEADERS = {
  "Content-Type": "application/json",
  "Connect-Protocol-Version": "1",
};

export interface E2EPodIdentity {
  podKey?: string;
  alias?: string;
  agentSlug?: string;
  status?: string;
}

export interface E2EPodListPage {
  items?: E2EPodIdentity[];
  total?: number | string;
  limit?: number;
  offset?: number;
}

export interface E2ECleanupSession {
  listPods(status: string, limit: number, offset: number): Promise<E2EPodListPage>;
  baseUrl: string;
  headers: Record<string, string>;
}

export async function createE2ECleanupSession(
  scope: "registered" | "stale",
): Promise<E2ECleanupSession> {
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
  if (!login.token) throw e2ECleanupError(`${scope} cleanup login returned no token`);
  const headers = { ...CONNECT_HEADERS, Authorization: `Bearer ${login.token}` };
  return {
    baseUrl,
    headers,
    listPods: (status, limit, offset) => requestJson(
      `${baseUrl}/proto.pod.v1.PodService/ListPods`,
      { method: "POST", headers, body: JSON.stringify({ orgSlug: TEST_ORG_SLUG, status, limit, offset }) },
      "list stale E2E pods",
    ),
  };
}

export async function getE2ECleanupPod(
  session: E2ECleanupSession,
  orgSlug: string,
  podKey: string,
  scope: "registered" | "stale",
): Promise<E2EPodIdentity> {
  return requestJson(
    `${session.baseUrl}/proto.pod.v1.PodService/GetPod`,
    { method: "POST", headers: session.headers, body: JSON.stringify({ orgSlug, podKey }) },
    `read ${scope} E2E pod ${podKey}`,
  );
}

export async function terminateE2ECleanupPod(
  session: E2ECleanupSession,
  orgSlug: string,
  podKey: string,
  scope: "registered" | "stale",
): Promise<void> {
  await requestJson(
    `${session.baseUrl}/proto.pod.v1.PodService/TerminatePod`,
    {
      method: "POST",
      headers: { ...session.headers, "X-E2E-Caller": cleanupCaller(scope) },
      body: JSON.stringify({ orgSlug, podKey }),
    },
    `terminate ${scope} E2E pod ${podKey}`,
  );
}

export function e2ECleanupError(message: string): Error {
  return new Error(`E2E pod cleanup: ${message}`);
}

async function requestJson<T>(url: string, init: RequestInit, action: string): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, init);
  } catch (cause) {
    throw e2ECleanupError(`${action} request failed: ${errorMessage(cause)}`);
  }
  if (!response.ok) throw e2ECleanupError(`${action} returned HTTP ${response.status}`);
  try {
    return await response.json() as T;
  } catch (cause) {
    throw e2ECleanupError(`${action} returned invalid JSON: ${errorMessage(cause)}`);
  }
}

function cleanupCaller(scope: "registered" | "stale"): string {
  return scope === "registered"
    ? "terminateRegisteredE2EPods"
    : "terminateStaleMarkedE2EPods";
}

function errorMessage(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

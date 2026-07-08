import { readDoWorkerJWT, readDoWorkerOrgSlug } from "./auth-session";
import { getDoWorkerHostConfig, hostFetch } from "./host-config";

const RESERVED_USER_LOCAL = "local";

export function buildAuthHeaders(
  init: RequestInit | undefined,
  currentUserId: string | null,
): Headers {
  const headers = new Headers(init?.headers);
  const jwt = readDoWorkerJWT();
  if (jwt && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${jwt}`);
  }
  const orgSlug = readDoWorkerOrgSlug();
  if (orgSlug && !headers.has("X-Organization-Slug")) {
    headers.set("X-Organization-Slug", orgSlug);
  }
  if (
    currentUserId &&
    currentUserId !== RESERVED_USER_LOCAL &&
    !headers.has("X-Forwarded-Email")
  ) {
    headers.set("X-Forwarded-Email", currentUserId);
  }
  return headers;
}

export async function authenticatedFetch(
  input: RequestInfo | URL,
  init?: RequestInit,
  currentUserId: string | null = null,
): Promise<Response> {
  const path = typeof input === "string" ? input : input.toString();
  const headers = buildAuthHeaders(init, currentUserId);
  return hostFetch(path, { ...init, headers, cache: "no-store" });
}

export function isEmbeddedHost(): boolean {
  return Boolean(getDoWorkerHostConfig().fetcher);
}

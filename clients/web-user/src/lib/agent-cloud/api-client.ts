import { readAgentCloudJWT, readAgentCloudOrgSlug } from "./auth-session";
import { getAgentCloudHostConfig, hostFetch } from "./host-config";

const RESERVED_USER_LOCAL = "local";

export function buildAuthHeaders(
  init: RequestInit | undefined,
  currentUserId: string | null,
): Headers {
  const headers = new Headers(init?.headers);
  const jwt = readAgentCloudJWT();
  if (jwt && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${jwt}`);
  }
  const orgSlug = readAgentCloudOrgSlug();
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
  return Boolean(getAgentCloudHostConfig().fetcher);
}

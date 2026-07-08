import { apiBaseUrl, resolveApiUrl } from "./api-config";
import { readAuthToken, readOrgSlug } from "./auth-store";

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers);
  const token = readAuthToken();
  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const org = readOrgSlug();
  if (org && !headers.has("X-Organization-Slug")) {
    headers.set("X-Organization-Slug", org);
  }
  const url = apiBaseUrl() ? resolveApiUrl(path) : path;
  return fetch(url, { ...init, headers, cache: "no-store" });
}

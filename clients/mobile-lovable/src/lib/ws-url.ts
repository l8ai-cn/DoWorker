import { readAuthToken, readOrgSlug } from "./auth-store";

export function resolveWebSocketUrl(path: string): string {
  const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
  let url = `${scheme}//${window.location.host}${path}`;
  const params = new URLSearchParams();
  const token = readAuthToken();
  if (token) params.set("token", token);
  const org = readOrgSlug();
  if (org) params.set("org_slug", org);
  const qs = params.toString();
  if (qs) url += (path.includes("?") ? "&" : "?") + qs;
  return url;
}

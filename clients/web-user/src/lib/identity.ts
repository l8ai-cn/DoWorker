/**
 * User identity discovery and session-expiry redirects.
 *
 * API transport lives in `lib/agent-cloud/`; this module owns `/v1/me` caching
 * and login redirect policy.
 */

import {
  authenticatedFetch as doWorkerAuthenticatedFetch,
  getCachedServerInfo,
  hostFetch,
  isEmbeddedHost,
  patchAgentCloudOrgSlug,
  readAgentCloudJWT,
  readAgentCloudOrgSlug,
} from "./agent-cloud";

const RESERVED_USER_LOCAL = "local";

let _currentUserId: string | null = null;
let _currentIsAdmin = false;
let _resolved = false;
let _resolvePromise: Promise<string | null> | null = null;
let _serverLoginUrl: string | null = null;

export function resetIdentity(): void {
  _currentUserId = null;
  _currentIsAdmin = false;
  _resolved = false;
  _resolvePromise = null;
  _serverLoginUrl = null;
}

function reconcileOrgSlug(
  preferred: string | undefined,
  allowed: string[] | undefined,
): void {
  const stored = readAgentCloudOrgSlug();
  const slugs = allowed ?? (preferred ? [preferred] : []);
  if (slugs.length === 0) {
    if (!stored && preferred) patchAgentCloudOrgSlug(preferred);
    return;
  }
  if (stored && slugs.includes(stored)) return;
  const next =
    preferred && slugs.includes(preferred) ? preferred : slugs.find(Boolean) ?? null;
  if (next) patchAgentCloudOrgSlug(next);
}

function _isOnLoginPath(): boolean {
  const path = window.location.pathname;
  return path === "/login" || path === "/register" || path.startsWith("/auth/login");
}

export async function resolveIdentity(): Promise<string | null> {
  if (_resolved) return _currentUserId;
  if (_resolvePromise) return _resolvePromise;
  _resolvePromise = (async () => {
    try {
      const headers = new Headers();
      const jwt = readAgentCloudJWT();
      if (jwt) {
        headers.set("Authorization", `Bearer ${jwt}`);
      }
      const res = await hostFetch("/v1/me", { headers, cache: "no-store" });
      if (res.status === 401) {
        try {
          const data = (await res.json()) as {
            user_id: null;
            login_url?: string;
          };
          if (data.login_url) {
            _serverLoginUrl = data.login_url;
            if (!_isOnLoginPath()) {
              const returnTo = encodeURIComponent(
                window.location.pathname + window.location.search,
              );
              window.location.href = `${data.login_url}?return_to=${returnTo}`;
              return null;
            }
          }
        } catch {
          // Response body was not JSON — fall through.
        }
      }
      if (res.ok) {
        const data = (await res.json()) as {
          user_id: string | null;
          is_admin?: boolean;
          org_slug?: string;
          org_slugs?: string[];
        };
        _currentUserId = data.user_id;
        _currentIsAdmin = data.is_admin ?? false;
        reconcileOrgSlug(data.org_slug, data.org_slugs);
      }
    } catch {
      // Server unreachable — leave as null.
    }
    _resolved = true;
    return _currentUserId;
  })();
  return _resolvePromise;
}

export function getCurrentUserId(): string | null {
  return _currentUserId;
}

export function getCurrentIsAdmin(): boolean {
  return _currentIsAdmin;
}

export function getCurrentAuthorId(): string | null {
  if (_currentUserId === null || _currentUserId === RESERVED_USER_LOCAL) {
    return null;
  }
  return _currentUserId;
}

export async function authenticatedFetch(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  const res = await doWorkerAuthenticatedFetch(input, init, _currentUserId);

  if (
    !isEmbeddedHost() &&
    res.status === 401 &&
    !input.toString().includes("/v1/me") &&
    !input.toString().includes("/auth/") &&
    !_isOnLoginPath()
  ) {
    const loginUrl = getCachedServerInfo()?.login_url ?? _serverLoginUrl;
    if (loginUrl) {
      window.location.href = `${loginUrl}?return_to=${encodeURIComponent(window.location.pathname + window.location.search)}`;
    }
  }
  return res;
}

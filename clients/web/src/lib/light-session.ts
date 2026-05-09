// Pure-TS reader for the PersistedSession blob written by Rust AuthManager.
// Marketing pages (/, /docs, /about, /blog, ...) read auth state through this
// module instead of pulling 21MB of wasm just to render a "Sign In / Console"
// CTA. The schema and url_slug algorithm mirror Rust SSOT at:
//   clients/core/crates/auth/src/state.rs (PersistedSession + url_slug)
//
// MUST NOT import from @/lib/wasm-core, @agentsmesh/service-runtime, or
// agentsmesh-wasm — that would defeat the whole purpose.

export interface LightSession {
  userId: number;
  currentOrgSlug: string | null;
  isAuthenticated: boolean;
  expiresAt: number;
}

const NAMESPACE_PREFIX = "agentsmesh-auth";

// Mirrors Rust state.rs::url_slug — keep in sync. Same algorithm runs in
// e2e-playwright/fixtures/blockstore.fixture.ts (live cross-check).
export function urlSlug(baseUrl: string): string {
  try {
    const u = new URL(baseUrl);
    const port = u.port ? `_${u.port}` : "";
    const raw = `${u.protocol.replace(":", "")}_${u.hostname.toLowerCase()}${port}`;
    return raw.replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64);
  } catch {
    return baseUrl.toLowerCase().replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64);
  }
}

export function sessionStorageKey(baseUrl: string): string {
  return `${NAMESPACE_PREFIX}/${urlSlug(baseUrl)}/session`;
}

interface PersistedSessionWire {
  access_token?: string;
  refresh_token?: string;
  expires_at?: number;
  base_url?: string;
  user_id?: number;
  current_org_slug?: string | null;
  schema_version?: number;
}

export function readLightSession(baseUrl?: string): LightSession | null {
  if (typeof window === "undefined") return null;
  const url = baseUrl ?? window.location.origin;
  const raw = window.localStorage.getItem(sessionStorageKey(url));
  if (!raw) return null;
  try {
    const s = JSON.parse(raw) as PersistedSessionWire;
    const now = Math.floor(Date.now() / 1000);
    return {
      userId: s.user_id ?? 0,
      currentOrgSlug: s.current_org_slug ?? null,
      expiresAt: s.expires_at ?? 0,
      isAuthenticated: !!s.access_token && (s.expires_at ?? 0) > now,
    };
  } catch {
    return null;
  }
}

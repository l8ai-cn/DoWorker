import { resolveApiBaseUrl } from "./env";

const STORAGE_PREFIX = "agent-cloud-auth";
const LEGACY_STORAGE_PREFIX = "agentcloud-auth";
const LOGGED_OUT_KEY = `${STORAGE_PREFIX}/logged-out`;

function sessionStorageKey(): string | null {
  const base = resolveApiBaseUrl();
  try {
    const u = new URL(base);
    const port = u.port ? `_${u.port}` : "";
    const raw = `${u.protocol.replace(":", "")}_${u.hostname.toLowerCase()}${port}`;
    const slug = raw.replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64);
    return `${STORAGE_PREFIX}/${slug}/session`;
  } catch {
    return null;
  }
}

function legacySessionStorageKey(): string | null {
  const base = resolveApiBaseUrl();
  try {
    const u = new URL(base);
    const port = u.port ? `_${u.port}` : "";
    const raw = `${u.protocol.replace(":", "")}_${u.hostname.toLowerCase()}${port}`;
    const slug = raw.replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64);
    return `${LEGACY_STORAGE_PREFIX}/${slug}/session`;
  } catch {
    return null;
  }
}

function readStoredJWT(): string | null {
  const key = sessionStorageKey();
  if (!key) return null;
  const legacyKey = legacySessionStorageKey();
  try {
    let raw = localStorage.getItem(key);
    if (!raw && legacyKey) {
      raw = localStorage.getItem(legacyKey);
      if (raw) {
        localStorage.setItem(key, raw);
        localStorage.removeItem(legacyKey);
      }
    }
    if (!raw) return null;
    const blob = JSON.parse(raw) as { access_token?: string; expires_at?: number };
    if (!blob.access_token) return null;
    // Align expiry semantics with web/Rust SSOT (light-session.ts,
    // auth/state.rs): a session missing expires_at is treated as expired,
    // not perpetually valid.
    if (!blob.expires_at || blob.expires_at * 1000 < Date.now()) return null;
    return blob.access_token;
  } catch {
    return null;
  }
}

export function isSessionLoggedOut(): boolean {
  try {
    return sessionStorage.getItem(LOGGED_OUT_KEY) === "1";
  } catch {
    return false;
  }
}

export function markSessionLoggedOut(): void {
  try {
    sessionStorage.setItem(LOGGED_OUT_KEY, "1");
  } catch {
    // pass
  }
}

export function clearLoggedOutMark(): void {
  try {
    sessionStorage.removeItem(LOGGED_OUT_KEY);
  } catch {
    // pass
  }
}

export function readAgentCloudJWT(): string | null {
  if (isSessionLoggedOut()) return null;
  const envJwt =
    (import.meta.env.VITE_DO_WORKER_JWT as string | undefined) ??
    (import.meta.env.VITE_AGENTCLOUD_JWT as string | undefined);
  if (envJwt) return envJwt;
  return readStoredJWT();
}

export function readAgentCloudOrgSlug(): string | null {
  const key = sessionStorageKey();
  if (!key) return null;
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return null;
    const blob = JSON.parse(raw) as { current_org_slug?: string | null };
    return blob.current_org_slug ?? null;
  } catch {
    return null;
  }
}

export function patchAgentCloudOrgSlug(orgSlug: string | null): void {
  const key = sessionStorageKey();
  if (!key || !orgSlug) return;
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return;
    const blob = JSON.parse(raw) as { current_org_slug?: string | null };
    if (blob.current_org_slug === orgSlug) return;
    blob.current_org_slug = orgSlug;
    localStorage.setItem(key, JSON.stringify(blob));
  } catch {
    // pass
  }
}

export interface PersistSessionInput {
  accessToken: string;
  refreshToken?: string;
  expiresIn: number;
  orgSlug?: string | null;
}

export function persistAgentCloudSession(input: PersistSessionInput): void {
  clearLoggedOutMark();
  const key = sessionStorageKey();
  if (!key) return;
  const legacyKey = legacySessionStorageKey();
  const expiresAt = Math.floor(Date.now() / 1000) + input.expiresIn;
  const blob = {
    access_token: input.accessToken,
    refresh_token: input.refreshToken ?? "",
    expires_at: expiresAt,
    base_url: resolveApiBaseUrl(),
    current_org_slug: input.orgSlug ?? null,
    schema_version: 1,
  };
  try {
    localStorage.setItem(key, JSON.stringify(blob));
    if (legacyKey) localStorage.removeItem(legacyKey);
  } catch {
    // localStorage blocked — JWT injection falls back to env only.
  }
}

export function clearAgentCloudSession(): void {
  const key = sessionStorageKey();
  const legacyKey = legacySessionStorageKey();
  if (!key) return;
  try {
    localStorage.removeItem(key);
    if (legacyKey) localStorage.removeItem(legacyKey);
  } catch {
    // pass
  }
}

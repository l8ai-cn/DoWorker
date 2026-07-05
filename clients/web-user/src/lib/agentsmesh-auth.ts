const STORAGE_PREFIX = "agentsmesh-auth";

function sessionStorageKey(): string | null {
  const base = import.meta.env.VITE_AGENTSMESH_API_URL ?? "http://localhost:10000";
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

export function readAgentsMeshJWT(): string | null {
  if (import.meta.env.VITE_AGENTSMESH_JWT) {
    return import.meta.env.VITE_AGENTSMESH_JWT as string;
  }
  const key = sessionStorageKey();
  if (!key) return null;
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return null;
    const blob = JSON.parse(raw) as { access_token?: string; expires_at?: number };
    if (!blob.access_token) return null;
    if (blob.expires_at && blob.expires_at * 1000 < Date.now()) return null;
    return blob.access_token;
  } catch {
    return null;
  }
}

export function readAgentsMeshOrgSlug(): string | null {
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

export interface PersistSessionInput {
  accessToken: string;
  refreshToken?: string;
  expiresIn: number;
  orgSlug?: string | null;
}

export function persistAgentsMeshSession(input: PersistSessionInput): void {
  const key = sessionStorageKey();
  if (!key) return;
  const expiresAt = Math.floor(Date.now() / 1000) + input.expiresIn;
  const blob = {
    access_token: input.accessToken,
    refresh_token: input.refreshToken ?? "",
    expires_at: expiresAt,
    base_url: import.meta.env.VITE_AGENTSMESH_API_URL ?? "http://localhost:10000",
    current_org_slug: input.orgSlug ?? null,
    schema_version: 1,
  };
  try {
    localStorage.setItem(key, JSON.stringify(blob));
  } catch {
    // localStorage blocked — JWT injection falls back to env only.
  }
}

export function clearAgentsMeshSession(): void {
  const key = sessionStorageKey();
  if (!key) return;
  try {
    localStorage.removeItem(key);
  } catch {
    // pass
  }
}

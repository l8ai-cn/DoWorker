import { getMobileAuthManager, mobileAuthSessionStorageKey } from "./mobile-auth-manager";

export interface AuthSession {
  token: string;
  expiresIn: number;
  orgSlug: string | null;
  userId: string | null;
  email: string | null;
}

export interface AuthIdentity {
  authenticated: boolean;
  email: string | null;
  orgSlug: string | null;
}

interface PersistedAuthSession {
  access_token: string;
  expires_at: number;
  current_org_slug?: string | null;
}

let currentEmail: string | null = null;
const authChangeEvent = "do-worker-mobile-auth-change";

function notifyAuthChange() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(authChangeEvent));
  }
}

export function subscribeAuthChanges(listener: () => void): () => void {
  if (typeof window === "undefined") return () => {};
  window.addEventListener(authChangeEvent, listener);
  return () => window.removeEventListener(authChangeEvent, listener);
}

function readRaw(): PersistedAuthSession | null {
  if (typeof localStorage === "undefined") return null;
  try {
    const raw = localStorage.getItem(mobileAuthSessionStorageKey());
    if (!raw) return null;
    return JSON.parse(raw) as PersistedAuthSession;
  } catch {
    return null;
  }
}

function parseWasmJson<T>(value: unknown): T | null {
  if (typeof value !== "string" || value.length === 0) return null;
  try {
    return JSON.parse(value) as T;
  } catch {
    return null;
  }
}

export function readAuthToken(): string | null {
  const env =
    (import.meta.env.VITE_DO_WORKER_JWT as string | undefined) ??
    (import.meta.env.VITE_AGENTSMESH_JWT as string | undefined);
  if (env) return env;
  const s = readRaw();
  if (!s?.access_token || s.expires_at * 1000 <= Date.now()) return null;
  return s.access_token;
}

export function readOrgSlug(): string | null {
  const env = import.meta.env.VITE_AGENTSMESH_ORG_SLUG as string | undefined;
  if (env) return env;
  return readRaw()?.current_org_slug ?? null;
}

export function readAuthEmail(): string | null {
  return currentEmail;
}

export async function restoreAuthIdentity(): Promise<AuthIdentity> {
  const manager = await getMobileAuthManager();
  const result = parseWasmJson<{
    kind?: string;
    user?: { email?: string };
    current_org?: { slug?: string } | null;
  }>(await manager.bootstrap());
  if (result?.kind !== "authenticated") {
    currentEmail = null;
    notifyAuthChange();
    return { authenticated: false, email: null, orgSlug: null };
  }
  currentEmail = result.user?.email ?? null;
  notifyAuthChange();
  return {
    authenticated: true,
    email: currentEmail,
    orgSlug: result.current_org?.slug ?? null,
  };
}

export async function logout(): Promise<void> {
  const manager = await getMobileAuthManager();
  try {
    await manager.bootstrap();
    await manager.logout();
  } finally {
    manager.clear_session();
    currentEmail = null;
    notifyAuthChange();
  }
}

export async function login(username: string, password: string): Promise<AuthSession> {
  const manager = await getMobileAuthManager();
  const rawSession = await manager.login(username, password);
  try {
    const data = JSON.parse(rawSession) as {
      token: string;
      refresh_token?: string;
      expires_in: number;
      user?: { id?: number; email?: string };
    };
    await manager.fetch_organizations();
    const org = parseWasmJson<{ slug?: string }>(manager.get_current_org_json());
    currentEmail = data.user?.email ?? username;
    notifyAuthChange();
    return {
      token: data.token,
      expiresIn: data.expires_in,
      orgSlug: org?.slug ?? null,
      userId: data.user?.id?.toString() ?? null,
      email: currentEmail,
    };
  } catch (error) {
    manager.clear_session();
    currentEmail = null;
    notifyAuthChange();
    throw error;
  }
}

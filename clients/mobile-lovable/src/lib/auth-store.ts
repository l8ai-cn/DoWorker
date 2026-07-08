const STORAGE_KEY = "agentsmesh-mobile-auth";

export interface AuthSession {
  token: string;
  expiresIn: number;
  orgSlug: string | null;
  userId: string | null;
  email: string | null;
  savedAt: number;
}

function readRaw(): AuthSession | null {
  if (typeof localStorage === "undefined") return null;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as AuthSession;
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
  if (!s?.token) return null;
  if (s.expiresIn > 0 && s.savedAt + s.expiresIn * 1000 < Date.now()) return null;
  return s.token;
}

export function readOrgSlug(): string | null {
  const env = import.meta.env.VITE_AGENTSMESH_ORG_SLUG as string | undefined;
  if (env) return env;
  return readRaw()?.orgSlug ?? null;
}

export function readAuthEmail(): string | null {
  return readRaw()?.email ?? null;
}

export function saveAuthSession(input: Omit<AuthSession, "savedAt">): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({ ...input, savedAt: Date.now() }));
}

export function clearAuthSession(): void {
  localStorage.removeItem(STORAGE_KEY);
}

export async function logout(): Promise<void> {
  const token = readAuthToken();
  try {
    const { apiFetch } = await import("./api-fetch");
    const headers: HeadersInit = {};
    if (token) headers.Authorization = `Bearer ${token}`;
    await apiFetch("/auth/logout", { method: "POST", headers });
  } catch {
    // Network failure — still clear local session.
  }
  clearAuthSession();
}

export async function login(username: string, password: string): Promise<AuthSession> {
  const { apiFetch } = await import("./api-fetch");
  const res = await apiFetch("/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `登录失败 (${res.status})`);
  }
  const data = (await res.json()) as {
    token: string;
    expires_in: number;
    org_slug?: string;
    user?: { id?: string };
  };
  const session: AuthSession = {
    token: data.token,
    expiresIn: data.expires_in,
    orgSlug: data.org_slug ?? null,
    userId: data.user?.id ?? username,
    email: username,
    savedAt: Date.now(),
  };
  saveAuthSession(session);
  return session;
}

export function apiBaseUrl(): string {
  const env =
    (import.meta.env.VITE_DO_WORKER_API_URL as string | undefined) ??
    (import.meta.env.VITE_AGENTSMESH_API_URL as string | undefined);
  return env?.replace(/\/$/, "") ?? "";
}

export function isLiveApiEnabled(): boolean {
  return Boolean(apiBaseUrl() || import.meta.env.DEV);
}

export function resolveApiUrl(path: string): string {
  const base = apiBaseUrl();
  if (!base) return path;
  return `${base}${path.startsWith("/") ? path : `/${path}`}`;
}

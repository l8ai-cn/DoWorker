/** Do Worker API base URL — SSOT for Vite env and auth key derivation. */
export function resolveApiBaseUrl(): string {
  return (
    (import.meta.env.VITE_DO_WORKER_API_URL as string | undefined) ??
    (import.meta.env.VITE_AGENTSMESH_API_URL as string | undefined) ??
    "http://localhost:10000"
  );
}

export function resolveDevProxyTarget(): string {
  return process.env.DO_WORKER_API_URL ?? process.env.AGENTSMESH_API_URL ?? "http://localhost:10000";
}

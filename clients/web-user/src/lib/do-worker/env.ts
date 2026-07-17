/** Do Worker API base URL — SSOT for browser auth key derivation. */
export function resolveApiBaseUrl(): string {
  return (
    import.meta.env.VITE_DO_WORKER_API_URL ??
    import.meta.env.VITE_AGENTSMESH_API_URL ??
    "http://localhost:10000"
  );
}

const DEFAULT_RETURN_TO = "/";

export function sanitizeReturnTo(raw: string | null): string {
  if (raw === null || raw === "") return DEFAULT_RETURN_TO;
  if (!raw.startsWith("/") || raw.startsWith("//") || raw.startsWith("/\\")) {
    return DEFAULT_RETURN_TO;
  }
  try {
    const resolved = new URL(raw, window.location.origin);
    if (resolved.origin !== window.location.origin) return DEFAULT_RETURN_TO;
    return resolved.pathname + resolved.search + resolved.hash;
  } catch {
    return DEFAULT_RETURN_TO;
  }
}

export { DEFAULT_RETURN_TO };

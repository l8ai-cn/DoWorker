// Single chokepoint for post-login redirect targets. Keeps every
// `/login?redirect=...` consumer + every cross-page link aligned on one
// open-redirect policy: only same-origin, path-relative URLs are accepted.
// Anything else (scheme-bearing, protocol-relative, empty, non-string) is
// dropped — the caller falls back to the default route.

export function safeRedirectPath(raw: string | null | undefined): string | null {
  if (typeof raw !== "string" || raw.length === 0) return null;
  if (raw[0] !== "/") return null;
  // Reject `//host` and `/\host` — both normalize to protocol-relative
  // cross-origin URLs in some browsers.
  if (raw.length > 1 && (raw[1] === "/" || raw[1] === "\\")) return null;
  return raw;
}

export function loginUrlWithRedirect(target: string): string {
  const safe = safeRedirectPath(target);
  return safe ? `/login?redirect=${encodeURIComponent(safe)}` : "/login";
}

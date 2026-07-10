function currentOrigin(): string {
  return typeof window === "undefined" ? "" : window.location.origin;
}

function segment(value: string): string {
  return encodeURIComponent(value);
}

function withOrigin(path: string, origin = currentOrigin()): string {
  return origin ? `${origin.replace(/\/$/, "")}${path}` : path;
}

export function buildPodMobileConsoleUrl(
  orgSlug: string,
  podKey: string,
  origin?: string,
): string {
  return withOrigin(`/${segment(orgSlug)}/mobile/pods/${segment(podKey)}`, origin);
}

export function buildPodMobilePreviewUrl(
  orgSlug: string,
  podKey: string,
  origin?: string,
): string {
  return `${buildPodMobileConsoleUrl(orgSlug, podKey, origin)}/preview`;
}

function numericField(record: Record<string, unknown>, snake: string, camel: string): number {
  const raw = record[snake] ?? record[camel];
  return typeof raw === "number" && Number.isFinite(raw) ? raw : 0;
}

export function podHasPreviewAccess(pod: unknown): boolean {
  if (!pod || typeof pod !== "object") return false;
  return numericField(pod as Record<string, unknown>, "preview_port", "previewPort") > 0;
}

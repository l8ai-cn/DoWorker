import type { ActivityType } from "@/stores/ide";

const STANDALONE_ROUTE_SEGMENTS = new Set(["api-access", "automation", "knowledge-base", "workers"]);

export function pathSegments(pathname: string): string[] {
  return pathname.split("/").filter(Boolean);
}

export function pathHasSegment(pathname: string, segment: string): boolean {
  return pathSegments(pathname).includes(segment);
}

export function resolveActivityFromPathname(pathname: string): ActivityType | null {
  if (pathHasSegment(pathname, "workspace")) return "workspace";
  if (pathHasSegment(pathname, "workers")) return "workspace";
  if (pathHasSegment(pathname, "tickets")) return "tickets";
  if (pathHasSegment(pathname, "channels")) return "channels";
  if (pathHasSegment(pathname, "mesh")) return "mesh";
  if (pathHasSegment(pathname, "loops")) return "loops";
  if (pathHasSegment(pathname, "workflows")) return "workflows";
  if (pathHasSegment(pathname, "experts")) return "experts";
  if (pathHasSegment(pathname, "automation")) return "automation";
  if (pathHasSegment(pathname, "api-access")) return "apiAccess";
  if (pathHasSegment(pathname, "knowledge-base")) return "knowledge";
  if (pathHasSegment(pathname, "blocks")) return "blocks";
  if (pathHasSegment(pathname, "infra")) return "infra";
  if (pathHasSegment(pathname, "repositories")) return "repositories";
  if (pathHasSegment(pathname, "runners")) return "runners";
  if (pathHasSegment(pathname, "marketplace")) return "marketplace";
  if (pathHasSegment(pathname, "skills")) return "skills";
  if (pathHasSegment(pathname, "settings")) return "settings";
  return null;
}

export function pathnameHidesIdeSidebar(pathname: string): boolean {
  if (pathname.startsWith("/settings") || pathname.startsWith("/support")) {
    return true;
  }
  return pathSegments(pathname).some((segment) => STANDALONE_ROUTE_SEGMENTS.has(segment));
}

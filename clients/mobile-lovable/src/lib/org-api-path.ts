import { readOrgSlug } from "./auth-store";

export function orgScopedApiPath(subpath: string): string {
  const org = readOrgSlug();
  if (!org) throw new Error("未选择组织");
  const tail = subpath.startsWith("/") ? subpath : `/${subpath}`;
  return `/api/v1/orgs/${encodeURIComponent(org)}${tail}`;
}

import { readCurrentOrg, readOrganizations } from "@/stores/auth";
import { safeRedirectPath } from "@/lib/auth/redirect";

// Desktop-only post-login navigator. Web has an analogous helper at
// clients/web/src/lib/auth/post-login.ts that fetches orgs via
// OrgApiService; here we read straight off the Rust SSOT cache because
// the desktop service shim doesn't expose OrgApiService and the
// auth/login flow already populates that cache via fetch_organizations()
// inside AuthManager.login.
export function navigateAfterLogin(
  push: (url: string) => void,
  redirect: string | null
): void {
  const safe = safeRedirectPath(redirect);
  if (safe) { push(safe); return; }
  const org = readCurrentOrg() ?? readOrganizations()[0];
  push(org ? `/${org.slug}/workspace` : "/onboarding");
}

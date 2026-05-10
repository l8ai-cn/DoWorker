"use client";

import { useEffect } from "react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useCurrentUser, useAuthStore } from "@/stores/auth";
import { CenteredSpinner } from "@/components/ui/spinner";
import { loginUrlWithRedirect } from "@/lib/auth/redirect";

// Single source of truth for the `_hasHydrated && !user → /login` guard.
// Wrap any subtree that requires an authenticated user.
//
// Once hydrated, anonymous visitors are sent to `/login?redirect=<here>` so
// the post-login navigation can land them back where they started.
// The redirect target preserves search + hash so deep links survive.
export function RequireAuth({ children }: { children: React.ReactNode }) {
  const _hasHydrated = useAuthStore((s) => s._hasHydrated);
  const user = useCurrentUser();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (!_hasHydrated || user) return;
    const search = searchParams?.toString();
    const hash = typeof window !== "undefined" ? window.location.hash : "";
    const target = `${pathname || "/"}${search ? `?${search}` : ""}${hash}`;
    router.replace(loginUrlWithRedirect(target));
  }, [_hasHydrated, user, router, pathname, searchParams]);

  if (!_hasHydrated) return <CenteredSpinner />;
  if (!user) return null;
  return <>{children}</>;
}

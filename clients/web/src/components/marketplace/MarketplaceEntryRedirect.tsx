"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { CenteredSpinner } from "@/components/ui/spinner";
import { useLightSession } from "@/hooks/useLightSession";
import { fetchFirstOrgSlug } from "@/lib/light-auth";

export function MarketplaceEntryRedirect({
  acquisition = false,
}: {
  acquisition?: boolean;
}) {
  const { hydrated, session } = useLightSession();
  const router = useRouter();
  const searchParams = useSearchParams();
  const suffix = acquisition ? "/acquire" : "";
  const query = searchParams.toString();

  useEffect(() => {
    if (!hydrated) return;
    const legacyPath = `/marketplace${suffix}${query ? `?${query}` : ""}`;
    if (!session?.isAuthenticated) {
      router.replace(`/login?redirect=${encodeURIComponent(legacyPath)}`);
      return;
    }
    const resolve = session.currentOrgSlug
      ? Promise.resolve(session.currentOrgSlug)
      : fetchFirstOrgSlug();
    resolve.then((orgSlug) => {
      router.replace(orgSlug ? `/${orgSlug}/marketplace${suffix}${query ? `?${query}` : ""}` : "/onboarding");
    });
  }, [hydrated, query, router, session?.currentOrgSlug, session?.isAuthenticated, suffix]);

  return <CenteredSpinner className="h-screen" />;
}

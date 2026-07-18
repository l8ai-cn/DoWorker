"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useLightSession } from "@/hooks/useLightSession";
import { getDefaultRoute } from "@/lib/default-route";

export function HomeSessionRedirect() {
  const router = useRouter();
  const { session, hydrated } = useLightSession();

  useEffect(() => {
    if (!hydrated || !session?.isAuthenticated || !session.currentOrgSlug) return;

    const referrer = document.referrer;
    const isInternalNavigation =
      referrer !== "" && new URL(referrer).origin === window.location.origin;

    if (!isInternalNavigation) {
      router.replace(getDefaultRoute(session.currentOrgSlug));
    }
  }, [hydrated, router, session]);

  return null;
}

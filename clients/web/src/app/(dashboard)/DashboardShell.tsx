"use client";

import React, { useEffect, useCallback, useState } from "react";
import { useRouter } from "next/navigation";
import { useCurrentUser, useCurrentOrg, useAuthStore } from "@/stores/auth";
import { useChannelMessageStore } from "@/stores/channelMessageStore";
import { ResponsiveShell } from "@/components/layout";
import { Spinner } from "@/components/ui/spinner";
import { RealtimeProvider } from "@/providers/RealtimeProvider";
import { initWasmCore, getAuthManager } from "@/lib/wasm-core";
import { useBrowserNotification } from "@/hooks";
import { handleNotificationEvent } from "@/stores/notificationHandler";
import type { RealtimeEvent } from "@/lib/realtime";

export default function DashboardShell({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const [wasmReady, setWasmReady] = useState(false);
  const user = useCurrentUser();
  const currentOrg = useCurrentOrg();
  const { permission, showNotification, requestPermission } = useBrowserNotification();

  useEffect(() => {
    initWasmCore().then(async () => {
      // Bootstrap protocol replaces the previous restore_session + manual
      // hydrate dance. It reads storage, validates the token, refreshes
      // when near expiry, and re-fetches identity / orgs from the server
      // — all atomic. Failure cleans storage and lands the user
      // anonymous (RootRedirect → /login).
      await useAuthStore.getState().bootstrap();

      // Routing helper: if the URL slug names a known org, switch to it.
      // The bootstrap already populated `currentOrg` from the persisted
      // `current_org_slug`, but a deep link to /{otherOrg}/... should
      // win over the persisted preference.
      try {
        const orgs = JSON.parse(getAuthManager().get_organizations_json() || "[]");
        if (orgs.length > 0) {
          const urlSlug = window.location.pathname.split("/")[1];
          const matchedOrg = orgs.find((o: { slug: string }) => o.slug === urlSlug);
          if (matchedOrg) {
            // switch_org updates AuthManager's PersistedSession.current_org_slug;
            // ApiClient reads it via shared AuthTokenStore on every request.
            getAuthManager().switch_org(matchedOrg.slug);
            useAuthStore.getState().setCurrentOrg(matchedOrg);
          }
        }
      } catch { /* noop */ }

      useAuthStore.getState().setHasHydrated(true);
      setWasmReady(true);
    });
  }, []);

  useEffect(() => {
    if (wasmReady && !user) {
      router.push("/login");
    }
  }, [user, router, wasmReady]);

  useEffect(() => {
    if (wasmReady && user && permission === "default") {
      const timer = setTimeout(() => { requestPermission(); }, 3000);
      return () => clearTimeout(timer);
    }
  }, [wasmReady, user, permission, requestPermission]);

  // Cross-cutting org-scoped state (e.g. ActivityBar's channels-unread
  // badge sits outside the org layout's gate) — refresh on org change.
  useEffect(() => {
    if (!wasmReady || !currentOrg) return;
    void useChannelMessageStore.getState().fetchUnreadCounts();
  }, [wasmReady, currentOrg]);

  const handleEvent = useCallback(
    (event: RealtimeEvent) => {
      handleNotificationEvent(event, {
        router,
        showBrowserNotification: (data) => {
          showNotification({
            title: data.title,
            body: data.body,
            tag: `notif-${data.link || data.title}`,
            onClick: () => {
              if (data.link && currentOrg?.slug) {
                router.push(`/${currentOrg.slug}${data.link}`);
              }
            },
          });
        },
      });
    },
    [showNotification, router, currentOrg]
  );

  if (!wasmReady) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <Spinner />
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return (
    <RealtimeProvider onEvent={handleEvent}>
      <ResponsiveShell>{children}</ResponsiveShell>
    </RealtimeProvider>
  );
}

"use client";

import { useEffect } from "react";
import { usePostHog } from "posthog-js/react";
import { useCurrentOrg, useCurrentUser } from "@/stores/auth";

export function PostHogIdentify() {
  const postHog = usePostHog();
  const user = useCurrentUser();
  const currentOrg = useCurrentOrg();

  useEffect(() => {
    if (!postHog) return;
    if (!user) {
      postHog.reset();
      return;
    }
    postHog.identify(String(user.id), {
      email: user.email,
      username: user.username,
      name: user.name,
    });
  }, [postHog, user]);

  useEffect(() => {
    if (!postHog || !currentOrg) return;
    postHog.group("organization", String(currentOrg.id), {
      name: currentOrg.name,
      slug: currentOrg.slug,
      subscription_plan: currentOrg.subscription_plan,
    });
  }, [currentOrg, postHog]);

  return null;
}

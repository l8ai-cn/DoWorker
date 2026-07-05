"use client";

import { useEffect, useState, use } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AuthStatusShell } from "@/components/auth/AuthStatusShell";
import { AuthShell } from "@/components/auth/AuthShell";
import type { InvitationInfo } from "@/lib/api/connect/invitationConnect";
import { ApiError } from "@/lib/api/api-types";
import {
  lightFetchInvitation,
  lightAcceptInvitation,
  lightFetchMe,
} from "@/lib/light-auth";
import { useLightSession } from "@/hooks/useLightSession";
import { useTranslations } from "next-intl";
import { getDefaultRoute } from "@/lib/default-route";
import { InviteAcceptCard } from "./InviteAcceptCard";

export default function InvitePage({ params }: { params: Promise<{ token: string }> }) {
  const resolvedParams = use(params);
  const router = useRouter();
  const t = useTranslations();
  const { session, hydrated } = useLightSession();
  const isSignedIn = !!session?.isAuthenticated;

  const [invitation, setInvitation] = useState<InvitationInfo | null>(null);
  const [meEmail, setMeEmail] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const inv = await lightFetchInvitation(resolvedParams.token);
        if (!cancelled) setInvitation(inv);
      } catch {
        if (!cancelled) setError(t("auth.invitePage.invalidDefault"));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [resolvedParams.token, t]);

  useEffect(() => {
    if (!isSignedIn) return;
    let cancelled = false;
    (async () => {
      const me = await lightFetchMe();
      if (!cancelled) setMeEmail(me?.email ?? null);
    })();
    return () => { cancelled = true; };
  }, [isSignedIn]);

  const handleAccept = async () => {
    if (!invitation) return;
    setAccepting(true);
    setError("");
    try {
      await lightAcceptInvitation(resolvedParams.token, invitation.organizationSlug);
      router.push(getDefaultRoute(invitation.organizationSlug));
    } catch (err: unknown) {
      setError(err instanceof ApiError && err.serverMessage
        ? err.serverMessage
        : t("auth.invitePage.acceptFailed"));
      setAccepting(false);
    }
  };

  if (loading || !hydrated) {
    return (
      <div className="auth-theme min-h-screen flex flex-col items-center justify-center bg-background gap-3">
        <CenteredSpinner />
        <p className="text-sm text-muted-foreground">{t("auth.invitePage.loading")}</p>
      </div>
    );
  }

  if (error && !invitation) {
    return (
      <AuthStatusShell title={t("auth.invitePage.invalidTitle")} subtitle={error} variant="error">
        <Link href="/login">
          <Button className="w-full">{t("auth.invitePage.goToSignIn")}</Button>
        </Link>
      </AuthStatusShell>
    );
  }

  if (invitation?.isExpired) {
    return (
      <AuthStatusShell
        title={t("auth.invitePage.expiredTitle")}
        subtitle={t("auth.invitePage.expiredDescription", { orgName: invitation.organizationName })}
        variant="error"
      >
        <Link href="/login">
          <Button className="w-full">{t("auth.invitePage.goToSignIn")}</Button>
        </Link>
      </AuthStatusShell>
    );
  }

  if (!invitation) return null;

  return (
    <AuthShell
      title={invitation.organizationName}
      subtitle={t("auth.invitePage.invitedBy", {
        inviterName: invitation.inviterName,
        role: invitation.role,
      })}
    >
      <InviteAcceptCard
        invitation={invitation}
        token={resolvedParams.token}
        isSignedIn={isSignedIn}
        meEmail={meEmail}
        accepting={accepting}
        error={error}
        onAccept={handleAccept}
      />
    </AuthShell>
  );
}

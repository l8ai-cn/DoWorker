"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import type { InvitationInfo } from "@/lib/api/connect/invitationConnect";
import { useTranslations } from "next-intl";

interface InviteAcceptCardProps {
  invitation: InvitationInfo;
  token: string;
  isSignedIn: boolean;
  meEmail: string | null;
  accepting: boolean;
  error: string;
  onAccept: () => void;
}

export function InviteAcceptCard({
  invitation,
  token,
  isSignedIn,
  meEmail,
  accepting,
  error,
  onAccept,
}: InviteAcceptCardProps) {
  const t = useTranslations();

  return (
    <div className="space-y-4">
      <div className="flex justify-center">
        <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center">
          <svg className="w-8 h-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
        </div>
      </div>

      {error && (
        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-lg ring-1 ring-destructive/20">{error}</div>
      )}

      {isSignedIn ? (
        <div className="space-y-3">
          {meEmail && (
            <p className="text-sm text-center text-muted-foreground">
              {t("auth.invitePage.signedInAs", { email: meEmail })}
            </p>
          )}
          <Button className="w-full" onClick={onAccept} disabled={accepting}>
            {accepting ? t("auth.invitePage.accepting") : t("auth.invitePage.accept")}
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          <p className="text-sm text-center text-muted-foreground">{t("auth.invitePage.signInPrompt")}</p>
          <Link href={`/login?redirect=/invite/${token}`}>
            <Button className="w-full">{t("auth.invitePage.signInToAccept")}</Button>
          </Link>
          <p className="text-sm text-center text-muted-foreground">
            {t("auth.invitePage.noAccount")}{" "}
            <Link href={`/register?redirect=/invite/${token}`} className="auth-link">
              {t("auth.invitePage.signUp")}
            </Link>
          </p>
        </div>
      )}

      <p className="text-center text-xs text-muted-foreground">
        {t("auth.invitePage.expiresOn", {
          date: new Date(invitation.expiresAt).toLocaleDateString(),
        })}
      </p>
    </div>
  );
}

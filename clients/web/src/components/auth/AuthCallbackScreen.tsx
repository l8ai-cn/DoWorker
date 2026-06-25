"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AuthShell } from "./AuthShell";
import { useTranslations } from "next-intl";

type CallbackStatus = "loading" | "success" | "error";

interface AuthCallbackScreenProps {
  status: CallbackStatus;
  errorMessage?: string;
}

function StatusIcon({ variant }: { variant: "loading" | "success" | "error" }) {
  if (variant === "loading") {
    return (
      <div className="flex justify-center">
        <CenteredSpinner size="lg" />
      </div>
    );
  }

  const isSuccess = variant === "success";
  return (
    <div className="flex justify-center">
      <div
        className={`w-16 h-16 rounded-full flex items-center justify-center ${
          isSuccess ? "bg-success-bg" : "bg-danger-bg"
        }`}
      >
        <svg className={`w-8 h-8 ${isSuccess ? "text-success" : "text-danger"}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
          {isSuccess ? (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          ) : (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          )}
        </svg>
      </div>
    </div>
  );
}

export function AuthCallbackScreen({ status, errorMessage }: AuthCallbackScreenProps) {
  const t = useTranslations();

  if (status === "loading") {
    return (
      <AuthShell title={t("auth.sso.completingSignIn")} subtitle={t("auth.sso.pleaseWait")}>
        <StatusIcon variant="loading" />
      </AuthShell>
    );
  }

  if (status === "success") {
    return (
      <AuthShell title={t("auth.sso.welcome")} subtitle={`${t("auth.sso.signInSuccess")} ${t("auth.sso.redirecting")}`}>
        <StatusIcon variant="success" />
      </AuthShell>
    );
  }

  return (
    <AuthShell title={t("auth.sso.signInFailed")} subtitle={errorMessage ?? t("auth.sso.callbackGenericError")}>
      <div className="space-y-4">
        <StatusIcon variant="error" />
        <div className="space-y-3">
          <Link href="/login">
            <Button className="w-full azure-gradient-bg text-white border-0">{t("auth.sso.tryAgain")}</Button>
          </Link>
          <Link href="/register">
            <Button variant="outline" className="w-full">{t("auth.sso.createAccount")}</Button>
          </Link>
        </div>
      </div>
    </AuthShell>
  );
}

export function resolveOAuthCallbackError(
  reason: string,
  t: (key: string) => string,
): string {
  const messages: Record<string, string> = {
    access_denied: t("auth.sso.callbackAccessDenied"),
    missing_token: t("auth.sso.callbackMissingToken"),
    authentication_failed: t("auth.sso.callbackGenericError"),
    missing_state: t("auth.sso.callbackGenericError"),
    invalid_state: t("auth.sso.callbackGenericError"),
  };
  return messages[reason] ?? t("auth.sso.callbackGenericError");
}

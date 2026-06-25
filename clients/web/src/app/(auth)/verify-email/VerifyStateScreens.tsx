"use client";

import { AuthStatusShell } from "@/components/auth/AuthStatusShell";
import { useTranslations } from "next-intl";

export function VerifyingScreen() {
  const t = useTranslations();
  return (
    <AuthStatusShell
      title={t("auth.verifyEmailPage.verifying")}
      subtitle={t("auth.verifyEmailPage.pleaseWait")}
      variant="loading"
    />
  );
}

export function SuccessScreen() {
  const t = useTranslations();
  return (
    <AuthStatusShell
      title={t("auth.verifyEmailPage.verificationSuccessTitle")}
      subtitle={t("auth.verifyEmailPage.redirecting")}
      variant="success"
    />
  );
}

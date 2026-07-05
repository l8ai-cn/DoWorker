"use client";

import { Suspense, useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AuthShell } from "@/components/auth/AuthShell";
import {
  lightVerifyEmail,
  lightResendVerification,
  resolvePostLoginUrlLight,
} from "@/lib/light-auth";
import { useTranslations } from "next-intl";
import { VerifyingScreen, SuccessScreen } from "./VerifyStateScreens";

type VerifyState = "idle" | "verifying" | "success" | "error";

function VerifyEmailContent() {
  const t = useTranslations();
  const router = useRouter();
  const searchParams = useSearchParams();
  const email = searchParams.get("email") || "";
  const token = searchParams.get("token") || "";

  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [verifyState, setVerifyState] = useState<VerifyState>("idle");

  const handleVerifyToken = useCallback(async (verificationToken: string) => {
    setVerifyState("verifying");
    setError("");
    setMessage("");

    try {
      await lightVerifyEmail(verificationToken);
      setVerifyState("success");
      setMessage(t("auth.verifyEmailPage.verificationSuccess"));
      const url = await resolvePostLoginUrlLight({ redirectParam: null });
      router.push(url);
    } catch (err) {
      setVerifyState("error");
      const errorMessage = err instanceof Error ? err.message : String(err);
      if (errorMessage.includes("already verified")) {
        setError(t("auth.verifyEmailPage.alreadyVerifiedError"));
      } else if (errorMessage.includes("expired") || errorMessage.includes("invalid")) {
        setError(t("auth.verifyEmailPage.invalidToken"));
      } else {
        setError(t("auth.verifyEmailPage.verificationFailed"));
      }
    }
  }, [router, t]);

  useEffect(() => {
    if (token && verifyState === "idle") {
      handleVerifyToken(token);
    }
  }, [token, verifyState, handleVerifyToken]);

  const handleResend = async () => {
    if (!email) {
      setError(t("auth.verifyEmailPage.emailMissing"));
      return;
    }
    setLoading(true);
    setError("");
    setMessage("");
    try {
      await lightResendVerification(email);
      setMessage(t("auth.verifyEmailPage.emailSent"));
    } catch {
      setError(t("auth.verifyEmailPage.resendFailed"));
    } finally {
      setLoading(false);
    }
  };

  if (verifyState === "verifying") return <VerifyingScreen />;
  if (verifyState === "success") return <SuccessScreen />;

  return (
    <AuthShell
      title={t("auth.verifyEmailPage.title")}
      subtitle={
        email
          ? `${t("auth.verifyEmailPage.subtitle", { email })} ${t("auth.verifyEmailPage.clickLink")}`
          : `${t("auth.verifyEmailPage.subtitleDefault")} ${t("auth.verifyEmailPage.clickLink")}`
      }
      footer={
        <p className="text-sm text-muted-foreground">
          {t("auth.verifyEmailPage.alreadyVerified")}{" "}
          <Link href="/login" className="auth-link">
            {t("auth.verifyEmailPage.signIn")}
          </Link>
        </p>
      }
    >
      <div className="flex justify-center mb-2">
        <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center">
          <svg className="w-8 h-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
        </div>
      </div>

      {message && (
        <div className="p-3 text-sm text-success bg-success-bg rounded-lg">{message}</div>
      )}
      {error && (
        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-lg ring-1 ring-destructive/20">{error}</div>
      )}

      <div className="space-y-3">
        <Button variant="outline" className="w-full" onClick={handleResend} disabled={loading || !email}>
          {loading ? t("auth.verifyEmailPage.sending") : t("auth.verifyEmailPage.resendEmail")}
        </Button>
        <p className="text-sm text-center text-muted-foreground">
          {t("auth.verifyEmailPage.wrongEmail")}{" "}
          <Link href="/register" className="auth-link">
            {t("auth.verifyEmailPage.signUpDifferent")}
          </Link>
        </p>
      </div>
    </AuthShell>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense
      fallback={
        <div className="auth-theme min-h-screen flex items-center justify-center bg-background">
          <CenteredSpinner />
        </div>
      }
    >
      <VerifyEmailContent />
    </Suspense>
  );
}

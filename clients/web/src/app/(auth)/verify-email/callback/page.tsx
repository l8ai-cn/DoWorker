"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AuthStatusShell } from "@/components/auth/AuthStatusShell";
import { lightVerifyEmail } from "@/lib/light-auth";
import { useTranslations } from "next-intl";

function VerifyEmailCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const t = useTranslations();

  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [error, setError] = useState("");

  useEffect(() => {
    const verifyEmail = async () => {
      if (!token) {
        setStatus("error");
        setError(t("auth.verifyEmailCallbackPage.tokenMissing"));
        return;
      }

      try {
        await lightVerifyEmail(token);
        setStatus("success");
        setTimeout(() => router.push("/onboarding"), 2000);
      } catch (err: unknown) {
        setStatus("error");
        if (err instanceof Error && err.message.includes("already verified")) {
          setError(t("auth.verifyEmailPage.alreadyVerifiedError"));
        } else {
          setError(t("auth.verifyEmailPage.invalidToken"));
        }
      }
    };

    verifyEmail();
  }, [token, router, t]);

  if (status === "loading") {
    return (
      <AuthStatusShell
        title={t("auth.verifyEmailPage.verifying")}
        subtitle={t("auth.verifyEmailPage.pleaseWait")}
        variant="loading"
      />
    );
  }

  if (status === "success") {
    return (
      <AuthStatusShell
        title={t("auth.verifyEmailCallbackPage.successTitle")}
        subtitle={t("auth.verifyEmailCallbackPage.successDescription")}
        variant="success"
      />
    );
  }

  return (
    <AuthStatusShell
      title={t("auth.verifyEmailCallbackPage.failedTitle")}
      subtitle={error}
      variant="error"
    >
      <div className="space-y-3">
        <Link href="/login">
          <Button className="w-full azure-gradient-bg text-white border-0">
            {t("auth.verifyEmailCallbackPage.signIn")}
          </Button>
        </Link>
        <Link href="/register">
          <Button variant="outline" className="w-full">
            {t("auth.verifyEmailCallbackPage.signUpAgain")}
          </Button>
        </Link>
      </div>
    </AuthStatusShell>
  );
}

export default function VerifyEmailCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="azure-theme min-h-screen flex items-center justify-center bg-background">
          <CenteredSpinner />
        </div>
      }
    >
      <VerifyEmailCallbackContent />
    </Suspense>
  );
}

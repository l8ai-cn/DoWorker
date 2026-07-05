"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { CenteredSpinner } from "@/components/ui/spinner";
import { useOAuthCallback } from "@/hooks/useOAuthCallback";
import { AuthCallbackScreen, resolveOAuthCallbackError } from "@/components/auth/AuthCallbackScreen";
import { useTranslations } from "next-intl";

function SSOCallbackContent() {
  const searchParams = useSearchParams();
  const t = useTranslations();
  const { status, errorReason } = useOAuthCallback(searchParams);

  return (
    <AuthCallbackScreen
      status={status}
      errorMessage={status === "error" ? resolveOAuthCallbackError(errorReason, t) : undefined}
    />
  );
}

export default function SSOCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="auth-theme min-h-screen flex items-center justify-center bg-background">
          <CenteredSpinner />
        </div>
      }
    >
      <SSOCallbackContent />
    </Suspense>
  );
}

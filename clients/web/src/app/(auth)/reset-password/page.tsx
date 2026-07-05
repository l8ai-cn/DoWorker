"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AuthShell } from "@/components/auth/AuthShell";
import { lightResetPassword } from "@/lib/light-auth";
import { useRedirectIfAuthenticated } from "@/hooks/useRedirectIfAuthenticated";
import { useTranslations } from "next-intl";

function ResetPasswordContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const t = useTranslations();
  useRedirectIfAuthenticated();

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!token) {
      setError(t("auth.resetPasswordPage.tokenMissing"));
      return;
    }
    if (password !== confirmPassword) {
      setError(t("auth.resetPasswordPage.passwordMismatch"));
      return;
    }
    if (password.length < 8) {
      setError(t("auth.resetPasswordPage.passwordTooShort"));
      return;
    }

    setLoading(true);
    setError("");

    try {
      await lightResetPassword({ token, newPassword: password });
      setSuccess(true);
      setTimeout(() => router.push("/login"), 2000);
    } catch {
      setError(t("auth.resetPasswordPage.resetFailed"));
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <AuthShell
        title={t("auth.resetPasswordPage.successTitle")}
        subtitle={t("auth.resetPasswordPage.successDescription")}
      >
        <div className="flex justify-center">
          <div className="w-16 h-16 rounded-full bg-success-bg flex items-center justify-center">
            <svg className="w-8 h-8 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </div>
      </AuthShell>
    );
  }

  if (!token) {
    return (
      <AuthShell
        title={t("auth.resetPasswordPage.invalidLinkTitle")}
        subtitle={t("auth.resetPasswordPage.invalidLinkDescription")}
      >
        <Link href="/forgot-password">
          <Button className="w-full">
            {t("auth.resetPasswordPage.requestNewLink")}
          </Button>
        </Link>
      </AuthShell>
    );
  }

  return (
    <AuthShell
      title={t("auth.resetPasswordPage.title")}
      subtitle={t("auth.resetPasswordPage.subtitle")}
      footer={
        <p className="text-center text-sm text-muted-foreground">
          {t("auth.resetPasswordPage.rememberPassword")}{" "}
          <Link href="/login" className="auth-link">
            {t("auth.resetPasswordPage.signIn")}
          </Link>
        </p>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-lg ring-1 ring-destructive/20">
            {error}
          </div>
        )}

        <div className="space-y-2">
          <label htmlFor="password" className="text-sm font-medium text-foreground">
            {t("auth.resetPasswordPage.newPasswordLabel")}
          </label>
          <Input
            id="password"
            type="password"
            placeholder={t("auth.resetPasswordPage.newPasswordPlaceholder")}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="confirmPassword" className="text-sm font-medium text-foreground">
            {t("auth.resetPasswordPage.confirmPasswordLabel")}
          </label>
          <Input
            id="confirmPassword"
            type="password"
            placeholder={t("auth.resetPasswordPage.confirmPasswordPlaceholder")}
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
          />
        </div>

        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? t("auth.resetPasswordPage.submitting") : t("auth.resetPasswordPage.submit")}
        </Button>
      </form>
    </AuthShell>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense
      fallback={
        <div className="auth-theme min-h-screen flex items-center justify-center bg-background">
          <CenteredSpinner />
        </div>
      }
    >
      <ResetPasswordContent />
    </Suspense>
  );
}

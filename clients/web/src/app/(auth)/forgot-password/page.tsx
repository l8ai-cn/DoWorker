"use client";

import { useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AuthShell } from "@/components/auth/AuthShell";
import { AuthStatusShell } from "@/components/auth/AuthStatusShell";
import { lightForgotPassword } from "@/lib/light-auth";
import { useRedirectIfAuthenticated } from "@/hooks/useRedirectIfAuthenticated";
import { useTranslations } from "next-intl";

export default function ForgotPasswordPage() {
  const t = useTranslations();
  useRedirectIfAuthenticated();
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await lightForgotPassword(email);
      setSubmitted(true);
    } catch {
      setError(t("auth.forgotPasswordPage.sendFailed"));
    } finally {
      setLoading(false);
    }
  };

  if (submitted) {
    return (
      <AuthStatusShell
        title={t("auth.forgotPasswordPage.checkEmail")}
        subtitle={t("auth.forgotPasswordPage.emailSentDescription", { email })}
        variant="success"
      >
        <Link href="/login">
          <Button variant="outline" className="w-full">
            {t("auth.forgotPasswordPage.backToSignIn")}
          </Button>
        </Link>
      </AuthStatusShell>
    );
  }

  return (
    <AuthShell
      title={t("auth.forgotPasswordPage.title")}
      subtitle={t("auth.forgotPasswordPage.subtitle")}
      footer={
        <p className="text-center text-sm text-muted-foreground">
          {t("auth.forgotPasswordPage.rememberPassword")}{" "}
          <Link href="/login" className="auth-link">
            {t("auth.forgotPasswordPage.signIn")}
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
          <label htmlFor="email" className="text-sm font-medium text-foreground">
            {t("auth.forgotPasswordPage.emailLabel")}
          </label>
          <Input
            id="email"
            type="email"
            placeholder={t("auth.forgotPasswordPage.emailPlaceholder")}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>

        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? t("auth.forgotPasswordPage.sending") : t("auth.forgotPasswordPage.sendResetLink")}
        </Button>
      </form>
    </AuthShell>
  );
}

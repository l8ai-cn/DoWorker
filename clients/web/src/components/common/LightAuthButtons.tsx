"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useLightSession } from "@/hooks/useLightSession";
import { discoverFirstOrgSlug } from "@/lib/light-auth";
import { updateLightSessionOrgSlug } from "@/lib/light-session";
import { useTranslations } from "next-intl";

interface LightAuthButtonsProps {
  size?: "sm" | "default";
  consoleVariant?: "primary" | "outline";
  showRegister?: boolean;
  onClick?: () => void;
  className?: string;
}

// Auth-aware CTA for marketing pages — reads PersistedSession from localStorage
// directly instead of going through wasm. Interface matches AuthButtons so
// PageHeader / Navbar can swap imports without changing call sites.
// Use this on routes that must stay wasm-free (/, /docs, /about, /blog, ...).
// Use AuthButtons inside (auth) / (dashboard) where wasm is already loaded.
export function LightAuthButtons({
  size = "default",
  consoleVariant = "primary",
  showRegister = false,
  onClick,
  className,
}: LightAuthButtonsProps) {
  const router = useRouter();
  const [consoleError, setConsoleError] = useState<string | null>(null);
  const { session, hydrated } = useLightSession();
  const t = useTranslations();

  if (!hydrated) return null;

  const isLoggedIn = !!session?.isAuthenticated;

  async function openConsole() {
    setConsoleError(null);
    let orgSlug = session?.currentOrgSlug;
    if (!orgSlug) {
      const result = await discoverFirstOrgSlug();
      if (result.status === "unavailable") {
        setConsoleError(t("landing.nav.consoleUnavailable"));
        return;
      }
      if (result.status === "empty") {
        onClick?.();
        router.push("/onboarding/create-org");
        return;
      }
      orgSlug = result.slug;
    }

    onClick?.();
    updateLightSessionOrgSlug(orgSlug);
    router.push(`/${orgSlug}/workspace`);
  }

  if (isLoggedIn) {
    return (
      <div className={`relative ${className ?? ""}`}>
        <Button
          type="button"
          size={size}
          variant={consoleVariant === "outline" ? "outline" : "default"}
          onClick={openConsole}
          className={
            consoleVariant === "primary"
              ? "bg-primary text-primary-foreground hover:bg-primary/90"
              : undefined
          }
        >
          {t("landing.nav.console")}
        </Button>
        {consoleError && (
          <p
            role="alert"
            className="absolute right-0 top-full z-[60] mt-2 w-72 max-w-[calc(100vw-2rem)] rounded-md border border-destructive/30 bg-background p-3 text-xs leading-5 text-foreground shadow-lg"
          >
            {consoleError}
          </p>
        )}
      </div>
    );
  }

  return (
    <div className={className}>
      <Link href="/login" onClick={onClick}>
        <Button variant={showRegister ? "ghost" : "outline"} size={size}>
          {t("landing.nav.signIn")}
        </Button>
      </Link>
      {showRegister && (
        <Link href="/register" onClick={onClick}>
          <Button
            size={size}
            className="bg-primary text-primary-foreground hover:bg-primary/90"
          >
            {t("landing.nav.getStarted")}
          </Button>
        </Link>
      )}
    </div>
  );
}

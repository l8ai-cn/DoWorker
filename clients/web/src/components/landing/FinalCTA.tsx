"use client";

import Link from "next/link";
import { ArrowRight, Store } from "lucide-react";
import { useTranslations } from "next-intl";

export function FinalCTA() {
  const t = useTranslations();

  return (
    <section className="border-y border-black/10 bg-[var(--expert-action)] py-20 text-[var(--expert-ink)] sm:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <p className="text-xs font-bold uppercase">
          {t("landing.workforce.expertHome.cta.eyebrow")}
        </p>
        <div className="mt-4 grid gap-8 lg:grid-cols-[1fr_auto] lg:items-end">
          <div>
            <h2 className="max-w-4xl text-3xl font-semibold leading-tight sm:text-5xl">
              {t("landing.workforce.expertHome.cta.title")}
            </h2>
            <p className="mt-5 max-w-3xl text-base leading-7 text-[var(--expert-ink)]/75">
              {t("landing.workforce.expertHome.cta.description")}
            </p>
          </div>
          <div className="flex flex-col gap-3 sm:flex-row lg:flex-col">
            <Link href="/register" className="inline-flex min-h-12 items-center justify-center gap-2 rounded-md bg-[var(--expert-ink)] px-5 text-sm font-semibold text-white">
              {t("landing.workforce.expertHome.cta.primary")}
              <ArrowRight className="h-4 w-4" />
            </Link>
            <Link href="/marketplace" className="inline-flex min-h-12 items-center justify-center gap-2 rounded-md border border-[var(--expert-ink)]/30 px-5 text-sm font-semibold transition-colors hover:bg-white/20">
              <Store className="h-4 w-4" />
              {t("landing.workforce.expertHome.cta.secondary")}
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}

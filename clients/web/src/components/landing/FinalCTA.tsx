"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";

export function FinalCTA() {
  const t = useTranslations();

  return (
    <section className="relative overflow-hidden bg-[var(--azure-bg-deeper)] py-28 sm:py-32">
      <div className="workforce-grid pointer-events-none absolute inset-0 opacity-20" aria-hidden="true" />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute bottom-0 left-1/2 h-[420px] w-[900px] -translate-x-1/2 rounded-full bg-[var(--azure-mint)]/12 blur-[140px]"
      />

      <div className="relative z-10 mx-auto max-w-3xl px-4 text-center sm:px-6 lg:px-8">
        <h2 className="font-headline text-4xl font-bold leading-[1.05] md:text-6xl">
          {t("landing.finalCta.title1")}
          <span className="mt-2 block bg-gradient-to-r from-[var(--azure-mint)] to-[var(--azure-cyan-soft)] bg-clip-text text-transparent">
            {t("landing.finalCta.title2")}
          </span>
        </h2>

        <p className="mt-6 text-lg font-light text-[var(--azure-text-muted)]">
          {t("landing.finalCta.description")}
        </p>

        <div className="mt-10 flex flex-col justify-center gap-4 sm:flex-row">
          <a href="#mission" className="w-full sm:w-auto">
            <button className="w-full min-h-12 rounded-full bg-[var(--azure-mint)] px-10 py-4 font-headline text-xs font-black uppercase tracking-[0.18em] text-[var(--azure-on-cyan)] shadow-[0_0_32px_rgba(20,184,166,0.3)] transition-transform motion-safe:hover:-translate-y-0.5 sm:w-auto sm:text-sm">
              {t("landing.finalCta.watchDemo")}
            </button>
          </a>
          <Link href="/register" className="w-full sm:w-auto">
            <button className="w-full min-h-12 rounded-full border border-[var(--azure-outline-variant)] px-10 py-4 font-headline text-xs font-bold uppercase tracking-[0.18em] text-foreground transition-colors hover:border-[var(--azure-mint)]/60 sm:w-auto sm:text-sm">
              {t("landing.finalCta.getStartedFree")}
            </button>
          </Link>
        </div>

        <p className="mt-6 text-sm text-[var(--azure-text-muted)]/70">
          {t("landing.finalCta.enterpriseNote")}{" "}
          <Link href="/demo" className="underline underline-offset-4 hover:text-[var(--azure-mint)] transition-colors">
            {t("landing.finalCta.contactUs")}
          </Link>
        </p>

        <div className="mt-12 flex flex-wrap justify-center gap-x-8 gap-y-3 text-sm text-[var(--azure-text-muted)]">
          <div className="flex items-center gap-2">
            <span className="h-1.5 w-1.5 rounded-full bg-[var(--azure-mint)]" />
            {t("landing.finalCta.freeTier")}
          </div>
          <div className="flex items-center gap-2">
            <span className="h-1.5 w-1.5 rounded-full bg-[var(--azure-mint)]" />
            {t("landing.finalCta.noCreditCard")}
          </div>
          <div className="flex items-center gap-2">
            <span className="h-1.5 w-1.5 rounded-full bg-[var(--azure-mint)]" />
            {t("landing.finalCta.selfHostedOption")}
          </div>
        </div>
      </div>
    </section>
  );
}

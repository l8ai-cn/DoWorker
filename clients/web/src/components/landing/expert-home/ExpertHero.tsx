"use client";

import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { useTranslations } from "next-intl";

import { ExpertControlSurface } from "./ExpertControlSurface";

export function ExpertHero() {
  const t = useTranslations();

  return (
    <section id="expert" className="relative scroll-mt-24 border-b border-white/8 bg-[var(--expert-bg)] pb-16 pt-28 sm:pb-20 sm:pt-36">
      <div className="pointer-events-none absolute inset-y-0 left-[8%] border-l border-white/[0.035]" aria-hidden="true" />
      <div className="pointer-events-none absolute inset-y-0 right-[8%] border-r border-white/[0.035]" aria-hidden="true" />
      <div className="relative mx-auto grid max-w-7xl items-center gap-12 px-4 sm:px-6 lg:px-8 xl:grid-cols-[1fr_1.1fr]">
        <div className="max-w-2xl">
          <p className="expert-eyebrow">
            {t("landing.workforce.expertHome.hero.eyebrow")}
          </p>
          <h1 className="mt-5 text-4xl font-semibold leading-[1.08] text-white sm:text-5xl">
            {t("landing.workforce.expertHome.hero.title")}
          </h1>
          <p className="mt-6 max-w-xl text-base leading-8 text-[var(--expert-muted)] sm:text-lg">
            {t("landing.workforce.expertHome.hero.description")}
          </p>
          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Link href="/marketplace" className="expert-primary-button">
              {t("landing.workforce.expertHome.hero.primaryAction")}
              <ArrowRight className="h-4 w-4" />
            </Link>
            <Link href="/product" className="expert-secondary-button">
              {t("landing.workforce.expertHome.hero.secondaryAction")}
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
          <p className="mt-7 border-l-2 border-[var(--expert-warning)] pl-4 text-sm leading-6 text-[var(--expert-text)]">
            {t("landing.workforce.expertHome.hero.proof")}
          </p>
        </div>
        <div className="min-w-0">
          <ExpertControlSurface />
        </div>
      </div>
    </section>
  );
}

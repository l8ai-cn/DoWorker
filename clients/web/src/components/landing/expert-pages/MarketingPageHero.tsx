"use client";

import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { useTranslations } from "next-intl";

import { productStoryRoutes } from "../marketing-routes";
import { marketingPageConfig, type MarketingPageId } from "./marketing-page-config";

export function MarketingPageHero({ page }: { page: MarketingPageId }) {
  const t = useTranslations();
  const config = marketingPageConfig[page];

  return (
    <section className="relative border-b border-white/8 bg-[var(--expert-bg)] pb-16 pt-28 sm:pb-20 sm:pt-36">
      <div className="pointer-events-none absolute inset-y-0 left-[8%] border-l border-white/[0.035]" aria-hidden="true" />
      <div className="pointer-events-none absolute inset-y-0 right-[8%] border-r border-white/[0.035]" aria-hidden="true" />
      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl">
          <p className="expert-eyebrow">
            {config.index} / {t(config.eyebrowKey)}
          </p>
          <h1 className="mt-5 text-4xl font-semibold leading-[1.08] text-white sm:text-6xl">
            {t(config.titleKey)}
          </h1>
          <p className="mt-6 max-w-3xl text-base leading-8 text-[var(--expert-muted)] sm:text-lg">
            {t(config.descriptionKey)}
          </p>
          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Link href="/marketplace" className="expert-primary-button">
              {t("landing.workforce.expertHome.hero.primaryAction")}
              <ArrowRight className="h-4 w-4" />
            </Link>
            <Link href={config.nextHref} className="expert-secondary-button">
              {t(config.nextLabelKey)}
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
          <p className="mt-7 border-l-2 border-[var(--expert-warning)] pl-4 text-sm leading-6 text-[var(--expert-text)]">
            {t("landing.workforce.expertHome.hero.proof")}
          </p>
        </div>

        <nav aria-label={t("landing.nav.product")} className="mt-14 grid border-y border-white/10 sm:grid-cols-2">
          {productStoryRoutes.map((route, index) => {
            const active = route.id === page;
            return (
              <Link
                key={route.id}
                href={route.href}
                aria-current={active ? "page" : undefined}
                className={`flex min-h-16 items-center justify-between border-b border-white/10 px-4 text-sm font-semibold transition-colors last:border-b-0 sm:border-b-0 sm:border-r sm:last:border-r-0 ${
                  active
                    ? "bg-[var(--expert-action-soft)] text-[var(--expert-action)]"
                    : "text-[var(--expert-muted)] hover:bg-white/[0.04] hover:text-white"
                }`}
              >
                <span>0{index + 1}</span>
                <span>{t(route.labelKey)}</span>
              </Link>
            );
          })}
        </nav>
      </div>
    </section>
  );
}

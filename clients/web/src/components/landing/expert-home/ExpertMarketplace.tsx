"use client";

import Link from "next/link";
import { ArrowRight, GitCompareArrows, Network, Rocket } from "lucide-react";
import { useTranslations } from "next-intl";

import type { LocalizedMarketApplication } from "./expert-home-content";

const icons = [Rocket, Network, GitCompareArrows];

export function ExpertMarketplace() {
  const t = useTranslations();
  const apps = t.raw("landing.workforce.expertHome.market.apps") as LocalizedMarketApplication[];

  return (
    <section id="marketplace" className="scroll-mt-24 border-y border-white/8 bg-[var(--expert-bg-soft)] py-24 sm:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="grid gap-8 lg:grid-cols-[0.8fr_1.2fr]">
          <div>
            <p className="expert-eyebrow">{t("landing.workforce.expertHome.market.eyebrow")}</p>
            <h2 className="mt-4 text-3xl font-semibold leading-tight text-white sm:text-5xl">
              {t("landing.workforce.expertHome.market.title")}
            </h2>
          </div>
          <p className="max-w-2xl text-base leading-7 text-[var(--expert-muted)] lg:pt-8">
            {t("landing.workforce.expertHome.market.description")}
          </p>
        </div>

        <div className="mt-12 grid gap-4 lg:grid-cols-3">
          {apps.map((app, index) => {
            const Icon = icons[index];
            return (
              <article key={app.slug} className="group rounded-lg border border-white/10 bg-[var(--expert-panel)] p-6 transition-all hover:-translate-y-1 hover:border-[var(--expert-status)]/45">
                <div className="flex items-center justify-between">
                  <Icon className="h-5 w-5 text-[var(--expert-status)]" />
                  <span className="rounded-full border border-[var(--expert-status)] px-2.5 py-1 text-[10px] font-semibold text-[var(--expert-status)]">
                    {t("landing.workforce.expertHome.market.status")}
                  </span>
                </div>
                <p className="mt-8 text-xs font-semibold text-[var(--expert-muted)]">0{index + 1}</p>
                <h3 className="mt-3 text-xl font-semibold text-white transition-colors group-hover:text-[var(--expert-status)]">{app.title}</h3>
                <p className="mt-3 min-h-20 text-sm leading-6 text-[var(--expert-muted)]">{app.description}</p>
                <code className="mt-5 block overflow-hidden text-ellipsis whitespace-nowrap border-t border-white/10 pt-4 text-xs text-[var(--expert-text)]">
                  {app.slug}
                </code>
              </article>
            );
          })}
        </div>

        <div className="mt-8 flex flex-col gap-5 border-l-2 border-[var(--expert-action)] bg-[var(--expert-action-soft)] p-5 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h3 className="font-semibold text-white">{t("landing.workforce.expertHome.market.plannedTitle")}</h3>
            <p className="mt-2 max-w-3xl text-sm leading-6 text-[var(--expert-muted)]">
              {t("landing.workforce.expertHome.market.plannedDescription")}
            </p>
          </div>
          <Link href="/marketplace" className="inline-flex shrink-0 items-center gap-2 text-sm font-semibold text-[var(--expert-action)]">
            {t("landing.workforce.expertHome.market.action")}
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </div>
    </section>
  );
}

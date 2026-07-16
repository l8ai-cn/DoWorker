"use client";

import { Fingerprint, HardDrive, KeyRound, ScrollText } from "lucide-react";
import { useTranslations } from "next-intl";

import { workerTypes, type LocalizedTrustItem } from "./expert-home-content";

const icons = [HardDrive, Fingerprint, KeyRound, ScrollText];

export function ExpertGovernance() {
  const t = useTranslations();
  const items = t.raw("landing.workforce.expertHome.trust.items") as LocalizedTrustItem[];

  return (
    <section id="governance" className="scroll-mt-24 bg-[var(--expert-paper)] py-24 text-[var(--expert-paper-ink)] sm:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="grid gap-8 lg:grid-cols-[0.8fr_1.2fr]">
          <div>
            <p className="expert-eyebrow expert-eyebrow-dark">
              {t("landing.workforce.expertHome.trust.eyebrow")}
            </p>
            <h2 className="mt-4 text-3xl font-semibold leading-tight sm:text-5xl">
              {t("landing.workforce.expertHome.trust.title")}
            </h2>
          </div>
          <p className="max-w-2xl text-base leading-7 text-[var(--expert-paper-muted)] lg:pt-8">
            {t("landing.workforce.expertHome.trust.description")}
          </p>
        </div>

        <div className="mt-12 grid gap-px overflow-hidden border border-black/10 bg-black/10 sm:grid-cols-2 lg:grid-cols-4">
          {items.map((item, index) => {
            const Icon = icons[index];
            return (
              <article key={item.id} className="group min-h-56 bg-white p-5 transition-colors hover:bg-[#FCFDFB]">
                <div className="flex items-center justify-between">
                  <Icon className="h-5 w-5 text-[var(--expert-paper-accent)]" />
                  <span className="text-[10px] font-semibold uppercase text-[var(--expert-paper-accent)]">
                    {item.status}
                  </span>
                </div>
                <p className="mt-9 text-xs font-semibold text-[var(--expert-paper-muted)]">0{index + 1}</p>
                <h3 className="mt-3 text-lg font-semibold transition-colors group-hover:text-[var(--expert-paper-accent)]">{item.title}</h3>
                <p className="mt-3 text-sm leading-6 text-[var(--expert-paper-muted)]">{item.description}</p>
              </article>
            );
          })}
        </div>

        <div className="mt-14 grid gap-6 border-t border-black/10 pt-8 lg:grid-cols-[0.5fr_1.5fr]">
          <div>
            <h3 className="text-xl font-semibold">{t("landing.workforce.expertHome.trust.runtimesTitle")}</h3>
            <p className="mt-3 text-sm leading-6 text-[var(--expert-paper-muted)]">
              {t("landing.workforce.expertHome.trust.runtimesDescription")}
            </p>
          </div>
          <ul className="grid grid-cols-2 gap-x-6 gap-y-3 sm:grid-cols-3 lg:grid-cols-4">
            {workerTypes.map((worker) => (
              <li key={worker.slug} className="border-b border-black/10 pb-2 text-sm font-medium">
                {worker.name}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}

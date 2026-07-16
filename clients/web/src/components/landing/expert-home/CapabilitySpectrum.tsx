"use client";

import {
  Factory,
  PanelsTopLeft,
  ShieldCheck,
  Store,
  Workflow,
} from "lucide-react";
import { useTranslations } from "next-intl";

import type { CapabilityLevel, LocalizedCapability } from "./expert-home-content";

const icons = [Factory, Store, PanelsTopLeft, Workflow, ShieldCheck];
const levelClasses: Record<CapabilityLevel, string> = {
  implemented: "border-[var(--expert-status)] text-[var(--expert-status)]",
  composable: "border-[var(--expert-info)] text-[var(--expert-info)]",
  planned: "border-[var(--expert-planned)] text-[var(--expert-planned)]",
};

export function CapabilitySpectrum({ showIntro = true }: { showIntro?: boolean }) {
  const t = useTranslations();
  const items = t.raw("landing.workforce.expertHome.capabilities.items") as LocalizedCapability[];

  return (
    <section id="capabilities" className="scroll-mt-24 border-y border-white/8 bg-[var(--expert-bg)] py-24 sm:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {showIntro && (
          <div className="grid gap-8 lg:grid-cols-[0.8fr_1.2fr] lg:items-end">
            <div>
              <p className="expert-eyebrow">
                {t("landing.workforce.expertHome.capabilities.eyebrow")}
              </p>
              <h2 className="mt-4 text-3xl font-semibold leading-tight text-white sm:text-5xl">
                {t("landing.workforce.expertHome.capabilities.title")}
              </h2>
            </div>
            <div>
              <p className="max-w-2xl text-base leading-7 text-[var(--expert-muted)]">
                {t("landing.workforce.expertHome.capabilities.description")}
              </p>
            </div>
          </div>
        )}
        {!showIntro && (
          <h2 className="sr-only">
            {t("landing.workforce.expertHome.capabilities.title")}
          </h2>
        )}

        <div className={`${showIntro ? "mt-5 " : ""}flex flex-wrap gap-2`}>
          {(["implemented", "composable", "planned"] as CapabilityLevel[]).map((level) => (
            <span key={level} className={`rounded-full border px-3 py-1 text-xs font-semibold ${levelClasses[level]}`}>
              {t(`landing.workforce.expertHome.capabilities.levels.${level}`)}
            </span>
          ))}
        </div>

        <div className={`${showIntro ? "mt-12 " : "mt-8 "}grid gap-px overflow-hidden rounded-lg border border-white/10 bg-white/10 sm:grid-cols-2 lg:grid-cols-5`}>
          {items.map((item, index) => {
            const Icon = icons[index];
            return (
              <article key={item.id} className="group min-h-48 bg-[var(--expert-panel)] p-5 transition-colors hover:bg-[#19232D]">
                <div className="flex items-start justify-between gap-4">
                  <Icon className="h-5 w-5 text-[var(--expert-text)]" />
                  <span className={`rounded-full border px-2.5 py-1 text-[10px] font-semibold ${levelClasses[item.level]}`}>
                    {t(`landing.workforce.expertHome.capabilities.levels.${item.level}`)}
                  </span>
                </div>
                <h3 className="mt-8 text-lg font-semibold text-white transition-colors group-hover:text-[var(--expert-action)]">{item.title}</h3>
                <p className="mt-3 text-sm leading-6 text-[var(--expert-muted)]">{item.description}</p>
              </article>
            );
          })}
        </div>
      </div>
    </section>
  );
}

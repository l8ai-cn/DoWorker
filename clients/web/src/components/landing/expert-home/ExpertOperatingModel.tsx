"use client";

import {
  Blocks,
  BrainCircuit,
  CheckCircle2,
  Cpu,
  Database,
  Plug,
  Workflow,
} from "lucide-react";
import { useTranslations } from "next-intl";

import type { LocalizedContentItem } from "./expert-home-content";

const partIcons = [Cpu, BrainCircuit, Blocks, Database, Plug, Workflow];

export function ExpertOperatingModel({ showIntro = true }: { showIntro?: boolean }) {
  const t = useTranslations();
  const parts = t.raw("landing.workforce.expertHome.operating.parts") as LocalizedContentItem[];
  const humanItems = t.raw("landing.workforce.expertHome.operating.humanItems") as LocalizedContentItem[];

  return (
    <section id="operating-model" className="scroll-mt-24 border-b border-white/8 bg-[var(--expert-bg-soft)] py-24 sm:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {showIntro && (
          <div className="max-w-6xl">
            <p className="expert-eyebrow">
              {t("landing.workforce.expertHome.operating.eyebrow")}
            </p>
            <h2 className="mt-4 text-3xl font-semibold leading-tight text-white sm:text-5xl">
              {t("landing.workforce.expertHome.operating.title")}
            </h2>
            <p className="mt-5 max-w-4xl text-base leading-7 text-[var(--expert-muted)]">
              {t("landing.workforce.expertHome.operating.description")}
            </p>
          </div>
        )}
        {!showIntro && (
          <h2 className="sr-only">
            {t("landing.workforce.expertHome.operating.title")}
          </h2>
        )}

        <div className={`${showIntro ? "mt-14 " : ""}border-y border-white/10 py-8`}>
          <p className="text-xs font-semibold uppercase text-[var(--expert-muted)]">
            {t("landing.workforce.expertHome.operating.formulaTitle")}
          </p>
          <div className="mt-6 grid overflow-hidden rounded-lg border border-white/10 sm:grid-cols-2 lg:grid-cols-3">
            {parts.map((part, index) => {
              const Icon = partIcons[index];
              return (
                <article
                  key={part.id}
                  className="group min-h-44 border-b border-white/10 bg-[var(--expert-panel)] p-5 transition-colors hover:bg-[#19232D] sm:border-r sm:[&:nth-child(2n)]:border-r-0 sm:[&:nth-child(n+5)]:border-b-0 lg:[&:nth-child(2n)]:border-r lg:[&:nth-child(3n)]:border-r-0 lg:[&:nth-child(n+4)]:border-b-0"
                >
                  <div className="flex items-center justify-between">
                    <span className="flex h-9 w-9 items-center justify-center rounded-md border border-white/10 text-[var(--expert-action)]">
                      <Icon className="h-4 w-4" />
                    </span>
                    <span className="text-xs font-semibold text-[var(--expert-muted)]">0{index + 1}</span>
                  </div>
                  <h3 className="mt-6 text-base font-semibold text-white">{part.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-[var(--expert-muted)]">{part.description}</p>
                </article>
              );
            })}
          </div>
        </div>

        <div className="mt-14 grid gap-10 lg:grid-cols-[0.8fr_1.2fr]">
          <div>
            <h3 className="text-2xl font-semibold text-white">
              {t("landing.workforce.expertHome.operating.humanTitle")}
            </h3>
            <p className="mt-4 text-base leading-7 text-[var(--expert-muted)]">
              {t("landing.workforce.expertHome.operating.humanDescription")}
            </p>
          </div>
          <div className="divide-y divide-white/10 border-y border-white/10">
            {humanItems.map((item) => (
              <article key={item.id} className="grid gap-3 py-5 sm:grid-cols-[24px_180px_1fr]">
                <CheckCircle2 className="mt-0.5 h-5 w-5 text-[var(--expert-status)]" />
                <h4 className="font-semibold text-white">{item.title}</h4>
                <p className="text-sm leading-6 text-[var(--expert-muted)]">{item.description}</p>
              </article>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

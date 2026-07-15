"use client";

import { useEffect, useId, useRef, useState, type KeyboardEvent } from "react";
import { ArrowRight, BookOpen, BriefcaseBusiness, ShoppingBag, Store } from "lucide-react";
import { useTranslations } from "next-intl";

import type { LocalizedSolution } from "./expert-home-content";

const icons = [ShoppingBag, BookOpen, BriefcaseBusiness, Store];

export function SolutionDomains({ showIntro = true }: { showIntro?: boolean }) {
  const t = useTranslations();
  const items = t.raw("landing.workforce.expertHome.solutions.items") as LocalizedSolution[];
  const [activeIndex, setActiveIndex] = useState(0);
  const baseId = useId();
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);

  useEffect(() => {
    const syncHash = () => {
      const index = items.findIndex(({ id }) => `#${id}` === window.location.hash);
      if (index >= 0) setActiveIndex(index);
    };
    syncHash();
    window.addEventListener("hashchange", syncHash);
    return () => window.removeEventListener("hashchange", syncHash);
  }, [items]);

  const activate = (index: number) => {
    setActiveIndex(index);
    tabRefs.current[index]?.focus();
  };

  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
    if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
    event.preventDefault();
    if (event.key === "Home") return activate(0);
    if (event.key === "End") return activate(items.length - 1);
    const delta = event.key === "ArrowRight" ? 1 : -1;
    activate((index + delta + items.length) % items.length);
  };

  return (
    <section id="solutions" className="scroll-mt-24 bg-[var(--expert-paper)] py-24 text-[var(--expert-paper-ink)] sm:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {showIntro && (
          <div className="max-w-3xl">
            <p className="expert-eyebrow expert-eyebrow-dark">
              {t("landing.workforce.expertHome.solutions.eyebrow")}
            </p>
            <h2 className="mt-4 text-3xl font-semibold leading-tight sm:text-5xl">
              {t("landing.workforce.expertHome.solutions.title")}
            </h2>
            <p className="mt-5 max-w-2xl text-base leading-7 text-[var(--expert-paper-muted)]">
              {t("landing.workforce.expertHome.solutions.description")}
            </p>
          </div>
        )}
        {!showIntro && (
          <h2 className="sr-only">
            {t("landing.workforce.expertHome.solutions.title")}
          </h2>
        )}

        <div className={`${showIntro ? "mt-12 " : ""}grid gap-10 lg:grid-cols-[340px_1fr]`}>
          <div
            role="tablist"
            aria-label={t("landing.workforce.expertHome.solutions.eyebrow")}
            className="grid gap-2 sm:grid-cols-2 lg:grid-cols-1"
          >
            {items.map((item, index) => {
              const Icon = icons[index];
              const selected = index === activeIndex;
              return (
                <button
                  key={item.id}
                  ref={(node) => { tabRefs.current[index] = node; }}
                  id={item.id === "marketplace" ? `${baseId}-tab-${index}` : item.id}
                  role="tab"
                  type="button"
                  aria-label={item.title}
                  aria-selected={selected}
                  aria-controls={`${baseId}-panel-${index}`}
                  tabIndex={selected ? 0 : -1}
                  onClick={() => setActiveIndex(index)}
                  onKeyDown={(event) => onKeyDown(event, index)}
                  className={`group flex min-h-20 items-center gap-4 rounded-md border px-4 py-3 text-left transition-all ${
                    selected
                      ? "border-[var(--expert-paper-ink)] bg-[var(--expert-paper-ink)] text-white shadow-[0_14px_34px_rgba(20,24,29,0.16)]"
                      : "border-black/10 bg-white text-[var(--expert-paper-ink)] hover:-translate-y-0.5 hover:border-black/30"
                  }`}
                >
                  <span
                    className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-md border ${
                      selected
                        ? "border-white/20 bg-white/10"
                        : "border-black/10 bg-[var(--expert-paper)]"
                    }`}
                  >
                    <Icon className="h-4 w-4" />
                  </span>
                  <span className="min-w-0">
                    <span
                      className={`block text-[10px] font-semibold ${
                        selected ? "text-white/60" : "text-[var(--expert-paper-muted)]"
                      }`}
                    >
                      0{index + 1}
                    </span>
                    <span className="mt-1 block text-sm font-semibold">{item.title}</span>
                  </span>
                </button>
              );
            })}
          </div>

          <div>
            {items.map((item, index) => (
              <article
                key={item.id}
                id={`${baseId}-panel-${index}`}
                role="tabpanel"
                aria-labelledby={item.id === "marketplace" ? `${baseId}-tab-${index}` : item.id}
                tabIndex={index === activeIndex ? 0 : -1}
                hidden={index !== activeIndex}
                className="min-h-[390px] border-t-2 border-[var(--expert-paper-ink)] pt-7"
              >
                <p className="text-xs font-semibold text-[var(--expert-paper-accent)]">
                  0{index + 1} / 0{items.length}
                </p>
                <h3 className="mt-3 text-3xl font-semibold">{item.title}</h3>
                <p className="mt-4 max-w-2xl text-lg leading-8 text-[var(--expert-paper-muted)]">
                  {item.description}
                </p>
                <div className="mt-8 grid gap-px overflow-hidden border-y border-black/10 bg-black/10 sm:grid-cols-2">
                  <div className="h-full bg-white px-5 py-6">
                    <p className="text-xs font-semibold uppercase text-[var(--expert-paper-muted)]">
                      {t("landing.workforce.expertHome.solutions.workflowLabel")}
                    </p>
                    <p className="mt-3 text-base font-medium leading-7">{item.chain}</p>
                  </div>
                  <div className="h-full bg-white px-5 py-6">
                    <p className="text-xs font-semibold uppercase text-[var(--expert-paper-muted)]">
                      {t("landing.workforce.expertHome.solutions.deliverablesLabel")}
                    </p>
                    <p className="mt-3 text-base font-medium leading-7">{item.outcome}</p>
                  </div>
                </div>
                <a
                  href={item.id === "marketplace" ? "/marketplace" : "/register"}
                  className="mt-7 inline-flex items-center gap-2 text-sm font-semibold text-[var(--expert-paper-accent)]"
                >
                  {item.action}
                  <ArrowRight className="h-4 w-4" />
                </a>
              </article>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

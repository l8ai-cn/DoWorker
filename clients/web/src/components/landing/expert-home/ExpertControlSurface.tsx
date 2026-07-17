"use client";

import { useState } from "react";
import { Check, Pause, Play, RotateCcw, SkipForward } from "lucide-react";
import { useTranslations } from "next-intl";

export function ExpertControlSurface() {
  const t = useTranslations();
  const steps = t.raw("landing.workforce.expertHome.console.steps") as string[];
  const [activeStep, setActiveStep] = useState(2);
  const [paused, setPaused] = useState(false);
  const complete = activeStep === steps.length - 1;
  const state = paused ? "paused" : complete ? "complete" : "live";

  const replay = () => {
    setPaused(false);
    setActiveStep(0);
  };

  const next = () => setActiveStep((step) => Math.min(step + 1, steps.length - 1));

  return (
    <section
      aria-label={t("landing.workforce.expertHome.console.title")}
      className="overflow-hidden rounded-lg border border-white/10 bg-[var(--expert-panel)] shadow-[var(--expert-shadow)]"
      tabIndex={-1}
    >
      <header className="flex items-center justify-between border-b border-white/10 px-4 py-3 sm:px-5">
        <div>
          <p className="text-sm font-semibold text-white">
            {t("landing.workforce.expertHome.console.title")}
          </p>
          <p className="mt-1 text-xs text-[var(--expert-muted)]">
            {t("landing.workforce.expertHome.console.status")}
          </p>
        </div>
        <span
          aria-live="polite"
          className={`flex items-center gap-2 text-xs font-medium ${
            paused ? "text-[var(--expert-warning)]" : "text-[var(--expert-status)]"
          }`}
        >
          <span
            className={`h-2 w-2 rounded-full ${
              paused ? "bg-[var(--expert-warning)]" : "bg-[var(--expert-status)]"
            }`}
          />
          {t(`landing.workforce.expertHome.console.${state}`)}
        </span>
      </header>

      <div className="grid gap-px bg-white/10 sm:grid-cols-3">
        {(["goal", "expert", "formula"] as const).map((item) => (
          <div key={item} className={`bg-[var(--expert-panel)] px-4 py-4 sm:px-5 ${item === "formula" ? "hidden sm:block" : ""}`}>
            <p className="text-[10px] font-semibold uppercase text-[var(--expert-muted)]">
              {t(`landing.workforce.expertHome.console.${item}Label`)}
            </p>
            <p className="mt-2 text-sm leading-6 text-white">
              {t(`landing.workforce.expertHome.console.${item}`)}
            </p>
          </div>
        ))}
      </div>

      <div className="px-4 py-5 sm:px-5">
        <p className="text-[10px] font-semibold uppercase text-[var(--expert-muted)]">
          {t("landing.workforce.expertHome.console.workflowLabel")}
        </p>
        <ol className="mt-4 grid grid-cols-3 gap-3 sm:grid-cols-6 sm:gap-2">
          {steps.map((step, index) => {
            const complete = index < activeStep;
            const active = index === activeStep;
            return (
              <li key={step} className="min-w-0">
                <div
                  aria-current={active ? "step" : undefined}
                  className={`flex h-8 w-8 items-center justify-center rounded-full border text-xs font-semibold ${
                    complete
                      ? "border-[var(--expert-status)] bg-[var(--expert-status)] text-[var(--expert-ink)]"
                      : active
                        ? "border-[var(--expert-action)] text-[var(--expert-action)]"
                        : "border-white/15 text-[var(--expert-muted)]"
                  }`}
                >
                  {complete ? <Check className="h-4 w-4" /> : index + 1}
                </div>
                <p className={`mt-2 text-xs leading-4 ${active ? "text-white" : "text-[var(--expert-muted)]"}`}>
                  {step}
                </p>
              </li>
            );
          })}
        </ol>

        <div className="mt-5 grid gap-3 sm:grid-cols-2">
          <div className="border-l-2 border-[var(--expert-warning)] bg-[var(--expert-warning-soft)] px-4 py-3">
            <p className="text-xs font-semibold text-[var(--expert-warning)]">
              {t("landing.workforce.expertHome.console.checkpointLabel")}
            </p>
            <p className="mt-1 text-xs leading-5 text-[var(--expert-text)]">
              {t("landing.workforce.expertHome.console.checkpoint")}
            </p>
          </div>
          <div className="border-l-2 border-[var(--expert-status)] bg-[var(--expert-status-soft)] px-4 py-3">
            <p className="text-xs font-semibold text-[var(--expert-status)]">
              {t("landing.workforce.expertHome.console.deliverableLabel")}
            </p>
            <p className="mt-1 text-xs leading-5 text-[var(--expert-text)]">
              {t("landing.workforce.expertHome.console.deliverable")}
            </p>
          </div>
        </div>
      </div>

      <footer className="flex flex-wrap gap-2 border-t border-white/10 px-4 py-3 sm:px-5">
        <button
          type="button"
          onClick={() => setPaused((value) => !value)}
          className="expert-icon-button"
          aria-label={t(`landing.workforce.expertHome.console.controls.${paused ? "resume" : "pause"}`)}
          title={t(`landing.workforce.expertHome.console.controls.${paused ? "resume" : "pause"}`)}
        >
          {paused ? <Play /> : <Pause />}
        </button>
        <button
          type="button"
          onClick={next}
          disabled={paused || complete}
          className="expert-icon-button"
          aria-label={t("landing.workforce.expertHome.console.controls.next")}
          title={t("landing.workforce.expertHome.console.controls.next")}
        >
          <SkipForward />
        </button>
        <button type="button" onClick={replay} className="expert-icon-button" aria-label={t("landing.workforce.expertHome.console.controls.replay")} title={t("landing.workforce.expertHome.console.controls.replay")}>
          <RotateCcw />
        </button>
      </footer>
    </section>
  );
}

"use client";

import { useTranslations } from "next-intl";

export function WhyTerminalBased() {
  const t = useTranslations();

  const benefits = [
    {
      icon: (
        <svg className="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      ),
      title: t("landing.whyTerminal.benefits.autonomous.title"),
      description: t("landing.whyTerminal.benefits.autonomous.description"),
    },
    {
      icon: (
        <svg className="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
      ),
      title: t("landing.whyTerminal.benefits.fullCapabilities.title"),
      description: t("landing.whyTerminal.benefits.fullCapabilities.description"),
    },
    {
      icon: (
        <svg className="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
      ),
      title: t("landing.whyTerminal.benefits.dataControl.title"),
      description: t("landing.whyTerminal.benefits.dataControl.description"),
    },
  ];

  const comparisonData = [
    { feature: t("landing.whyTerminal.comparison.autonomy"), ide: t("landing.whyTerminal.comparison.ide.autonomy"), terminal: t("landing.whyTerminal.comparison.terminal.autonomy") },
    { feature: t("landing.whyTerminal.comparison.capabilities"), ide: t("landing.whyTerminal.comparison.ide.capabilities"), terminal: t("landing.whyTerminal.comparison.terminal.capabilities") },
    { feature: t("landing.whyTerminal.comparison.environment"), ide: t("landing.whyTerminal.comparison.ide.environment"), terminal: t("landing.whyTerminal.comparison.terminal.environment") },
    { feature: t("landing.whyTerminal.comparison.multiAgent"), ide: t("landing.whyTerminal.comparison.ide.multiAgent"), terminal: t("landing.whyTerminal.comparison.terminal.multiAgent") },
    { feature: t("landing.whyTerminal.comparison.selfHosted"), ide: t("landing.whyTerminal.comparison.ide.selfHosted"), terminal: t("landing.whyTerminal.comparison.terminal.selfHosted") },
  ];

  return (
    <section className="py-32 relative overflow-hidden" id="why-terminal">
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[700px] bg-[var(--azure-cyan)]/5 blur-[150px] rounded-full pointer-events-none" />

      <div className="container mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="text-center mb-20">
          <h2 className="font-headline text-4xl md:text-5xl font-bold mb-6 leading-tight">
            {t("landing.whyTerminal.title")}{" "}
            <span className="azure-gradient-text">{t("landing.whyTerminal.titleHighlight")}</span>
            {t("landing.whyTerminal.titleEnd")}
          </h2>
          <p className="text-[var(--azure-text-muted)] max-w-2xl mx-auto text-lg font-light">
            {t("landing.whyTerminal.description")}
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8 mb-16 max-w-6xl mx-auto">
          {benefits.map((benefit, index) => (
            <div
              key={index}
              className="azure-glass p-8 rounded-3xl border border-white/5 hover:border-[var(--azure-cyan)]/30 transition-all group"
            >
              <div className="w-14 h-14 rounded-2xl bg-[var(--azure-cyan)]/10 flex items-center justify-center text-[var(--azure-cyan)] mb-6 group-hover:azure-glow-cyan transition-all">
                {benefit.icon}
              </div>
              <h3 className="font-headline text-xl font-bold mb-3">{benefit.title}</h3>
              <p className="text-[var(--azure-text-muted)] leading-relaxed font-light">{benefit.description}</p>
            </div>
          ))}
        </div>

        <div className="max-w-4xl mx-auto azure-glass rounded-3xl border border-white/5 overflow-hidden">
          <div className="grid grid-cols-3 border-b border-white/5">
            <div className="p-5" />
            <div className="p-5 text-center">
              <div className="font-headline text-sm font-bold text-[var(--azure-text-muted)]">
                {t("landing.whyTerminal.comparison.idePlugins")}
              </div>
              <div className="text-xs text-[var(--azure-text-muted)]/60 mt-1">
                {t("landing.whyTerminal.comparison.idePluginsSubtitle")}
              </div>
            </div>
            <div className="p-5 text-center bg-[var(--azure-cyan)]/[0.06]">
              <div className="font-headline text-sm font-bold azure-gradient-text">
                {t("landing.whyTerminal.comparison.terminalBased")}
              </div>
              <div className="text-xs text-[var(--azure-cyan)]/70 mt-1">
                {t("landing.whyTerminal.comparison.terminalBasedSubtitle")}
              </div>
            </div>
          </div>

          {comparisonData.map((row, index) => (
            <div key={index} className="grid grid-cols-3 border-b border-white/5 last:border-b-0">
              <div className="p-5 font-medium text-sm text-foreground/90">{row.feature}</div>
              <div className="p-5 text-sm text-center text-[var(--azure-text-muted)]">
                <span className="inline-flex items-center gap-1.5">
                  <svg className="w-4 h-4 text-[var(--azure-text-muted)]/50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  {row.ide}
                </span>
              </div>
              <div className="p-5 text-sm text-center bg-[var(--azure-cyan)]/[0.06]">
                <span className="inline-flex items-center gap-1.5 text-[var(--azure-mint)]">
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  {row.terminal}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

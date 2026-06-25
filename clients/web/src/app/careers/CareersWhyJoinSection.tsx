"use client";

import { useTranslations } from "next-intl";

const WHY_JOIN_KEYS = [
  { icon: "innovation", path: "careers.whyJoin.innovation" },
  { icon: "remote", path: "careers.whyJoin.remote" },
  { icon: "equity", path: "careers.whyJoin.equity" },
] as const;

function WhyJoinIcon({ type }: { type: (typeof WHY_JOIN_KEYS)[number]["icon"] }) {
  if (type === "innovation") {
    return (
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
      />
    );
  }
  if (type === "remote") {
    return (
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    );
  }
  return (
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
    />
  );
}

export function CareersWhyJoinSection() {
  const t = useTranslations();

  return (
    <section className="py-16 px-4 bg-surface-muted/40">
      <div className="container mx-auto max-w-4xl text-center">
        <h2 className="text-3xl font-bold mb-8">{t("careers.whyJoin.title")}</h2>
        <div className="grid md:grid-cols-3 gap-8">
          {WHY_JOIN_KEYS.map(({ icon, path }) => (
            <div key={icon} className="surface-card-interactive p-6 motion-interactive text-left">
              <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mb-4 mx-auto md:mx-0">
                <svg className="w-6 h-6 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <WhyJoinIcon type={icon} />
                </svg>
              </div>
              <h3 className="text-lg font-semibold mb-2 text-center md:text-left">{t(`${path}.title`)}</h3>
              <p className="text-muted-foreground text-center md:text-left">{t(`${path}.content`)}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

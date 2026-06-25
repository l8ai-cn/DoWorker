"use client";

import { useTranslations } from "next-intl";

export function TryItSection() {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <div className="surface-card bg-primary/5 ring-1 ring-primary/20 p-6">
        <h2 className="text-xl font-semibold mb-4 text-foreground">{t("docs.gettingStarted.tryIt.title")}</h2>
        <p className="text-muted-foreground mb-4">{t("docs.gettingStarted.tryIt.description")}</p>
        <ol className="list-decimal list-inside text-muted-foreground space-y-2 mb-4">
          {Array.from({ length: 5 }, (_, i) => (
            <li key={i}>{t(`docs.gettingStarted.tryIt.item${i + 1}`)}</li>
          ))}
        </ol>
        <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
          {t("docs.gettingStarted.tryIt.tip")}
        </div>
      </div>
    </section>
  );
}

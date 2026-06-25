"use client";

import { useTranslations } from "next-intl";
import { PageHeader, PageFooter } from "@/components/common";
import type { ChangelogChangeType, ChangelogEntry } from "./changelog-types";

const TYPE_COLORS: Record<ChangelogChangeType, string> = {
  added: "bg-success/20 text-success",
  changed: "bg-info/20 text-info",
  fixed: "bg-warning/20 text-warning",
  removed: "bg-danger/20 text-danger",
};

export default function ChangelogPage() {
  const t = useTranslations();
  const entries = t.raw("entries") as ChangelogEntry[];

  return (
    <div className="azure-theme min-h-screen bg-background">
      <PageHeader />

      <main className="container mx-auto px-4 py-12 max-w-4xl">
        <h1 className="text-4xl font-bold mb-4">{t("changelogPage.title")}</h1>
        <p className="text-muted-foreground mb-12">{t("changelogPage.description")}</p>

        <div className="space-y-12">
          {entries.map((entry) => (
            <article key={entry.version} className="surface-card p-6 motion-interactive">
              <div className="flex items-center gap-4 mb-6">
                <h2 className="text-2xl font-bold">v{entry.version}</h2>
                <time className="text-sm text-muted-foreground">
                  {new Date(entry.date).toLocaleDateString(undefined, {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })}
                </time>
              </div>

              <div className="space-y-6 pl-4 border-l border-primary/20">
                {entry.changes.map((change, idx) => (
                  <div key={idx}>
                    <span
                      className={`inline-block px-2 py-1 rounded text-xs font-medium mb-3 ${TYPE_COLORS[change.type]}`}
                    >
                      {t(`changelogPage.types.${change.type}`)}
                    </span>
                    <ul className="space-y-2">
                      {change.items.map((item, itemIdx) => (
                        <li key={itemIdx} className="text-muted-foreground flex items-start gap-2">
                          <span className="text-primary mt-1.5">•</span>
                          {item}
                        </li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>
            </article>
          ))}
        </div>
      </main>

      <PageFooter />
    </div>
  );
}

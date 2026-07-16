"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";

const NEXT_LINKS = [
  { href: "/docs/concepts/workers", titleKey: "agentpod", descKey: "agentpodDesc" },
  { href: "/docs/features/tickets", titleKey: "tickets", descKey: "ticketsDesc" },
  { href: "/docs/guides/multi-agent-workflows", titleKey: "multiAgent", descKey: "multiAgentDesc" },
] as const;

export function NextStepsSection() {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <h2 className="text-2xl font-semibold mb-4 text-foreground">
        {t("docs.gettingStarted.nextSteps.title")}
      </h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {NEXT_LINKS.map(({ href, titleKey, descKey }) => (
          <Link key={href} href={href} className="surface-card-interactive p-4 block motion-interactive">
            <h3 className="font-medium mb-1">{t(`docs.gettingStarted.nextSteps.${titleKey}`)}</h3>
            <p className="text-sm text-muted-foreground">
              {t(`docs.gettingStarted.nextSteps.${descKey}`)}
            </p>
          </Link>
        ))}
      </div>
    </section>
  );
}

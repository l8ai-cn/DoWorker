"use client";

import { PageHeader, PageFooter } from "@/components/common";
import { EnterpriseFeatures, SelfHostedCTA } from "@/components/landing";
import { useTranslations } from "next-intl";

export default function EnterprisePage() {
  const t = useTranslations();

  return (
    <div className="azure-theme min-h-screen bg-background">
      <PageHeader />

      <section className="pt-20 pb-8 px-4">
        <div className="container mx-auto max-w-4xl text-center">
          <h1 className="text-4xl md:text-5xl font-bold mb-6">
            {t("enterprise.hero.title")}{" "}
            <span className="text-primary">{t("enterprise.hero.titleHighlight")}</span>
          </h1>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            {t("enterprise.hero.subtitle")}
          </p>
        </div>
      </section>

      <EnterpriseFeatures />
      <SelfHostedCTA />

      <PageFooter />
    </div>
  );
}

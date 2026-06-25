"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { PageHeader, PageFooter } from "@/components/common";
import { CareersWhyJoinSection } from "./CareersWhyJoinSection";
import { useTranslations } from "next-intl";

interface JobPosition {
  id: string;
  titleKey: string;
  descriptionKey: string;
  responsibilities: string[];
  icon: React.ReactNode;
}

const positions: JobPosition[] = [
  {
    id: "harnessEngineer",
    titleKey: "careers.positions.harnessEngineer.title",
    descriptionKey: "careers.positions.harnessEngineer.description",
    responsibilities: [
      "careers.positions.harnessEngineer.resp1",
      "careers.positions.harnessEngineer.resp2",
      "careers.positions.harnessEngineer.resp3",
    ],
    icon: (
      <svg
        className="w-8 h-8"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
        />
      </svg>
    ),
  },
  {
    id: "producer",
    titleKey: "careers.positions.producer.title",
    descriptionKey: "careers.positions.producer.description",
    responsibilities: [
      "careers.positions.producer.resp1",
      "careers.positions.producer.resp2",
      "careers.positions.producer.resp3",
    ],
    icon: (
      <svg
        className="w-8 h-8"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
        />
      </svg>
    ),
  },
  {
    id: "growthHacker",
    titleKey: "careers.positions.growthHacker.title",
    descriptionKey: "careers.positions.growthHacker.description",
    responsibilities: [
      "careers.positions.growthHacker.resp1",
      "careers.positions.growthHacker.resp2",
      "careers.positions.growthHacker.resp3",
    ],
    icon: (
      <svg
        className="w-8 h-8"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
        />
      </svg>
    ),
  },
];

export default function CareersPage() {
  const t = useTranslations();

  return (
    <div className="azure-theme min-h-screen bg-background">
      <PageHeader />

      {/* Hero Section */}
      <section className="py-20 px-4 text-center">
        <div className="container mx-auto max-w-4xl">
          <h1 className="text-4xl md:text-5xl font-bold mb-6">
            {t("careers.hero.title")}
          </h1>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto mb-8">
            {t("careers.hero.subtitle")}
          </p>
        </div>
      </section>

      {/* Positions */}
      <section className="py-16 px-4">
        <div className="container mx-auto max-w-5xl">
          <h2 className="text-3xl font-bold mb-12 text-center">
            {t("careers.openPositions")}
          </h2>

          <div className="space-y-8">
            {positions.map((position) => (
              <div
                key={position.id}
                className="surface-card-interactive p-8 motion-interactive"
              >
                <div className="flex items-start gap-6">
                  <div className="w-16 h-16 rounded-xl bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                    {position.icon}
                  </div>
                  <div className="flex-1">
                    <h3 className="text-2xl font-bold mb-2">
                      {t(position.titleKey)}
                    </h3>
                    <p className="text-muted-foreground mb-4">
                      {t(position.descriptionKey)}
                    </p>
                    <div className="mb-6">
                      <h4 className="font-semibold mb-2">
                        {t("careers.responsibilities")}
                      </h4>
                      <ul className="space-y-1">
                        {position.responsibilities.map((resp, idx) => (
                          <li
                            key={idx}
                            className="text-muted-foreground flex items-start gap-2"
                          >
                            <span className="text-primary">•</span>
                            {t(resp)}
                          </li>
                        ))}
                      </ul>
                    </div>
                    <Link href="mailto:recruiter@agentsmesh.ai">
                      <Button>{t("careers.applyNow")}</Button>
                    </Link>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <CareersWhyJoinSection />

      {/* CTA */}
      <section className="py-16 px-4">
        <div className="container mx-auto max-w-4xl text-center">
          <h2 className="text-2xl font-bold mb-4">{t("careers.cta.title")}</h2>
          <p className="text-muted-foreground mb-6">{t("careers.cta.content")}</p>
          <Link href="mailto:recruiter@agentsmesh.ai">
            <Button size="lg">{t("careers.cta.button")}</Button>
          </Link>
        </div>
      </section>

      <PageFooter />
    </div>
  );
}

"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";

const STEPS = ["runtime", "configuration", "workspace", "preflight"] as const;

export default function FirstWorkerTutorialPage() {
  const t = useTranslations("docs.workerTutorial");

  return (
    <div>
      <h1 className="text-4xl font-bold mb-4">{t("title")}</h1>
      <p className="max-w-3xl text-muted-foreground leading-relaxed mb-8">
        {t("description")}
      </p>

      <section className="mb-8 rounded-xl border border-border bg-muted/30 p-5 sm:p-6">
        <h2 className="text-xl font-semibold text-foreground">{t("beforeYouStart.title")}</h2>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-muted-foreground">
          <li>{t("beforeYouStart.runner")}</li>
          <li>{t("beforeYouStart.runtime")}</li>
          <li>{t("beforeYouStart.resource")}</li>
        </ul>
      </section>

      {STEPS.map((step, index) => (
        <TutorialStep
          key={step}
          number={index + 1}
          title={t(`steps.${step}.title`)}
          description={t(`steps.${step}.description`)}
          details={t(`steps.${step}.details`)}
        />
      ))}

      <section className="mb-8 surface-card rounded-xl p-5 sm:p-6">
        <h2 className="text-xl font-semibold text-foreground">{t("result.title")}</h2>
        <p className="mt-2 text-muted-foreground leading-relaxed">{t("result.description")}</p>
        <Link href="/docs/concepts/workers" className="mt-4 inline-block text-sm font-medium text-primary hover:underline">
          {t("result.catalogLink")}
        </Link>
      </section>

      <DocNavigation />
    </div>
  );
}

function TutorialStep({
  number,
  title,
  description,
  details,
}: {
  number: number;
  title: string;
  description: string;
  details: string;
}) {
  return (
    <section className="mb-8 surface-card rounded-xl p-5 sm:p-6">
      <div className="flex items-center gap-3">
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-sm font-bold text-primary-foreground">
          {number}
        </div>
        <h2 className="text-xl font-semibold text-foreground">{title}</h2>
      </div>
      <p className="mt-4 text-muted-foreground leading-relaxed">{description}</p>
      <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{details}</p>
    </section>
  );
}

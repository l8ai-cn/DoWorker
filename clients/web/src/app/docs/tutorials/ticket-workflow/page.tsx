"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocStepHeader } from "@/components/docs/DocStepHeader";

export default function TicketWorkflowTutorialPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-2">
        {t("docs.tutorials.ticketWorkflow.title")}
      </h1>
      <p className="text-sm text-muted-foreground mb-8">
        {t("docs.tutorials.ticketWorkflow.difficulty")}
      </p>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.tutorials.ticketWorkflow.description")}
      </p>

      {/* Step 1 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={1} titleKey="docs.tutorials.ticketWorkflow.step1.title" />
          <p className="text-muted-foreground">
            {t("docs.tutorials.ticketWorkflow.step1.description")}
          </p>
        </div>
      </section>

      {/* Step 2 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={2} titleKey="docs.tutorials.ticketWorkflow.step2.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.ticketWorkflow.step2.description")}
          </p>
          <p className="font-medium mb-3">
            {t("docs.tutorials.ticketWorkflow.step2.fields")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.ticketWorkflow.step2.field1")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step2.field2")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step2.field3")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step2.field4")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step2.field5")}</li>
          </ol>
        </div>
      </section>

      {/* Step 3 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={3} titleKey="docs.tutorials.ticketWorkflow.step3.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.ticketWorkflow.step3.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.ticketWorkflow.step3.item1")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step3.item2")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step3.item3")}</li>
          </ul>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.ticketWorkflow.step3.tip")}
          </div>
        </div>
      </section>

      {/* Step 4 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={4} titleKey="docs.tutorials.ticketWorkflow.step4.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.ticketWorkflow.step4.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.ticketWorkflow.step4.item1")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step4.item2")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step4.item3")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step4.item4")}</li>
          </ul>
        </div>
      </section>

      {/* Step 5 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={5} titleKey="docs.tutorials.ticketWorkflow.step5.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.ticketWorkflow.step5.description")}
          </p>
          <div className="space-y-3">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span className="px-2 py-1 bg-muted rounded">
                {t("docs.tutorials.ticketWorkflow.step5.item1")}
              </span>
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span className="px-2 py-1 bg-muted rounded">
                {t("docs.tutorials.ticketWorkflow.step5.item2")}
              </span>
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span className="px-2 py-1 bg-muted rounded">
                {t("docs.tutorials.ticketWorkflow.step5.item3")}
              </span>
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span className="px-2 py-1 bg-muted rounded">
                {t("docs.tutorials.ticketWorkflow.step5.item4")}
              </span>
            </div>
          </div>
        </div>
      </section>

      {/* Step 6 */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={6} titleKey="docs.tutorials.ticketWorkflow.step6.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.ticketWorkflow.step6.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.ticketWorkflow.step6.item1")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step6.item2")}</li>
            <li>{t("docs.tutorials.ticketWorkflow.step6.item3")}</li>
          </ul>
        </div>
      </section>

      {/* Next Steps */}
      <section className="mb-8">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.tutorials.ticketWorkflow.nextSteps.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.tutorials.ticketWorkflow.nextSteps.description")}
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Link
            href="/docs/tutorials/multi-agent-collaboration"
            className="surface-card-interactive p-4 block"
          >
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.ticketWorkflow.nextSteps.item1")}
            </p>
          </Link>
          <Link
            href="/docs/tutorials/automated-loops"
            className="surface-card-interactive p-4 block"
          >
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.ticketWorkflow.nextSteps.item2")}
            </p>
          </Link>
        </div>
      </section>

      <DocNavigation />
    </div>
  );
}

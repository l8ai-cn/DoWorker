"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocStepHeader } from "@/components/docs/DocStepHeader";

export default function McpAndSkillsTutorialPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-2">
        {t("docs.tutorials.mcpSkills.title")}
      </h1>
      <p className="text-sm text-muted-foreground mb-8">
        {t("docs.tutorials.mcpSkills.difficulty")}
      </p>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.tutorials.mcpSkills.description")}
      </p>

      {/* What Are MCP Tools? */}
      <section className="mb-8">
        <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-6">
          <h2 className="text-xl font-semibold mb-4 text-foreground">
            {t("docs.tutorials.mcpSkills.whatIsMcp.title")}
          </h2>
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.mcpSkills.whatIsMcp.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.mcpSkills.whatIsMcp.item1")}</li>
            <li>{t("docs.tutorials.mcpSkills.whatIsMcp.item2")}</li>
            <li>{t("docs.tutorials.mcpSkills.whatIsMcp.item3")}</li>
            <li>{t("docs.tutorials.mcpSkills.whatIsMcp.item4")}</li>
            <li>{t("docs.tutorials.mcpSkills.whatIsMcp.item5")}</li>
            <li>{t("docs.tutorials.mcpSkills.whatIsMcp.item6")}</li>
            <li>{t("docs.tutorials.mcpSkills.whatIsMcp.item7")}</li>
          </ul>
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.mcpSkills.whatIsMcp.autoNote")}
          </div>
        </div>
      </section>

      {/* Step 1: Built-in vs Custom */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={1} titleKey="docs.tutorials.mcpSkills.step1.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.mcpSkills.step1.description")}
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
              <h4 className="font-medium mb-2">
                {t("docs.tutorials.mcpSkills.step1.builtinTitle")}
              </h4>
              <p className="text-sm text-muted-foreground">
                {t("docs.tutorials.mcpSkills.step1.builtinDesc")}
              </p>
            </div>
            <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
              <h4 className="font-medium mb-2">
                {t("docs.tutorials.mcpSkills.step1.customTitle")}
              </h4>
              <p className="text-sm text-muted-foreground">
                {t("docs.tutorials.mcpSkills.step1.customDesc")}
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Step 2: Install Custom MCP Server */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={2} titleKey="docs.tutorials.mcpSkills.step2.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.mcpSkills.step2.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.mcpSkills.step2.item1")}</li>
            <li>{t("docs.tutorials.mcpSkills.step2.item2")}</li>
            <li>{t("docs.tutorials.mcpSkills.step2.item3")}</li>
            <li>{t("docs.tutorials.mcpSkills.step2.item4")}</li>
            <li>{t("docs.tutorials.mcpSkills.step2.item5")}</li>
            <li>{t("docs.tutorials.mcpSkills.step2.item6")}</li>
            <li>{t("docs.tutorials.mcpSkills.step2.item7")}</li>
          </ol>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.mcpSkills.step2.tip")}
          </div>
        </div>
      </section>

      {/* Step 3: Verify */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={3} titleKey="docs.tutorials.mcpSkills.step3.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.mcpSkills.step3.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.mcpSkills.step3.item1")}</li>
            <li>{t("docs.tutorials.mcpSkills.step3.item2")}</li>
            <li>{t("docs.tutorials.mcpSkills.step3.item3")}</li>
            <li>{t("docs.tutorials.mcpSkills.step3.item4")}</li>
          </ol>
        </div>
      </section>

      {/* What Are Skills? */}
      <section className="mb-8">
        <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-6">
          <h2 className="text-xl font-semibold mb-4 text-foreground">
            {t("docs.tutorials.mcpSkills.whatAreSkills.title")}
          </h2>
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.mcpSkills.whatAreSkills.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2">
            <li>{t("docs.tutorials.mcpSkills.whatAreSkills.item1")}</li>
            <li>{t("docs.tutorials.mcpSkills.whatAreSkills.item2")}</li>
          </ul>
        </div>
      </section>

      {/* Step 4: Built-in Skills */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={4} titleKey="docs.tutorials.mcpSkills.step4.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.mcpSkills.step4.description")}
          </p>
          <div className="space-y-4">
            <div className="surface-card p-4">
              <h3 className="font-medium mb-2">
                {t("docs.tutorials.mcpSkills.step4.channelTitle")}
              </h3>
              <p className="text-sm text-muted-foreground">
                {t("docs.tutorials.mcpSkills.step4.channelDesc")}
              </p>
            </div>
            <div className="surface-card p-4">
              <h3 className="font-medium mb-2">
                {t("docs.tutorials.mcpSkills.step4.delegateTitle")}
              </h3>
              <p className="text-sm text-muted-foreground">
                {t("docs.tutorials.mcpSkills.step4.delegateDesc")}
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Step 5: Install Custom Skills */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={5} titleKey="docs.tutorials.mcpSkills.step5.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.mcpSkills.step5.description")}
          </p>
          <ol className="list-decimal list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.mcpSkills.step5.item1")}</li>
            <li>{t("docs.tutorials.mcpSkills.step5.item2")}</li>
            <li>{t("docs.tutorials.mcpSkills.step5.item3")}</li>
            <li>{t("docs.tutorials.mcpSkills.step5.item4")}</li>
          </ol>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.mcpSkills.step5.tip")}
          </div>
        </div>
      </section>

      {/* Step 6: Per-Repository Config */}
      <section className="mb-8">
        <div className="surface-card p-6">
          <DocStepHeader step={6} titleKey="docs.tutorials.mcpSkills.step6.title" />
          <p className="text-muted-foreground mb-4">
            {t("docs.tutorials.mcpSkills.step6.description")}
          </p>
          <ul className="list-disc list-inside text-muted-foreground space-y-2 mb-4">
            <li>{t("docs.tutorials.mcpSkills.step6.item1")}</li>
            <li>{t("docs.tutorials.mcpSkills.step6.item2")}</li>
            <li>{t("docs.tutorials.mcpSkills.step6.item3")}</li>
          </ul>
          <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 text-sm text-muted-foreground">
            {t("docs.tutorials.mcpSkills.step6.tip")}
          </div>
        </div>
      </section>

      {/* Next Steps */}
      <section className="mb-8">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.tutorials.mcpSkills.nextSteps.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.tutorials.mcpSkills.nextSteps.description")}
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Link
            href="/docs/tutorials/first-pod"
            className="surface-card-interactive p-4 block"
          >
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.mcpSkills.nextSteps.item1")}
            </p>
          </Link>
          <Link
            href="/docs/tutorials/multi-agent-collaboration"
            className="surface-card-interactive p-4 block"
          >
            <p className="text-sm text-muted-foreground">
              {t("docs.tutorials.mcpSkills.nextSteps.item2")}
            </p>
          </Link>
        </div>
      </section>

      <DocNavigation />
    </div>
  );
}

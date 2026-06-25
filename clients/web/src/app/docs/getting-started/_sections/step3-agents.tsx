"use client";

import { useTranslations } from "next-intl";
import { DocStepHeader } from "@/components/docs/DocStepHeader";
import { LinkInText } from "../_components/link-in-text";

function AgentCliBlock({
  titleKey,
  installKey,
  envKey,
  hintKey,
  href,
  linkLabel,
}: {
  titleKey: string;
  installKey: string;
  envKey: string;
  hintKey: string;
  href: string;
  linkLabel: string;
}) {
  const t = useTranslations();

  return (
    <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
      <h4 className="font-medium mb-2">{t(titleKey)}</h4>
      <div className="font-mono text-sm overflow-x-auto space-y-1">
        <pre className="text-success">{t(installKey)}</pre>
        <pre className="text-success">{t(envKey)}</pre>
      </div>
      <p className="text-sm text-muted-foreground mt-2">
        {t(hintKey)}{" "}
        <a href={href} className="text-primary hover:underline" target="_blank" rel="noopener noreferrer">
          {linkLabel}
        </a>
      </p>
    </div>
  );
}

export function Step3AgentsSection() {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <div className="surface-card p-6">
        <DocStepHeader step={3} titleKey="docs.gettingStarted.step3.title" />
        <p className="text-muted-foreground mb-4">{t("docs.gettingStarted.step3.description")}</p>
        <div className="space-y-4">
          <AgentCliBlock
            titleKey="docs.gettingStarted.step3.claudeCode"
            installKey="docs.gettingStarted.step3.claudeCodeInstall"
            envKey="docs.gettingStarted.step3.claudeCodeEnv"
            hintKey="docs.gettingStarted.step3.claudeCodeHint"
            href="https://console.anthropic.com"
            linkLabel="console.anthropic.com"
          />
          <AgentCliBlock
            titleKey="docs.gettingStarted.step3.codexCli"
            installKey="docs.gettingStarted.step3.codexCliInstall"
            envKey="docs.gettingStarted.step3.codexCliEnv"
            hintKey="docs.gettingStarted.step3.codexCliHint"
            href="https://platform.openai.com"
            linkLabel="platform.openai.com"
          />
          <AgentCliBlock
            titleKey="docs.gettingStarted.step3.geminiCli"
            installKey="docs.gettingStarted.step3.geminiCliInstall"
            envKey="docs.gettingStarted.step3.geminiCliEnv"
            hintKey="docs.gettingStarted.step3.geminiCliHint"
            href="https://aistudio.google.com"
            linkLabel="aistudio.google.com"
          />
        </div>
        <div className="rounded-lg bg-surface-muted/50 shadow-[var(--shadow-soft)] ring-1 ring-border/15 p-4 mt-4 text-sm text-muted-foreground">
          {t("docs.gettingStarted.step3.tip")}
        </div>
        <p className="text-sm text-muted-foreground mt-4">
          <LinkInText
            raw={t.raw("docs.gettingStarted.step3.seeSetup")}
            linkHref="/docs/tutorials/mcp-and-skills"
            linkLabel={t("docs.nav.tutorialMcpSkills")}
          />
        </p>
      </div>
    </section>
  );
}

"use client";

import { useTranslations } from "next-intl";
import { DocStepHeader } from "@/components/docs/DocStepHeader";
import { LinkInText } from "../_components/link-in-text";

export function Step4GitSection() {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <div className="surface-card p-6">
        <DocStepHeader step={4} titleKey="docs.gettingStarted.step4.title" />
        <p className="text-muted-foreground mb-4">{t("docs.gettingStarted.step4.description")}</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
            <h4 className="font-medium mb-2">{t("docs.gettingStarted.step4.github")}</h4>
            <p className="text-sm text-muted-foreground">
              {t("docs.gettingStarted.step4.githubDesc")}{" "}
              <a
                href="https://github.com/settings/tokens"
                className="text-primary hover:underline"
                target="_blank"
                rel="noopener noreferrer"
              >
                {t("docs.gettingStarted.step4.githubTokenUrl")}
              </a>
            </p>
          </div>
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
            <h4 className="font-medium mb-2">{t("docs.gettingStarted.step4.gitlab")}</h4>
            <p className="text-sm text-muted-foreground">{t("docs.gettingStarted.step4.gitlabDesc")}</p>
          </div>
        </div>
        <p className="text-sm text-muted-foreground">
          <LinkInText
            raw={t.raw("docs.gettingStarted.step4.seeGuide")}
            linkHref="/docs/tutorials/git-setup"
            linkLabel={t("docs.nav.tutorialGitSetup")}
          />
        </p>
      </div>
    </section>
  );
}

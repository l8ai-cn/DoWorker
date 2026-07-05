"use client";

import { useTranslations } from "next-intl";
import { DocStepHeader } from "@/components/docs/DocStepHeader";
import { LinkInText } from "../_components/link-in-text";

interface Step2RunnerSectionProps {
  serverUrl: string;
}

export function Step2RunnerSection({ serverUrl }: Step2RunnerSectionProps) {
  const t = useTranslations();

  return (
    <section className="mb-8">
      <div className="surface-card p-6">
        <DocStepHeader step={2} titleKey="docs.gettingStarted.step2.title" />
        <p className="text-muted-foreground mb-4">{t("docs.gettingStarted.step2.description")}</p>
        <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm overflow-x-auto mb-4">
          <pre className="text-success">{`# Download and install the runner
curl -fsSL ${serverUrl}/install.sh | sh`}</pre>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
            <h4 className="font-medium mb-2">{t("docs.gettingStarted.step2.methodToken")}</h4>
            <div className="font-mono text-sm overflow-x-auto">
              <pre className="text-success">{`do-worker-runner register \\
  --server ${serverUrl} \\
  --token <YOUR_TOKEN>
do-worker-runner run`}</pre>
            </div>
          </div>
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
            <h4 className="font-medium mb-2">{t("docs.gettingStarted.step2.methodLogin")}</h4>
            <div className="font-mono text-sm overflow-x-auto">
              <pre className="text-success">{`do-worker-runner login
do-worker-runner run`}</pre>
            </div>
          </div>
        </div>
        <p className="text-sm text-muted-foreground">
          <LinkInText
            raw={t.raw("docs.gettingStarted.step2.seeSetup")}
            linkHref="/docs/runners/setup"
            linkLabel={t("docs.nav.runnerSetup")}
          />
        </p>
      </div>
    </section>
  );
}

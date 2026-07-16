"use client";

import { useServerUrl } from "@/hooks/useServerUrl";
import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { Step1AccountSection } from "./_sections/step1-account";
import { Step2RunnerSection } from "./_sections/step2-runner";
import { Step3ModelResourceSection } from "./_sections/step3-model-resource";
import { Step4GitSection } from "./_sections/step4-git";
import { Step5PodSection } from "./_sections/step5-pod";
import { Step6InteractSection } from "./_sections/step6-interact";
import { TryItSection } from "./_sections/try-it-section";
import { NextStepsSection } from "./_sections/next-steps-section";

export default function GettingStartedPage() {
  const serverUrl = useServerUrl();
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">{t("docs.gettingStarted.title")}</h1>
      <p className="text-muted-foreground leading-relaxed mb-8">{t("docs.gettingStarted.description")}</p>

      <Step1AccountSection />
      <Step2RunnerSection serverUrl={serverUrl} />
      <Step3ModelResourceSection />
      <Step4GitSection />
      <Step5PodSection />
      <Step6InteractSection />
      <TryItSection />
      <NextStepsSection />

      <DocNavigation />
    </div>
  );
}

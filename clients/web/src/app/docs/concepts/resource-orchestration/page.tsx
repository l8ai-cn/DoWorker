"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { ResourceContractSection } from "./resource-contract-section";
import { ResourceEditorSection } from "./resource-editor-section";

export default function ResourceOrchestrationPage() {
  const t = useTranslations("resourceOrchestration");

  return (
    <div>
      <h1 className="mb-5 text-4xl font-bold text-foreground">{t("title")}</h1>
      <p className="mb-10 max-w-3xl leading-relaxed text-muted-foreground">
        {t("description")}
      </p>
      <ResourceContractSection />
      <ResourceEditorSection />
      <DocNavigation />
    </div>
  );
}

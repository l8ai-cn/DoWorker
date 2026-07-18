"use client";

import { useTranslations } from "next-intl";
import { editorSteps, securityItems } from "./resource-orchestration-data";

const yamlExample = `apiVersion: agentsmesh.io/v1alpha1
kind: Expert
metadata:
  name: delivery-reviewer
  namespace: acme
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
  promptRef:
    kind: Prompt
    name: delivery-review`;

export function ResourceEditorSection() {
  const t = useTranslations("resourceOrchestration");

  return (
    <>
      <section className="mb-12">
        <h2 className="text-2xl font-semibold text-foreground">
          {t("editorTitle")}
        </h2>
        <p className="mt-3 max-w-3xl leading-relaxed text-muted-foreground">
          {t("editorDescription")}
        </p>
        <ol className="mt-5 space-y-3">
          {editorSteps.map((step, index) => (
            <li key={step} className="flex gap-3 text-muted-foreground">
              <span className="font-semibold text-primary">{index + 1}.</span>
              <span className="min-w-0 leading-relaxed">{t(step)}</span>
            </li>
          ))}
        </ol>
        <h3
          id="yaml"
          className="mt-8 scroll-mt-24 text-lg font-semibold text-foreground"
        >
          {t("yamlTitle")}
        </h3>
        <p className="mt-2 max-w-3xl text-sm leading-relaxed text-muted-foreground">
          {t("yamlLimits")}
        </p>
        <pre className="mt-4 overflow-x-auto rounded-lg bg-surface-muted p-4 text-sm text-foreground">
          <code>{yamlExample}</code>
        </pre>
      </section>

      <section className="mb-12">
        <h2 className="text-2xl font-semibold text-foreground">
          {t("securityTitle")}
        </h2>
        <p className="mt-3 max-w-3xl leading-relaxed text-muted-foreground">
          {t("securityDescription")}
        </p>
        <ul className="mt-5 space-y-3 text-muted-foreground">
          {securityItems.map((item) => (
            <li key={item} className="leading-relaxed">
              {t(item)}
            </li>
          ))}
        </ul>
      </section>

      <section className="mb-12 border-l-4 border-primary bg-primary/5 px-5 py-4">
        <h2 className="text-lg font-semibold text-foreground">
          {t("boundaryTitle")}
        </h2>
        <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
          {t("goalLoopBoundary")}
        </p>
        <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
          {t("reconciliationBoundary")}
        </p>
      </section>
    </>
  );
}

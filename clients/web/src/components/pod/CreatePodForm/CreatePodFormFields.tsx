"use client";

import { useEffect, useMemo } from "react";
import { AlertMessage } from "@/components/ui/alert-message";
import { Button } from "@/components/ui/button";
import type { WorkerCreateController } from "../hooks/workerCreateController";
import { workerPreflightHasBlockingIssues } from "../hooks/workerCreateController";
import { WorkerCreateStepper } from "./WorkerCreateStepper";
import { WorkerCreateStepNav } from "./WorkerCreateStepNav";
import { WorkerPreflightStep } from "./WorkerPreflightStep";
import { WorkerRuntimeStep } from "./WorkerRuntimeStep";
import { WorkerTypeConfigStep } from "./WorkerTypeConfigStep";
import { WorkerWorkspaceStep } from "./WorkerWorkspaceStep";

interface CreatePodFormFieldsProps {
  controller: WorkerCreateController;
  initialWizardStep?: 1 | 2 | 3 | 4;
  promptPlaceholder?: string;
  onCancel?: () => void;
  t: (key: string) => string;
}

export function CreatePodFormFields({
  controller,
  initialWizardStep = 1,
  promptPlaceholder,
  onCancel,
  t,
}: CreatePodFormFieldsProps) {
  const { state, validity } = controller;
  const step = state.step;
  const steps = useMemo(
    () => [
      stepDefinition(1, t("workerCreate.steps.runtime"), validity.runtime, true),
      stepDefinition(
        2,
        t("workerCreate.steps.typeConfig"),
        validity.typeConfig,
        validity.accessible(2),
      ),
      stepDefinition(
        3,
        t("workerCreate.steps.workspace"),
        validity.workspace,
        validity.accessible(3),
      ),
      stepDefinition(
        4,
        t("workerCreate.steps.preflight"),
        preflightReady(controller),
        validity.accessible(4),
      ),
    ],
    [controller, t, validity],
  );

  useEffect(() => {
    if (
      initialWizardStep !== 1 &&
      state.step === 1 &&
      validity.accessible(initialWizardStep)
    ) {
      void controller.goToStep(initialWizardStep);
    }
  }, [controller, initialWizardStep, state.step, validity]);

  const nextStep = step < 4 ? ((step + 1) as 2 | 3 | 4) : 4;
  const canNext = step < 4 && validity.accessible(nextStep);

  return (
    <div className="space-y-5">
      <div className="block md:hidden">
        <WorkerCreateStepper
          steps={steps}
          current={step}
          onChange={(next) => void controller.goToStep(next)}
        />
      </div>
      <div className="flex flex-col gap-6 md:flex-row md:items-start">
        <div className="hidden w-56 shrink-0 md:sticky md:top-6 md:block">
          <WorkerCreateStepper
            steps={steps}
            current={step}
            orientation="vertical"
            onChange={(next) => void controller.goToStep(next)}
          />
        </div>
        <div className="min-w-0 flex-1">
          <header className="mb-5 border-b border-border pb-4">
            <h2 className="text-lg font-semibold">{t(stepTitle(step))}</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {t(stepDescription(step))}
            </p>
          </header>

          {step === 1 && (
            <WorkerRuntimeStep
              draft={state.draft}
              options={controller.options}
              modelResources={controller.modelResources}
              onPatch={controller.patchDraft}
              onWorkerTypeChange={controller.changeWorkerType}
              t={t}
            />
          )}
          {step === 2 && (
            <WorkerTypeConfigStep
              draft={state.draft}
              options={controller.options}
              credentialBundles={controller.credentialBundles}
              onPatch={controller.patchDraft}
              t={t}
            />
          )}
          {step === 3 && (
            <WorkerWorkspaceStep
              controller={controller}
              promptPlaceholder={promptPlaceholder}
              t={t}
            />
          )}
          {step === 4 && (
            <WorkerPreflightStep
              preflight={state.preflight}
              creating={state.create.status === "loading"}
              onRetry={() => void controller.runPreflight()}
              onCreate={() => void controller.createWorker()}
              t={t}
            />
          )}

          <AsyncErrors controller={controller} />
          <WorkerCreateStepNav
            step={step}
            canNext={canNext}
            onBack={() => void controller.goToStep(previousStep(step))}
            onNext={() => void controller.goToStep(nextStep)}
            t={t}
          />
          {onCancel && (
            <div className="mt-3 flex justify-start">
              <Button type="button" variant="ghost" size="sm" onClick={onCancel}>
                {t("ide.createPod.cancel")}
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function AsyncErrors({ controller }: { controller: WorkerCreateController }) {
  const { fill, create } = controller.state;
  const fillIssues = fill.status === "ready" ? fill.data.issues : [];
  return (
    <div className="mt-5 space-y-3">
      {fill.status === "error" && <AlertMessage type="error" message={fill.error} />}
      {create.status === "error" && <AlertMessage type="error" message={create.error} />}
      {fillIssues.length > 0 && (
        <AlertMessage
          type={fillIssues.some((issue) => issue.severity === "blocking") ? "error" : "warning"}
          message={fillIssues.map((issue) => issue.message).join(" ")}
        />
      )}
    </div>
  );
}

function preflightReady(controller: WorkerCreateController): boolean {
  return (
    controller.state.preflight.status === "ready" &&
    !workerPreflightHasBlockingIssues(controller.state.preflight.data) &&
    Boolean(controller.state.preflight.data.resolved_spec_json?.trim())
  );
}

function previousStep(step: 1 | 2 | 3 | 4): 1 | 2 | 3 {
  return ({ 1: 1, 2: 1, 3: 2, 4: 3 } as const)[step];
}

function stepDefinition(
  id: 1 | 2 | 3 | 4,
  label: string,
  complete: boolean,
  accessible: boolean,
) {
  return { id, label, complete, accessible };
}

function stepTitle(step: 1 | 2 | 3 | 4): string {
  return `workerCreate.stepContent.${step}.title`;
}

function stepDescription(step: 1 | 2 | 3 | 4): string {
  return `workerCreate.stepContent.${step}.description`;
}

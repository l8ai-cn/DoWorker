"use client";

import { useMemo, useState } from "react";
import type { AgentData, ConfigField, RepositoryData, RunnerData } from "@/lib/api";
import type { CreatePodFormState } from "../hooks";
import { AdvancedFormSection } from "./AdvancedFormSection";
import { WorkerCreateStepper, type WorkerCreateStepId } from "./WorkerCreateStepper";
import { WorkerCreateStepNav } from "./WorkerCreateStepNav";
import { WorkerMoreOptionsSection } from "./WorkerMoreOptionsSection";
import { WorkerCreateModeToggle } from "./WorkerCreateModeToggle";
import { WorkerSourceModePanel } from "./WorkerSourceModePanel";
import {
  WorkerStepRuntimePanel,
  WorkerStepAgentPanel,
} from "./WorkerStepPanels";
import { WorkerStepCapabilitiesPanel } from "./WorkerStepCapabilitiesPanel";
import { step1Summary, step2Summary, step3Summary } from "./workerCreateStepSummaries";

interface CreatePodFormFieldsProps {
  form: CreatePodFormState;
  agents: AgentData[];
  runners: RunnerData[];
  repositories: RepositoryData[];
  selectedRunner: RunnerData | null;
  setSelectedRunnerId: (id: number | null) => void;
  configFields: ConfigField[];
  loadingConfig: boolean;
  configValues: Record<string, unknown>;
  handleConfigChange: (key: string, value: unknown) => void;
  hasOnlineRunners: boolean;
  promptPlaceholder?: string;
  showPerpetual: boolean;
  initialWizardStep?: 1 | 2 | 3;
  initialExpertSlug?: string;
  t: (key: string) => string;
  actions?: React.ReactNode;
}

export function CreatePodFormFields(props: CreatePodFormFieldsProps) {
  const { form, showPerpetual, t, initialWizardStep = 1, repositories, actions } = props;
  const [step, setStep] = useState<WorkerCreateStepId>(initialWizardStep);
  const agentReady = Boolean(form.selectedAgent);

  const selectedRepoSlug = useMemo(
    () => repositories.find((r) => r.id === form.selectedRepository)?.slug,
    [repositories, form.selectedRepository],
  );

  const stepDefs = useMemo(
    () => [
      {
        id: 1 as const,
        label: t("ide.createPod.stepperRuntime"),
        summary: step1Summary(
          form.selectedAgent,
          form.interactionMode,
          form.perpetual,
          form.destroyPolicy,
          selectedRepoSlug,
          form.selectedBranch,
          t,
        ),
        complete: agentReady,
        accessible: true,
      },
      {
        id: 2 as const,
        label: t("ide.createPod.stepperCapabilities"),
        summary: step2Summary(
          form.selectedKnowledgeMounts.length,
          form.selectedSkillSlugs.length,
          t,
        ),
        complete:
          form.selectedKnowledgeMounts.length > 0 || form.selectedSkillSlugs.length > 0,
        accessible: agentReady,
      },
      {
        id: 3 as const,
        label: t("ide.createPod.stepperAgent"),
        summary: step3Summary(
          form.rawLayerMode,
          Boolean(form.agentfileLayer?.trim()),
          t,
        ),
        complete: Boolean(form.agentfileLayer?.trim()) || form.rawLayerMode,
        accessible: agentReady,
      },
    ],
    [form, agentReady, selectedRepoSlug, t],
  );

  const panelProps = { ...props, showPerpetual };
  const sourceMode = form.rawLayerMode;

  const canNext =
    step === 1 ? agentReady : step === 2 ? true : false;

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <WorkerCreateModeToggle
          sourceMode={sourceMode}
          onChange={form.setRawLayerMode}
          t={t}
        />
      </div>

      <div className={sourceMode ? undefined : "flex flex-col gap-6 md:flex-row md:items-start"}>
        {!sourceMode && (
          <div className="w-full md:w-64 md:shrink-0">
            <div className="block md:hidden">
              <WorkerCreateStepper
                steps={stepDefs}
                current={step}
                onChange={setStep}
                orientation="horizontal"
              />
            </div>
            <div className="hidden md:block md:sticky md:top-6">
              <WorkerCreateStepper
                steps={stepDefs}
                current={step}
                onChange={setStep}
                orientation="vertical"
              />
            </div>
          </div>
        )}

        <div className="flex-1 min-w-0 space-y-4">
          <div className="rounded-lg border border-border bg-card p-4 shadow-xs md:p-5">
            {sourceMode ? (
              <WorkerSourceModePanel {...panelProps} />
            ) : (
              <>
                {step === 1 && <WorkerStepRuntimePanel {...panelProps} />}
                {step === 2 && agentReady && <WorkerStepCapabilitiesPanel {...panelProps} />}
                {step === 3 && agentReady && <WorkerStepAgentPanel {...panelProps} />}

                <WorkerCreateStepNav
                  step={step}
                  canNext={canNext}
                  onBack={() => setStep((s) => (s > 1 ? ((s - 1) as WorkerCreateStepId) : s))}
                  onNext={() => setStep((s) => (s < 3 ? ((s + 1) as WorkerCreateStepId) : s))}
                  t={t}
                />
              </>
            )}
          </div>

          {agentReady && (
            <WorkerMoreOptionsSection t={t}>
              <AdvancedFormSection form={form} />
            </WorkerMoreOptionsSection>
          )}

          {form.error && (
            <div
              role="alert"
              aria-live="assertive"
              className="rounded-md border border-destructive/30 bg-destructive/10 p-3"
            >
              <p className="text-sm text-destructive">{form.error}</p>
            </div>
          )}

          {actions}
        </div>
      </div>
    </div>
  );
}

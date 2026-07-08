"use client";

import { useMemo, useState } from "react";
import type { AgentData, ConfigField, RepositoryData, RunnerData } from "@/lib/api";
import type { CreatePodFormState } from "../hooks";
import { AdvancedFormSection } from "./AdvancedFormSection";
import { WorkerCreateStepper, type WorkerCreateStepId } from "./WorkerCreateStepper";
import { WorkerCreateStepNav } from "./WorkerCreateStepNav";
import { WorkerMoreOptionsSection } from "./WorkerMoreOptionsSection";
import {
  WorkerStepRuntimePanel,
  WorkerStepCapabilitiesPanel,
  WorkerStepAgentPanel,
} from "./WorkerStepPanels";
import { step1Summary, step2Summary, step3Summary } from "./workerCreateStepSummaries";
import { ExpertPickerSection } from "@/components/experts/ExpertPickerSection";

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
}

export function CreatePodFormFields(props: CreatePodFormFieldsProps) {
  const { form, showPerpetual, t, initialWizardStep = 1, initialExpertSlug, repositories } = props;
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

  const canNext =
    step === 1 ? agentReady : step === 2 ? true : false;

  return (
    <div className="space-y-2">
      {showPerpetual && (
        <ExpertPickerSection
          form={form}
          setSelectedRunnerId={props.setSelectedRunnerId}
          initialExpertSlug={initialExpertSlug}
        />
      )}

      <WorkerCreateStepper steps={stepDefs} current={step} onChange={setStep} />

      <div className="rounded-lg border border-border bg-card p-4 shadow-xs md:p-5">
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
      </div>

      {agentReady && (
        <WorkerMoreOptionsSection t={t}>
          <AdvancedFormSection
            form={form}
            configFields={props.configFields}
            loadingConfig={props.loadingConfig}
            configValues={props.configValues}
            handleConfigChange={props.handleConfigChange}
          />
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
    </div>
  );
}

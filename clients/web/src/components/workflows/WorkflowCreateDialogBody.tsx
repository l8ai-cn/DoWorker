"use client";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ResponsiveDialogBody } from "@/components/ui/responsive-dialog";
import { WorkerImageSelect } from "@/components/pod/CreatePodForm/WorkerImageSelect";
import { PromptInput } from "@/components/pod/CreatePodForm/PromptInput";
import type { AgentData, ConfigField, EnvBundleSummary, RepositoryData, RunnerData } from "@/lib/api";
import type { UseWorkflowFormResult } from "./useWorkflowForm";
import { WorkflowPodConfigSection } from "./WorkflowPodConfigSection";
import { WorkflowScheduleSection } from "./WorkflowScheduleSection";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";

interface Props {
  form: UseWorkflowFormResult;
  availableAgents: AgentData[];
  runners: RunnerData[];
  compatibleRunners: RunnerData[];
  repositories: RepositoryData[];
  envBundles: EnvBundleSummary[];
  modelResources: EffectiveResource[];
  selectedModelResourceId: number | null;
  onSelectModelResource: (id: number | null) => void;
  loadingModelResources: boolean;
  modelResourceError: string | null;
  modelResourceRequired: boolean;
  configFields: ConfigField[];
  configValues: Record<string, unknown>;
  loadingConfig: boolean;
  loadingBundles: boolean;
  onConfigChange: (key: string, value: unknown) => void;
  t: (key: string) => string;
  /** When true, render without ResponsiveDialogBody padding wrapper (inline panel). */
  embedded?: boolean;
}

export function WorkflowCreateDialogBody({
  form,
  availableAgents,
  runners,
  compatibleRunners,
  repositories,
  envBundles,
  modelResources,
  selectedModelResourceId,
  onSelectModelResource,
  loadingModelResources,
  modelResourceError,
  modelResourceRequired,
  configFields,
  configValues,
  loadingConfig,
  loadingBundles,
  onConfigChange,
  t,
  embedded = false,
}: Props) {
  const fields = (
    <>
      <div className="space-y-1.5">
        <Label>{t("workflows.name")}</Label>
        <Input value={form.name} onChange={(e) => form.setName(e.target.value)} placeholder="daily-code-review" />
      </div>

      <div className="space-y-1.5">
        <Label>{t("workflows.description")}</Label>
        <Input
          value={form.description}
          onChange={(e) => form.setDescription(e.target.value)}
          placeholder={t("workflows.descriptionPlaceholder")}
        />
      </div>

      <WorkerImageSelect
        images={availableAgents}
        selectedImageSlug={form.selectedAgentSlug}
        onSelect={form.setSelectedAgentSlug}
        hasOnlineClusters={runners.length > 0}
        t={t}
      />

      {form.selectedAgentSlug && (
        <>
          <PromptInput
            value={form.promptTemplate}
            onChange={form.setPromptTemplate}
            placeholder={t("workflows.promptPlaceholder")}
            t={t}
          />

          <WorkflowPodConfigSection
            agentSlug={form.selectedAgentSlug}
            runners={compatibleRunners}
            repositories={repositories}
            envBundles={envBundles}
            configFields={configFields}
            configValues={configValues}
            loadingConfig={loadingConfig}
            loadingBundles={loadingBundles}
            selectedRunnerId={form.selectedRunnerId}
            onSelectRunner={form.setSelectedRunnerId}
            modelResources={modelResources}
            selectedModelResourceId={selectedModelResourceId}
            onSelectModelResource={onSelectModelResource}
            loadingModelResources={loadingModelResources}
            modelResourceError={modelResourceError}
            modelResourceRequired={modelResourceRequired}
            selectedRuntimeBundleNames={form.selectedRuntimeBundleNames}
            onSelectRuntimeBundles={form.setSelectedRuntimeBundleNames}
            selectedRepositoryId={form.selectedRepositoryId}
            onSelectRepository={form.setSelectedRepositoryId}
            selectedBranch={form.selectedBranch}
            onChangeBranch={form.setSelectedBranch}
            onConfigChange={onConfigChange}
            t={t}
          />
        </>
      )}

      <WorkflowScheduleSection
        cronEnabled={form.cronEnabled}
        onCronEnabledChange={form.setCronEnabled}
        cronExpression={form.cronExpression}
        onCronExpressionChange={form.setCronExpression}
        executionMode={form.executionMode}
        onExecutionModeChange={form.setExecutionMode}
        sandboxStrategy={form.sandboxStrategy}
        onSandboxStrategyChange={form.setSandboxStrategy}
        concurrencyPolicy={form.concurrencyPolicy}
        onConcurrencyPolicyChange={form.setConcurrencyPolicy}
        timeoutMinutes={form.timeoutMinutes}
        onTimeoutMinutesChange={form.setTimeoutMinutes}
        maxConcurrentRuns={form.maxConcurrentRuns}
        onMaxConcurrentRunsChange={form.setMaxConcurrentRuns}
        maxRetainedRuns={form.maxRetainedRuns}
        onMaxRetainedRunsChange={form.setMaxRetainedRuns}
        sessionPersistence={form.sessionPersistence}
        onSessionPersistenceChange={form.setSessionPersistence}
        callbackUrl={form.callbackUrl}
        onCallbackUrlChange={form.setCallbackUrl}
        t={t}
      />
    </>
  );

  if (embedded) {
    return <div className="space-y-4">{fields}</div>;
  }
  return <ResponsiveDialogBody className="space-y-4">{fields}</ResponsiveDialogBody>;
}

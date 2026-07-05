"use client";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ResponsiveDialogBody } from "@/components/ui/responsive-dialog";
import { WorkerImageSelect } from "@/components/pod/CreatePodForm/WorkerImageSelect";
import { PromptInput } from "@/components/pod/CreatePodForm/PromptInput";
import type { AgentData, ConfigField, EnvBundleSummary, RepositoryData, RunnerData } from "@/lib/api";
import type { UseLoopFormResult } from "./useLoopForm";
import { LoopPodConfigSection } from "./LoopPodConfigSection";
import { LoopScheduleSection } from "./LoopScheduleSection";

interface Props {
  form: UseLoopFormResult;
  availableAgents: AgentData[];
  runners: RunnerData[];
  compatibleRunners: RunnerData[];
  repositories: RepositoryData[];
  envBundles: EnvBundleSummary[];
  configFields: ConfigField[];
  configValues: Record<string, unknown>;
  loadingConfig: boolean;
  loadingBundles: boolean;
  onConfigChange: (key: string, value: unknown) => void;
  t: (key: string) => string;
}

export function LoopCreateDialogBody({
  form,
  availableAgents,
  runners,
  compatibleRunners,
  repositories,
  envBundles,
  configFields,
  configValues,
  loadingConfig,
  loadingBundles,
  onConfigChange,
  t,
}: Props) {
  return (
    <ResponsiveDialogBody className="space-y-4">
      <div className="space-y-1.5">
        <Label>{t("loops.name")}</Label>
        <Input value={form.name} onChange={(e) => form.setName(e.target.value)} placeholder="daily-code-review" />
      </div>

      <div className="space-y-1.5">
        <Label>{t("loops.description")}</Label>
        <Input
          value={form.description}
          onChange={(e) => form.setDescription(e.target.value)}
          placeholder={t("loops.descriptionPlaceholder")}
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
            placeholder={t("loops.promptPlaceholder")}
            t={t}
          />

          <LoopPodConfigSection
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
            selectedCredentialName={form.selectedCredentialName}
            onSelectCredential={form.setSelectedCredentialName}
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

      <LoopScheduleSection
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
    </ResponsiveDialogBody>
  );
}

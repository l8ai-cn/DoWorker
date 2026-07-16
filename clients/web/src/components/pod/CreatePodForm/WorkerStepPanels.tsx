import type { AgentData, ConfigField, RepositoryData, RunnerData } from "@/lib/api";
import type { CreatePodFormState } from "../hooks";
import { RunnerSelect } from "./RunnerSelect";
import { WorkerImageSelect } from "./WorkerImageSelect";
import { InteractionModeToggle } from "./InteractionModeToggle";
import { AutomationLevelSelect } from "./AutomationLevelSelect";
import { PromptInput } from "./PromptInput";
import { WorkerDurationSection } from "./WorkerDurationSection";
import { BranchInput } from "./RepositorySelect";
import { WorkerRepositoryField } from "./WorkerRepositoryField";
import { WorkerAgentInstructionsSection } from "./WorkerAgentInstructionsSection";
import { WorkerCredentialModelSection } from "./WorkerCredentialModelSection";

export interface StepPanelsProps {
  form: CreatePodFormState;
  agents: AgentData[];
  runners: RunnerData[];
  repositories: RepositoryData[];
  selectedRunner: RunnerData | null;
  setSelectedRunnerId: (id: number | null) => void;
  configFields: ConfigField[];
  hasOnlineRunners: boolean;
  loadingConfig: boolean;
  configValues: Record<string, unknown>;
  handleConfigChange: (key: string, value: unknown) => void;
  promptPlaceholder?: string;
  showPerpetual: boolean;
  initialExpertSlug?: string;
  t: (key: string) => string;
}

export function WorkerStepRuntimePanel({
  form,
  agents,
  runners,
  selectedRunner,
  setSelectedRunnerId,
  hasOnlineRunners,
  configFields,
  loadingConfig,
  configValues,
  handleConfigChange,
  promptPlaceholder,
  showPerpetual,
  t,
}: StepPanelsProps) {
  return (
    <div className="space-y-5 animate-in fade-in duration-200">
      <p className="text-sm text-muted-foreground">{t("ide.createPod.stepRuntimePanelHint")}</p>
      {hasOnlineRunners && (
        <RunnerSelect
          runners={runners}
          selectedRunnerId={selectedRunner?.id ?? null}
          onSelect={setSelectedRunnerId}
          error={form.validationErrors.runner}
          t={t}
        />
      )}
      <WorkerImageSelect
        images={agents}
        selectedImageSlug={form.selectedAgent}
        onSelect={form.setSelectedAgent}
        hasOnlineClusters={hasOnlineRunners}
        error={form.validationErrors.agent}
        t={t}
      />
      <WorkerRepositoryField
        value={form.selectedRepository}
        onChange={form.setSelectedRepository}
      />
      {form.selectedRepository && (
        <BranchInput
          value={form.selectedBranch}
          onChange={form.setSelectedBranch}
          error={form.validationErrors.branch}
          t={t}
        />
      )}
      {form.selectedAgent && (
        <>
          <WorkerCredentialModelSection
            agentSlug={form.selectedAgentSlug}
            modelResources={form.modelResources}
            selectedModelResourceId={form.selectedModelResourceId}
            onSelectModelResource={form.setSelectedModelResourceId}
            loadingModelResources={form.loadingModelResources}
            modelResourceError={form.modelResourceError}
            modelResourceValidationError={form.validationErrors.modelResource}
            envBundles={form.envBundles}
            loadingBundles={form.loadingBundles}
            bundleLoadError={form.bundleLoadError}
            runtimeBundleValidationError={form.validationErrors.runtimeBundles}
            selectedRuntimeBundleNames={form.selectedRuntimeBundleNames}
            onSelectRuntimeBundles={form.setSelectedRuntimeBundleNames}
            configFields={configFields}
            configValues={configValues}
            loadingConfig={loadingConfig}
            onConfigChange={handleConfigChange}
            rawLayerMode={form.rawLayerMode}
            t={t}
          />
          <div className="space-y-4 border-t border-border pt-4">
            <AutomationLevelSelect
              value={form.automationLevel}
              onChange={form.setAutomationLevel}
              supportedModes={form.supportedModes}
              t={t}
            />
            {!form.rawLayerMode && form.automationLevel !== "autonomous" && (
              <InteractionModeToggle
                supportedModes={form.supportedModes}
                interactionMode={form.interactionMode}
                onModeChange={form.setInteractionMode}
              />
            )}
            {showPerpetual && <WorkerDurationSection form={form} t={t} />}
            <PromptInput
              value={form.prompt}
              onChange={form.setPrompt}
              placeholder={promptPlaceholder}
              t={t}
            />
          </div>
        </>
      )}
    </div>
  );
}

export function WorkerStepAgentPanel({ form, t }: StepPanelsProps) {
  return (
    <div className="animate-in fade-in duration-200">
      <WorkerAgentInstructionsSection generatedLayer={form.agentfileLayer} t={t} />
    </div>
  );
}

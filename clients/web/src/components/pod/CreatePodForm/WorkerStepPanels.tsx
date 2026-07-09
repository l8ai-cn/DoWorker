"use client";

import { BookOpen, Braces, KeyRound, Sparkles, Wallet } from "lucide-react";
import { CustomEnvSection } from "@/components/settings/CustomEnvSection";
import type { AgentData, ConfigField, RepositoryData, RunnerData } from "@/lib/api";
import type { CreatePodFormState } from "../hooks";
import { RunnerSelect } from "./RunnerSelect";
import { WorkerImageSelect } from "./WorkerImageSelect";
import { InteractionModeToggle } from "./InteractionModeToggle";
import { AutomationLevelSelect } from "./AutomationLevelSelect";
import { PromptInput } from "./PromptInput";
import { WorkerDurationSection } from "./WorkerDurationSection";
import { CapabilityConfigPanel } from "./CapabilityConfigPanel";
import { KnowledgeBaseMountSelect } from "./KnowledgeBaseMountSelect";
import { SkillMultiSelect } from "./SkillMultiSelect";
import { WorkerBudgetSection } from "./WorkerBudgetSection";
import { WorkerModelBindingSelect } from "./WorkerModelBindingSelect";
import { BranchInput } from "./RepositorySelect";
import { WorkerRepositoryField } from "./WorkerRepositoryField";
import { WorkerAgentInstructionsSection } from "./WorkerAgentInstructionsSection";
import { WorkerCredentialModelSection } from "./WorkerCredentialModelSection";

import { ExpertPickerSection } from "@/components/experts/ExpertPickerSection";

const EMPTY_DECLARED_KEYS: Set<string> = new Set();

interface StepPanelsProps {
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
            envBundles={form.envBundles}
            loadingBundles={form.loadingBundles}
            selectedCredentialName={form.selectedCredentialName}
            onSelectCredential={form.setSelectedCredentialName}
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

export function WorkerStepCapabilitiesPanel({
  form,
  showPerpetual,
  setSelectedRunnerId,
  initialExpertSlug,
  t,
}: StepPanelsProps) {
  return (
    <div className="space-y-4 animate-in fade-in duration-200">
      <p className="text-sm text-muted-foreground">{t("ide.createPod.stepCapabilitiesPanelHint")}</p>

      {showPerpetual && (
        <ExpertPickerSection
          form={form}
          setSelectedRunnerId={setSelectedRunnerId}
          initialExpertSlug={initialExpertSlug}
        />
      )}

      <CapabilityConfigPanel
        icon={BookOpen}
        title={t("ide.createPod.knowledgeConfigTitle")}
        description={t("ide.createPod.knowledgeConfigDescription")}
        testId="worker-knowledge-select"
      >
        <KnowledgeBaseMountSelect
          embedded
          selectedMounts={form.selectedKnowledgeMounts}
          onChange={form.setSelectedKnowledgeMounts}
        />
      </CapabilityConfigPanel>
      <CapabilityConfigPanel
        icon={Sparkles}
        title={t("ide.createPod.skillConfigTitle")}
        description={t("ide.createPod.skillConfigDescription")}
        testId="worker-skills-select"
      >
        <div className="space-y-3">
          {!form.selectedRepository && (
            <p className="text-sm text-muted-foreground">
              {t("ide.createPod.skillsRequireRepositoryHint")}
            </p>
          )}
          <SkillMultiSelect
            embedded
            skills={form.repoSkills}
            selectedSlugs={form.selectedSkillSlugs}
            onChange={form.setSelectedSkillSlugs}
            loading={form.loadingSkills}
            repositorySelected={Boolean(form.selectedRepository)}
            t={t}
          />
        </div>
      </CapabilityConfigPanel>
      <CapabilityConfigPanel
        icon={KeyRound}
        title={t("ide.createPod.modelBindingTitle")}
        description={t("ide.createPod.modelBindingDescription")}
        testId="worker-model-binding-select"
      >
        <WorkerModelBindingSelect
          selectedVirtualKeyId={form.selectedVirtualKeyId}
          onSelect={form.setSelectedVirtualKeyId}
          t={t}
        />
      </CapabilityConfigPanel>
      <CapabilityConfigPanel
        icon={Wallet}
        title={t("ide.createPod.budgetConfigTitle")}
        description={t("ide.createPod.budgetConfigDescription")}
        testId="worker-budget-select"
      >
        <WorkerBudgetSection
          tokenBudget={form.tokenBudget}
          onChange={form.setTokenBudget}
          t={t}
        />
      </CapabilityConfigPanel>
      <CapabilityConfigPanel
        icon={Braces}
        title={t("ide.createPod.customEnvTitle")}
        description={t("ide.createPod.customEnvDescription")}
        testId="worker-custom-env"
      >
        <CustomEnvSection
          entries={form.customEnv}
          declaredKeys={EMPTY_DECLARED_KEYS}
          onChange={form.setCustomEnv}
          isEditing={false}
          valueType="text"
          t={t}
        />
        <p className="mt-3 text-xs text-muted-foreground">
          {t("ide.createPod.customEnvPlaintextWarning")}
        </p>
      </CapabilityConfigPanel>
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

"use client";

import type { AgentData, ConfigField, RepositoryData, RunnerData } from "@/lib/api";
import type { CreatePodFormState } from "../hooks";
import { AdvancedFormSection } from "./AdvancedFormSection";
import { CreatePodModeSection } from "./CreatePodModeSection";
import { AgentSelect } from "./AgentSelect";
import { InteractionModeToggle } from "./InteractionModeToggle";
import { PromptInput } from "./PromptInput";

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
  t: (key: string) => string;
}

export function CreatePodFormFields({
  form,
  agents,
  runners,
  repositories,
  selectedRunner,
  setSelectedRunnerId,
  configFields,
  loadingConfig,
  configValues,
  handleConfigChange,
  hasOnlineRunners,
  promptPlaceholder,
  showPerpetual,
  t,
}: CreatePodFormFieldsProps) {
  return (
    <div className="space-y-4">
      <AgentSelect
        agents={agents}
        selectedAgentSlug={form.selectedAgent}
        onSelect={form.setSelectedAgent}
        hasOnlineRunners={hasOnlineRunners}
        error={form.validationErrors.agent}
        t={t}
      />

      {form.selectedAgent && !form.rawLayerMode && (
        <InteractionModeToggle
          supportedModes={form.supportedModes}
          interactionMode={form.interactionMode}
          onModeChange={form.setInteractionMode}
        />
      )}

      {form.selectedAgent && (
        <PromptInput
          value={form.prompt}
          onChange={form.setPrompt}
          placeholder={promptPlaceholder}
          t={t}
        />
      )}

      {form.selectedAgent && (
        <CreatePodModeSection t={t}>
          <AdvancedFormSection
            form={form}
            runners={runners}
            repositories={repositories}
            selectedRunner={selectedRunner}
            setSelectedRunnerId={setSelectedRunnerId}
            configFields={configFields}
            loadingConfig={loadingConfig}
            configValues={configValues}
            handleConfigChange={handleConfigChange}
            showPerpetual={showPerpetual}
          />
        </CreatePodModeSection>
      )}

      {form.error && (
        <div
          role="alert"
          aria-live="assertive"
          className="bg-destructive/10 border border-destructive/30 rounded-md p-3"
        >
          <p className="text-sm text-destructive">{form.error}</p>
        </div>
      )}
    </div>
  );
}

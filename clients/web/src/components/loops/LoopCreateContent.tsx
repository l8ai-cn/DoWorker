"use client";

import { useEffect, useMemo, useState } from "react";
import { Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { usePodCreationData } from "@/components/pod/hooks";
import { requiresModelResource, useWorkerModelResources } from "@/components/pod/hooks/useWorkerModelResources";
import { useConfigOptions } from "@/components/ide/hooks";
import {
  hasRunnerForAgent,
  runnersSupportingAgent,
} from "@/lib/runner-agent-capabilities";
import type { LoopData } from "@/lib/viewModels/loop";
import { LoopNlCreate } from "./LoopNlCreate";
import { useLoopForm } from "./useLoopForm";
import { useLoopEnvBundles } from "./useLoopEnvBundles";
import { LoopCreateDialogBody } from "./LoopCreateDialogBody";

interface LoopCreateContentProps {
  editLoop?: LoopData;
  onCreated: (createdLoop?: LoopData) => void;
  onCancel?: () => void;
  /** When false (edit mode), hide the AI guide section. */
  showAiSection?: boolean;
}

/**
 * Shared Loop create UI: AI guide on top, manual form below — same linkage
 * pattern as CreateWorkerPageContent (NlWorkerCreate + CreatePodForm).
 */
export function LoopCreateContent({
  editLoop,
  onCreated,
  onCancel,
  showAiSection = !editLoop,
}: LoopCreateContentProps) {
  const t = useTranslations();
  const [wizardIdea, setWizardIdea] = useState<string | undefined>();
  const form = useLoopForm({ open: true, editLoop, initialIdea: wizardIdea, onCreated, t });
  const modelResources = useWorkerModelResources(
    form.selectedAgentSlug,
    editLoop?.model_resource_id ?? null,
  );

  const {
    runners,
    repositories,
    selectedRunner,
    setSelectedRunnerId: setPodSelectedRunnerId,
    availableAgents,
  } = usePodCreationData(true);

  useEffect(() => {
    setPodSelectedRunnerId(form.selectedRunnerId);
  }, [form.selectedRunnerId, setPodSelectedRunnerId]);

  const {
    fields: configFields,
    loading: loadingConfig,
    config: configValues,
    updateConfig: handleConfigChange,
  } = useConfigOptions(selectedRunner?.id || null, form.selectedAgentSlug);

  const compatibleRunners = useMemo(
    () => runnersSupportingAgent(runners, form.selectedAgentSlug),
    [runners, form.selectedAgentSlug],
  );
  const selectedRunnerCompatible =
    !form.selectedRunnerId || compatibleRunners.some((r) => r.id === form.selectedRunnerId);
  const canSubmitWithRunner =
    hasRunnerForAgent(runners, form.selectedAgentSlug) && selectedRunnerCompatible;

  const [configOverridesRestored, setConfigOverridesRestored] = useState(false);
  useEffect(() => {
    if (editLoop?.config_overrides && configFields.length > 0 && !configOverridesRestored) {
      Object.entries(editLoop.config_overrides).forEach(([key, value]) => {
        handleConfigChange(key, value);
      });
      setConfigOverridesRestored(true);
    }
  }, [editLoop, configFields, configOverridesRestored, handleConfigChange]);

  useEffect(() => {
    if (!form.selectedRepositoryId) {
      form.setSelectedBranch("");
      return;
    }
    const repo = repositories.find((r) => r.id === form.selectedRepositoryId);
    if (repo?.default_branch) form.setSelectedBranch(repo.default_branch);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [form.selectedRepositoryId, repositories]);

  const { envBundles, loadingBundles } = useLoopEnvBundles({
    open: true,
    agentSlug: form.selectedAgentSlug,
  });

  useEffect(() => {
    if (!editLoop || envBundles.length === 0) return;
    const kindByName = new Map(envBundles.map((b) => [b.name, b.kind]));
    const saved = editLoop.used_env_bundles ?? [];
    form.setSelectedRuntimeBundleNames(saved.filter((n) => kindByName.get(n) === "runtime"));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editLoop, envBundles]);

  useEffect(() => {
    if (
      availableAgents.length > 0 &&
      form.selectedAgentSlug &&
      !availableAgents.find((a) => a.slug === form.selectedAgentSlug)
    ) {
      form.setSelectedAgentSlug(null);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [availableAgents, form.selectedAgentSlug]);

  const formTitle = form.isEdit ? t("loops.editLoop") : t("loops.manualSectionTitle");
  const modelResourceRequired = requiresModelResource(form.selectedAgentSlug);
  const canSubmitWithModelResource =
    !modelResourceRequired ||
    Boolean(
      modelResources.selectedModelResource &&
        !modelResources.loadingModelResources &&
        !modelResources.modelResourceError,
    );

  return (
    <div className="space-y-6">
      {showAiSection && <LoopNlCreate onNeedsWizard={setWizardIdea} />}

      {showAiSection && (
        <div className="relative flex items-center gap-3">
          <div className="h-px flex-1 bg-border/60" />
          <span className="text-xs text-muted-foreground">{t("loops.manualSectionDivider")}</span>
          <div className="h-px flex-1 bg-border/60" />
        </div>
      )}

      <section className="space-y-4">
        <h2 className="text-sm font-medium">{formTitle}</h2>
        <LoopCreateDialogBody
          form={form}
          availableAgents={availableAgents}
          runners={runners}
          compatibleRunners={compatibleRunners}
          repositories={repositories}
          envBundles={envBundles}
          modelResources={modelResources.modelResources}
          selectedModelResourceId={modelResources.selectedModelResourceId}
          onSelectModelResource={modelResources.setSelectedModelResourceId}
          loadingModelResources={modelResources.loadingModelResources}
          modelResourceError={modelResources.modelResourceError}
          modelResourceRequired={modelResourceRequired}
          configFields={configFields}
          configValues={configValues}
          loadingConfig={loadingConfig}
          loadingBundles={loadingBundles}
          onConfigChange={handleConfigChange}
          t={t}
          embedded
        />
        <div className="flex justify-end gap-2 pt-2">
          {onCancel && (
            <Button variant="outline" onClick={onCancel}>
              {t("common.cancel")}
            </Button>
          )}
          <Button
            onClick={() => form.submit(
              configValues,
              modelResources.selectedModelResourceId,
              modelResourceRequired,
            )}
            disabled={
              form.loading ||
              !form.name.trim() ||
              !form.promptTemplate.trim() ||
              !form.selectedAgentSlug ||
              !canSubmitWithRunner ||
              !canSubmitWithModelResource
            }
          >
            {form.loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {form.isEdit ? t("common.save") : t("loops.createLoop")}
          </Button>
        </div>
      </section>
    </div>
  );
}

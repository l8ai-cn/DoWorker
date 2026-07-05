"use client";

import React, { useState, useMemo, useEffect, useRef } from "react";
import { useTranslations } from "next-intl";
import { CenteredSpinner } from "@/components/ui/spinner";
import { usePodCreationData, useCreatePodForm } from "../hooks";
import { useConfigOptions } from "@/components/ide/hooks";
import { CreatePodFormProps } from "./types";
import { mergeConfig } from "./presets";
import { useCreatePodSubmitHandler } from "./useCreatePodSubmitHandler";
import { useCreatePodRunnerCompatibility } from "./useCreatePodRunnerCompatibility";
import { CreatePodFormActions } from "./CreatePodFormActions";
import { CreatePodFormFields } from "./CreatePodFormFields";

export function CreatePodForm({
  config,
  enabled = true,
  className,
}: CreatePodFormProps) {
  const t = useTranslations();
  const prevEnabledRef = useRef(enabled);
  const promptInitializedRef = useRef(false);
  const repoInitializedRef = useRef(false);

  const mergedConfig = useMemo(() => mergeConfig(config), [config]);

  const {
    context, promptGenerator, onSuccess, onError, onCancel,
    initialAgentSlug, initialPrompt,
  } = mergedConfig;

  const [selectedAgentSlug, setSelectedAgentSlug] = useState<string | null>(null);
  const agentInitializedRef = useRef(false);

  const {
    runners,
    repositories,
    loading: loadingData,
    selectedRunner,
    setSelectedRunnerId,
    availableAgents,
  } = usePodCreationData(enabled);

  const {
    fields: configFields,
    loading: loadingConfig,
    config: configValues,
    updateConfig: handleConfigChange,
    resetConfig: resetConfig,
  } = useConfigOptions(
    selectedRunner?.id || null,
    selectedAgentSlug
  );

  const form = useCreatePodForm(
    availableAgents,
    repositories,
    onSuccess,
    configValues,
    { repositoryId: context?.ticket?.repositoryId ?? null },
  );

  useEffect(() => {
    setSelectedAgentSlug(form.selectedAgent);
  }, [form.selectedAgent]);

  useEffect(() => {
    if (prevEnabledRef.current && !enabled) {
      form.reset();
      resetConfig();
      setSelectedRunnerId(null);
      setSelectedAgentSlug(null);
      promptInitializedRef.current = false;
      repoInitializedRef.current = false;
      agentInitializedRef.current = false;
    }
    prevEnabledRef.current = enabled;
  }, [enabled]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (
      enabled &&
      initialAgentSlug &&
      !agentInitializedRef.current &&
      availableAgents.some((a) => a.slug === initialAgentSlug)
    ) {
      form.setSelectedAgent(initialAgentSlug);
      agentInitializedRef.current = true;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, initialAgentSlug, availableAgents]);

  const defaultPrompt = useMemo(() => {
    if (initialPrompt) return initialPrompt;
    if (promptGenerator && context) {
      return promptGenerator(context);
    }
    return "";
  }, [initialPrompt, promptGenerator, context]);

  useEffect(() => {
    if (
      enabled &&
      context?.ticket?.repositoryId &&
      !form.selectedRepository &&
      !repoInitializedRef.current &&
      repositories.length > 0
    ) {
      form.setSelectedRepository(context.ticket.repositoryId);
      repoInitializedRef.current = true;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, context?.ticket?.repositoryId, form.selectedRepository, form.setSelectedRepository, repositories]);

  useEffect(() => {
    if (enabled && defaultPrompt && !form.prompt && !promptInitializedRef.current) {
      form.setPrompt(defaultPrompt);
      promptInitializedRef.current = true;
    }
    // form is a stable object from custom hook, only track specific values
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, defaultPrompt, form.prompt, form.setPrompt]);

  const handleCreate = useCreatePodSubmitHandler(
    form, selectedRunner, configValues, context, onError,
  );

  const { canCreate } = useCreatePodRunnerCompatibility({
    runners,
    selectedAgent: form.selectedAgent,
    selectedRunner,
    setSelectedRunnerId,
  });

  return (
    <div className={className}>
      {loadingData ? (
        <CenteredSpinner className="py-8" />
      ) : (
        <CreatePodFormFields
          form={form}
          agents={availableAgents}
          runners={runners}
          repositories={repositories}
          selectedRunner={selectedRunner}
          setSelectedRunnerId={setSelectedRunnerId}
          configFields={configFields}
          loadingConfig={loadingConfig}
          configValues={configValues}
          handleConfigChange={handleConfigChange}
          hasOnlineRunners={runners.length > 0}
          promptPlaceholder={mergedConfig.promptPlaceholder}
          showPerpetual={mergedConfig.scenario === "workspace"}
          t={t}
        />
      )}

      <CreatePodFormActions
        onCancel={onCancel}
        onCreate={handleCreate}
        disabled={!canCreate || form.loading || loadingData}
        loading={form.loading}
        t={t}
      />
    </div>
  );
}

export default CreatePodForm;

export * from "./types";
export * from "./presets";

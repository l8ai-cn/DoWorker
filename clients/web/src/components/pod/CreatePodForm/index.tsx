"use client";

import { useEffect, useMemo } from "react";
import { useTranslations } from "next-intl";
import { NlWorkerCreate } from "@/components/workers/NlWorkerCreate";
import { useRepositories, useRepositoryStore } from "@/stores/repository";
import { useWorkerCreateDraft } from "../hooks";
import { CreatePodFormFields } from "./CreatePodFormFields";
import { mergeConfig } from "./presets";
import type { CreatePodFormProps } from "./types";

export function CreatePodForm({
  config,
  enabled = true,
  className,
}: CreatePodFormProps) {
  const t = useTranslations();
  const mergedConfig = useMemo(() => mergeConfig(config), [config]);
  const repositories = useRepositories();
  const fetched = useRepositoryStore((state) => state.fetched);
  const loadingRepositories = useRepositoryStore((state) => state.isLoading);
  const fetchRepositories = useRepositoryStore(
    (state) => state.fetchRepositories,
  );
  const initialTask = mergedConfig.initialPrompt
    ?? (mergedConfig.promptGenerator && mergedConfig.context
      ? mergedConfig.promptGenerator(mergedConfig.context)
      : "");
  const controller = useWorkerCreateDraft({
    enabled,
    repositories,
    initialWorkerTypeSlug: mergedConfig.initialAgentSlug,
    initialTask: initialTask || undefined,
    initialRepositoryId: mergedConfig.context?.ticket?.repositoryId ?? null,
    ticketSlug: mergedConfig.context?.ticket?.slug,
    onSuccess: mergedConfig.onSuccess,
    onError: mergedConfig.onError,
  });

  useEffect(() => {
    if (enabled && !fetched && !loadingRepositories) {
      void fetchRepositories();
    }
  }, [enabled, fetchRepositories, fetched, loadingRepositories]);

  return (
    <div className={className}>
      <NlWorkerCreate
        prompt={controller.state.fillPrompt}
        filling={controller.state.fill.status === "loading"}
        onPromptChange={controller.setFillPrompt}
        onFill={(prompt) => void controller.fillWithAI(prompt)}
      />
      <CreatePodFormFields
        controller={controller}
        initialWizardStep={mergedConfig.initialWizardStep}
        promptPlaceholder={mergedConfig.promptPlaceholder}
        onCancel={mergedConfig.onCancel}
        t={t}
      />
    </div>
  );
}

export default CreatePodForm;

export * from "./types";
export * from "./presets";

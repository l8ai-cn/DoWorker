"use client";

import { useState, useCallback, useEffect } from "react";
import { useLoopStore } from "@/stores/loop";
import { toast } from "sonner";
import type { LoopData } from "@/lib/viewModels/loop";

export interface UseLoopFormResult {
  name: string;
  setName: (v: string) => void;
  description: string;
  setDescription: (v: string) => void;
  promptTemplate: string;
  setPromptTemplate: (v: string) => void;

  selectedAgentSlug: string | null;
  setSelectedAgentSlug: (v: string | null) => void;
  selectedRunnerId: number | null;
  setSelectedRunnerId: (v: number | null) => void;
  selectedRepositoryId: number | null;
  setSelectedRepositoryId: (v: number | null) => void;
  selectedBranch: string;
  setSelectedBranch: (v: string) => void;
  selectedRuntimeBundleNames: string[];
  setSelectedRuntimeBundleNames: (v: string[]) => void;
  executionMode: string;
  setExecutionMode: (v: string) => void;
  cronEnabled: boolean;
  setCronEnabled: (v: boolean) => void;
  cronExpression: string;
  setCronExpression: (v: string) => void;
  sandboxStrategy: string;
  setSandboxStrategy: (v: string) => void;
  concurrencyPolicy: string;
  setConcurrencyPolicy: (v: string) => void;
  timeoutMinutes: number;
  setTimeoutMinutes: (v: number) => void;
  callbackUrl: string;
  setCallbackUrl: (v: string) => void;
  sessionPersistence: boolean;
  setSessionPersistence: (v: boolean) => void;
  maxConcurrentRuns: number;
  setMaxConcurrentRuns: (v: number) => void;
  maxRetainedRuns: number;
  setMaxRetainedRuns: (v: number) => void;

  loading: boolean;
  isEdit: boolean;
  submit: (
    configValues: Record<string, unknown>,
    modelResourceId: number | null,
    modelResourceRequired: boolean,
  ) => Promise<void>;
}

export function useLoopForm(args: {
  open: boolean;
  editLoop?: LoopData;
  initialIdea?: string;
  onCreated: (createdLoop?: LoopData) => void;
  t: (key: string) => string;
}): UseLoopFormResult {
  const { open, editLoop, initialIdea, onCreated, t } = args;
  const createLoop = useLoopStore((s) => s.createLoop);
  const updateLoop = useLoopStore((s) => s.updateLoop);
  const isEdit = !!editLoop;

  const [loading, setLoading] = useState(false);

  const [name, setName] = useState(editLoop?.name || "");
  const [description, setDescription] = useState(editLoop?.description || "");
  const [promptTemplate, setPromptTemplate] = useState(editLoop?.prompt_template || "");

  const [selectedAgentSlug, setSelectedAgentSlug] = useState<string | null>(editLoop?.agent_slug || null);
  const [selectedRunnerId, setSelectedRunnerId] = useState<number | null>(editLoop?.runner_id || null);
  const [selectedRepositoryId, setSelectedRepositoryId] = useState<number | null>(editLoop?.repository_id || null);
  const [selectedBranch, setSelectedBranch] = useState(editLoop?.branch_name || "");
  const [selectedRuntimeBundleNames, setSelectedRuntimeBundleNames] = useState<string[]>([]);

  const [executionMode, setExecutionMode] = useState<string>(editLoop?.execution_mode || "autopilot");
  const [cronEnabled, setCronEnabled] = useState(!!editLoop?.cron_expression);
  const [cronExpression, setCronExpression] = useState(editLoop?.cron_expression || "");
  const [sandboxStrategy, setSandboxStrategy] = useState<string>(editLoop?.sandbox_strategy || "persistent");
  const [concurrencyPolicy, setConcurrencyPolicy] = useState<string>(editLoop?.concurrency_policy || "skip");
  const [timeoutMinutes, setTimeoutMinutes] = useState(editLoop?.timeout_minutes || 60);
  const [callbackUrl, setCallbackUrl] = useState(editLoop?.callback_url || "");
  const [sessionPersistence, setSessionPersistence] = useState(editLoop?.session_persistence ?? true);
  const [maxConcurrentRuns, setMaxConcurrentRuns] = useState(editLoop?.max_concurrent_runs || 1);
  const [maxRetainedRuns, setMaxRetainedRuns] = useState(editLoop?.max_retained_runs || 0);

  useEffect(() => {
    if (!open) return;
    setName(editLoop?.name || "");
    setDescription(editLoop?.description || "");
    setPromptTemplate(editLoop?.prompt_template || initialIdea || "");
    setSelectedAgentSlug(editLoop?.agent_slug || null);
    setSelectedRunnerId(editLoop?.runner_id || null);
    setSelectedRepositoryId(editLoop?.repository_id || null);
    setSelectedBranch(editLoop?.branch_name || "");
    setSelectedRuntimeBundleNames([]);
    setExecutionMode(editLoop?.execution_mode || "autopilot");
    setCronEnabled(!!editLoop?.cron_expression);
    setCronExpression(editLoop?.cron_expression || "");
    setSandboxStrategy(editLoop?.sandbox_strategy || "persistent");
    setConcurrencyPolicy(editLoop?.concurrency_policy || "skip");
    setTimeoutMinutes(editLoop?.timeout_minutes || 60);
    setCallbackUrl(editLoop?.callback_url || "");
    setSessionPersistence(editLoop?.session_persistence ?? true);
    setMaxConcurrentRuns(editLoop?.max_concurrent_runs || 1);
    setMaxRetainedRuns(editLoop?.max_retained_runs || 0);
    setLoading(false);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  useEffect(() => {
    if (!open || editLoop || !initialIdea) return;
    setPromptTemplate(initialIdea);
  }, [open, editLoop, initialIdea]);

  const submit = useCallback(
    async (
      configValues: Record<string, unknown>,
      modelResourceId: number | null,
      modelResourceRequired: boolean,
    ) => {
      if (!name.trim() || !promptTemplate.trim() || !selectedAgentSlug) return;
      if (modelResourceRequired && !modelResourceId) return;

      setLoading(true);
      try {
        const data = {
          name: name.trim(),
          description: description || undefined,
          agent_slug: selectedAgentSlug,
          prompt_template: promptTemplate,
          runner_id: selectedRunnerId || undefined,
          repository_id: selectedRepositoryId || undefined,
          branch_name: selectedBranch || undefined,
          model_resource_id: modelResourceId || undefined,
          used_env_bundles: selectedRuntimeBundleNames.filter(Boolean),
          config_overrides: Object.keys(configValues).length > 0 ? configValues : undefined,
          execution_mode: executionMode,
          cron_expression: cronEnabled && cronExpression ? cronExpression : "",
          sandbox_strategy: sandboxStrategy,
          concurrency_policy: concurrencyPolicy,
          timeout_minutes: timeoutMinutes,
          callback_url: callbackUrl || undefined,
          session_persistence: sessionPersistence,
          max_concurrent_runs: maxConcurrentRuns,
          max_retained_runs: maxRetainedRuns,
        };

        if (isEdit && editLoop) {
          await updateLoop(editLoop.slug, data);
          toast.success(t("loops.updated"));
          onCreated();
        } else {
          const res = await createLoop(data);
          toast.success(t("loops.created"));
          onCreated(res.loop);
        }
      } catch (err) {
        toast.error(isEdit ? t("loops.updateFailed") : t("loops.createFailed"), {
          description: (err as Error).message,
        });
      } finally {
        setLoading(false);
      }
    },
    [
      name, description, promptTemplate, selectedAgentSlug, selectedRunnerId,
      selectedRepositoryId, selectedBranch, selectedRuntimeBundleNames,
      executionMode, cronEnabled, cronExpression, sandboxStrategy, concurrencyPolicy,
      timeoutMinutes, callbackUrl, sessionPersistence, maxConcurrentRuns,
      maxRetainedRuns, isEdit, editLoop, createLoop, updateLoop, onCreated, t,
    ]
  );

  return {
    name, setName,
    description, setDescription,
    promptTemplate, setPromptTemplate,
    selectedAgentSlug, setSelectedAgentSlug,
    selectedRunnerId, setSelectedRunnerId,
    selectedRepositoryId, setSelectedRepositoryId,
    selectedBranch, setSelectedBranch,
    selectedRuntimeBundleNames, setSelectedRuntimeBundleNames,
    executionMode, setExecutionMode,
    cronEnabled, setCronEnabled,
    cronExpression, setCronExpression,
    sandboxStrategy, setSandboxStrategy,
    concurrencyPolicy, setConcurrencyPolicy,
    timeoutMinutes, setTimeoutMinutes,
    callbackUrl, setCallbackUrl,
    sessionPersistence, setSessionPersistence,
    maxConcurrentRuns, setMaxConcurrentRuns,
    maxRetainedRuns, setMaxRetainedRuns,
    loading, isEdit, submit,
  };
}

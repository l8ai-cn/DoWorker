"use client";

import { useState, useCallback, useEffect } from "react";
import { useWorkflowStore } from "@/stores/workflow";
import { toast } from "sonner";
import type { WorkflowData } from "@/lib/viewModels/workflow";

export interface UseWorkflowFormResult {
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

export function useWorkflowForm(args: {
  open: boolean;
  editWorkflow?: WorkflowData;
  initialIdea?: string;
  onCreated: (createdWorkflow?: WorkflowData) => void;
  t: (key: string) => string;
}): UseWorkflowFormResult {
  const { open, editWorkflow, initialIdea, onCreated, t } = args;
  const createWorkflow = useWorkflowStore((s) => s.createWorkflow);
  const updateWorkflow = useWorkflowStore((s) => s.updateWorkflow);
  const isEdit = !!editWorkflow;

  const [loading, setLoading] = useState(false);

  const [name, setName] = useState(editWorkflow?.name || "");
  const [description, setDescription] = useState(editWorkflow?.description || "");
  const [promptTemplate, setPromptTemplate] = useState(editWorkflow?.prompt_template || "");

  const [selectedAgentSlug, setSelectedAgentSlug] = useState<string | null>(editWorkflow?.agent_slug || null);
  const [selectedRunnerId, setSelectedRunnerId] = useState<number | null>(editWorkflow?.runner_id || null);
  const [selectedRepositoryId, setSelectedRepositoryId] = useState<number | null>(editWorkflow?.repository_id || null);
  const [selectedBranch, setSelectedBranch] = useState(editWorkflow?.branch_name || "");
  const [selectedRuntimeBundleNames, setSelectedRuntimeBundleNames] = useState<string[]>([]);

  const [executionMode, setExecutionMode] = useState<string>(editWorkflow?.execution_mode || "autopilot");
  const [cronEnabled, setCronEnabled] = useState(!!editWorkflow?.cron_expression);
  const [cronExpression, setCronExpression] = useState(editWorkflow?.cron_expression || "");
  const [sandboxStrategy, setSandboxStrategy] = useState<string>(editWorkflow?.sandbox_strategy || "persistent");
  const [concurrencyPolicy, setConcurrencyPolicy] = useState<string>(editWorkflow?.concurrency_policy || "skip");
  const [timeoutMinutes, setTimeoutMinutes] = useState(editWorkflow?.timeout_minutes || 60);
  const [callbackUrl, setCallbackUrl] = useState(editWorkflow?.callback_url || "");
  const [sessionPersistence, setSessionPersistence] = useState(editWorkflow?.session_persistence ?? true);
  const [maxConcurrentRuns, setMaxConcurrentRuns] = useState(editWorkflow?.max_concurrent_runs || 1);
  const [maxRetainedRuns, setMaxRetainedRuns] = useState(editWorkflow?.max_retained_runs || 0);

  useEffect(() => {
    if (!open) return;
    setName(editWorkflow?.name || "");
    setDescription(editWorkflow?.description || "");
    setPromptTemplate(editWorkflow?.prompt_template || initialIdea || "");
    setSelectedAgentSlug(editWorkflow?.agent_slug || null);
    setSelectedRunnerId(editWorkflow?.runner_id || null);
    setSelectedRepositoryId(editWorkflow?.repository_id || null);
    setSelectedBranch(editWorkflow?.branch_name || "");
    setSelectedRuntimeBundleNames([]);
    setExecutionMode(editWorkflow?.execution_mode || "autopilot");
    setCronEnabled(!!editWorkflow?.cron_expression);
    setCronExpression(editWorkflow?.cron_expression || "");
    setSandboxStrategy(editWorkflow?.sandbox_strategy || "persistent");
    setConcurrencyPolicy(editWorkflow?.concurrency_policy || "skip");
    setTimeoutMinutes(editWorkflow?.timeout_minutes || 60);
    setCallbackUrl(editWorkflow?.callback_url || "");
    setSessionPersistence(editWorkflow?.session_persistence ?? true);
    setMaxConcurrentRuns(editWorkflow?.max_concurrent_runs || 1);
    setMaxRetainedRuns(editWorkflow?.max_retained_runs || 0);
    setLoading(false);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  useEffect(() => {
    if (!open || editWorkflow || !initialIdea) return;
    setPromptTemplate(initialIdea);
  }, [open, editWorkflow, initialIdea]);

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

        if (isEdit && editWorkflow) {
          await updateWorkflow(editWorkflow.slug, data);
          toast.success(t("workflows.updated"));
          onCreated();
        } else {
          const res = await createWorkflow(data);
          toast.success(t("workflows.created"));
          onCreated(res.workflow);
        }
      } catch (err) {
        toast.error(isEdit ? t("workflows.updateFailed") : t("workflows.createFailed"), {
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
      maxRetainedRuns, isEdit, editWorkflow, createWorkflow, updateWorkflow, onCreated, t,
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

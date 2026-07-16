"use client";

import { useCallback, useState } from "react";
import { applyLoopAIDraft } from "@/lib/api/facade/loopProgramConnect";
import type { LoopWorkbenchSnapshot } from "@/lib/viewModels/loop-program";
import type {
  LoopAIMode,
  LoopAIProposal,
  LoopAIRepairTarget,
} from "./loop-ai-assistant-types";
import type { LoopAIMessages } from "./loop-workbench-messages";
import { requestLoopAIProposal } from "./request-loop-ai-proposal";
import { useLoopAIResources } from "./use-loop-ai-resources";

interface UseLoopAIAssistantInput {
  orgSlug: string;
  locale: string;
  snapshot: LoopWorkbenchSnapshot;
  messages: LoopAIMessages;
  onApplied: (snapshot: LoopWorkbenchSnapshot) => void;
}

export function useLoopAIAssistant({
  orgSlug,
  locale,
  snapshot,
  messages,
  onApplied,
}: UseLoopAIAssistantInput) {
  const resourceState = useLoopAIResources(orgSlug, messages.resourceError);
  const [open, setOpen] = useState(false);
  const [mode, setMode] = useState<LoopAIMode>("generate");
  const [prompt, setPrompt] = useState("");
  const [resourceSelection, setResourceSelection] = useState({
    orgSlug,
    resourceId: "",
  });
  const [busy, setBusy] = useState(false);
  const [requestError, setRequestError] = useState<string>();
  const [proposal, setProposal] = useState<LoopAIProposal>();
  const [repairTarget, setRepairTarget] = useState<LoopAIRepairTarget>();
  const selectedResourceId =
    resourceSelection.orgSlug === orgSlug ? resourceSelection.resourceId : "";

  const clearResult = useCallback(() => {
    setProposal(undefined);
    setRequestError(undefined);
  }, []);

  const changeOpen = useCallback((nextOpen: boolean) => {
    setOpen(nextOpen);
    if (!nextOpen) {
      clearResult();
      setBusy(false);
      setPrompt("");
      setRepairTarget(undefined);
    }
  }, [clearResult]);

  const changeMode = useCallback((nextMode: LoopAIMode) => {
    setMode(nextMode);
    setRepairTarget(undefined);
    clearResult();
  }, [clearResult]);

  const openRepair = useCallback((target: LoopAIRepairTarget) => {
    clearResult();
    setPrompt("");
    setRepairTarget(target);
    setOpen(true);
  }, [clearResult]);

  const submit = useCallback(async () => {
    if (mode !== "generate" && !repairTarget) return;
    const modelResourceId = Number(selectedResourceId);
    const trimmedPrompt = prompt.trim();
    if (
      !resourceState.resources.some((resource) => resource.id === selectedResourceId) ||
      !Number.isSafeInteger(modelResourceId) ||
      modelResourceId <= 0 ||
      (!repairTarget && !trimmedPrompt)
    ) {
      return;
    }

    setBusy(true);
    clearResult();
    try {
      const nextProposal = await requestLoopAIProposal({
        orgSlug,
        locale,
        snapshot,
        modelResourceId,
        prompt: trimmedPrompt,
        repairTarget,
      });
      if (nextProposal.proposedSource === snapshot.source) {
        setRequestError(messages.unchanged);
        return;
      }
      setProposal(nextProposal);
    } catch {
      setRequestError(repairTarget ? messages.repair.error : messages.generationError);
    } finally {
      setBusy(false);
    }
  }, [
    clearResult,
    locale,
    messages.generationError,
    messages.repair.error,
    messages.unchanged,
    mode,
    orgSlug,
    prompt,
    repairTarget,
    resourceState.resources,
    selectedResourceId,
    snapshot,
  ]);

  const confirm = useCallback(async () => {
    if (!proposal) return;
    setBusy(true);
    setRequestError(undefined);
    try {
      const result = await applyLoopAIDraft(proposal.response);
      if (!result.applied) {
        setRequestError(messages.stale);
        return;
      }
      onApplied(result.snapshot);
      changeOpen(false);
    } catch {
      setRequestError(proposal.repair ? messages.repair.error : messages.generationError);
    } finally {
      setBusy(false);
    }
  }, [
    changeOpen,
    messages.generationError,
    messages.repair.error,
    messages.stale,
    onApplied,
    proposal,
  ]);

  return {
    ...resourceState,
    open,
    mode,
    prompt,
    selectedResourceId,
    busy,
    requestError,
    proposal,
    repairTarget,
    setOpen: changeOpen,
    setMode: changeMode,
    setPrompt,
    setSelectedResourceId: (resourceId: string) =>
      setResourceSelection({ orgSlug, resourceId }),
    openRepair,
    submit,
    back: clearResult,
    confirm,
  };
}

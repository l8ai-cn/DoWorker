import { useEffect, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@/lib/routing";
import type { Host } from "@/hooks/useHosts";
import { useModelConfigs } from "@/hooks/useModelConfigs";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import { setPendingInitialPrompt } from "@/store/chatStore";
import { appendPromptHistoryEntry } from "@/hooks/usePromptHistory";
import { hostSupportsAgent } from "@/lib/host-agent-match";
import { hostDisplayLabel } from "@/lib/hostDisplayLabel";
import { buildMentionPreamble, type MentionItem } from "@/lib/composerMentions";
import { matchSkillInvocation, sanitizeInitialPrompt } from "./newChatWorkspace";
import { harnessWarningMessageText } from "./newChatHarnessWarning";
import { createNewChatSession, newChatCreateDisabledReason } from "./newChatSessionCreation";
import { assignNewChatProject } from "./newChatProjectAssignment";

export function useNewChatSubmitState(input: NewChatSubmitInput) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [modelConfigId, setModelConfigId] = useState<number | null>(null);
  const [workerTokenBudget, setWorkerTokenBudget] = useState<number | null>(null);
  const showWorkerModelPicker = input.selectedAgent?.requiresModelResource === true;
  useEffect(() => {
    if (!showWorkerModelPicker) setModelConfigId(null);
  }, [showWorkerModelPicker]);
  const workerModelResources = useModelConfigs(showWorkerModelPicker);
  const workerModelResourceReady =
    !showWorkerModelPicker ||
    (!workerModelResources.isLoading &&
      !workerModelResources.isError &&
      modelConfigId !== null &&
      Boolean(workerModelResources.data?.some((resource) => resource.id === modelConfigId)));
  const workerPlanDisabledReason = newChatCreateDisabledReason({
    agent: input.selectedAgent,
    hostId: input.selectedHostId,
    workspace: input.workspaceTrimmed,
    sandboxSelected: input.sandboxSelected,
    sandboxRepoUrl: input.sandboxRepoUrl,
    sandboxRepoBranch: input.sandboxRepoBranch,
    branchName: input.branchName,
    modelResourceId: showWorkerModelPicker ? modelConfigId : null,
    tokenBudget: workerTokenBudget,
  });
  const canSubmit =
    input.message.trim().length > 0 &&
    input.selectedAgent !== undefined &&
    workerModelResourceReady &&
    workerPlanDisabledReason === null &&
    (input.sandboxSelected
      ? input.sandboxRepoValid
      : input.selectedHostId !== null &&
        input.workspaceValid &&
        input.selectedHost !== undefined &&
        hostSupportsAgent(input.selectedHost, input.selectedAgent)) &&
    !creating;
  const submitDisabledReason = canSubmit ? null : disabledReason(input, {
    workerModelResourceReady,
    workerPlanDisabledReason,
  });

  const handleCreate = async () => {
    if (!canSubmit) return;
    setCreating(true);
    setCreateError(null);
    try {
      const data = await createNewChatSession({
        agent: input.selectedAgent,
        hostId: input.selectedHostId,
        workspace: input.workspaceTrimmed,
        sandboxSelected: input.sandboxSelected,
        sandboxRepoUrl: input.sandboxRepoUrl,
        sandboxRepoBranch: input.sandboxRepoBranch,
        branchName: input.branchName,
        modelResourceId: showWorkerModelPicker ? modelConfigId : null,
        tokenBudget: workerTokenBudget,
      });
      await assignNewChatProject(data.id, input.selectedProject, queryClient);
      if (!input.sandboxSelected) input.addRecent(input.workspaceTrimmed);
      void queryClient.refetchQueries({ queryKey: ["conversations"] });
      void queryClient.invalidateQueries({ queryKey: ["directory-sessions"] });
      const initialPrompt =
        buildMentionPreamble(input.mentionedItems, input.selectedAgent?.harness ?? null) +
        sanitizeInitialPrompt(input.message);
      setPendingInitialPrompt(data.id, {
        text: initialPrompt,
        skill: input.isNativeTerminalAgent
          ? null
          : matchSkillInvocation(initialPrompt, input.selectedAgent?.skills ?? []),
        files: input.files,
      });
      appendPromptHistoryEntry(initialPrompt, data.id);
      input.markSubmitted();
      navigate(`/c/${data.id}`);
    } catch (error) {
      setCreateError(
        error instanceof Error ? error.message : "Couldn't reach the server. Check your connection and try again.",
      );
    } finally {
      setCreating(false);
    }
  };

  return {
    creating,
    createError,
    showWorkerModelPicker,
    modelConfigId,
    setModelConfigId,
    workerTokenBudget,
    setWorkerTokenBudget,
    canSubmit,
    submitDisabledReason,
    handleCreate,
  };
}

function disabledReason(
  input: NewChatSubmitInput,
  plan: { workerModelResourceReady: boolean; workerPlanDisabledReason: string | null },
) {
  if (input.sandboxSelected && !input.sandboxRepoValid) return "Please enter a valid repository URL";
  if (
    !input.sandboxSelected &&
    input.selectedAgent &&
    input.selectedHost &&
    !hostSupportsAgent(input.selectedHost, input.selectedAgent)
  ) {
    return "This host cannot run the selected agent";
  }
  if (input.selectedAgentUnavailableReason) {
    return harnessWarningMessageText(
      input.selectedAgent?.display_name,
      input.selectedHost ? hostDisplayLabel(input.selectedHost, { thisMachineHostId: input.thisMachineHostId }) : undefined,
      input.selectedAgentUnavailableReason,
    );
  }
  if (!input.sandboxSelected && (!input.selectedHostId || !input.workspaceValid)) {
    return input.chooseHostWorkspaceText;
  }
  if (!plan.workerModelResourceReady) return "Select an AI resource";
  if (plan.workerPlanDisabledReason) return plan.workerPlanDisabledReason;
  if (input.message.trim().length === 0) return input.enterMessageText;
  return null;
}

type NewChatSubmitInput = {
  selectedAgent: AvailableAgent | undefined;
  selectedHost: Host | undefined;
  selectedHostId: string | null;
  thisMachineHostId: string | null;
  sandboxSelected: boolean;
  sandboxRepoValid: boolean;
  sandboxRepoUrl: string;
  sandboxRepoBranch: string;
  workspaceTrimmed: string;
  workspaceValid: boolean;
  branchName: string;
  message: string;
  files: File[];
  mentionedItems: MentionItem[];
  selectedProject: string;
  selectedAgentUnavailableReason: string | null;
  isNativeTerminalAgent: boolean;
  chooseHostWorkspaceText: string;
  enterMessageText: string;
  addRecent: (workspace: string) => void;
  markSubmitted: () => void;
};

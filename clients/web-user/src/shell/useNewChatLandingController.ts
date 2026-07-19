import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "@/lib/routing";
import { getCliServerUrl } from "@/lib/host";
import { hostDisplayLabel } from "@/lib/hostDisplayLabel";
import { isNativeCodingAgent } from "@/lib/nativeCodingAgents";
import { useAutoGrowTextarea } from "@/hooks/useAutoGrowTextarea";
import { useNativeServerSwitcherForMainSurface } from "@/hooks/useNativeServerSwitcher";
import { useI18n } from "@/i18n/I18nProvider";
import {
  getNewChatLandingDraft,
  preserveNewChatLandingDraft,
  type NewChatLandingDraft,
} from "./newChatLandingDraft";
import {
  HarnessWarningMessage,
  harnessUnconfiguredOnHost,
  harnessUnavailableReasonOnHost,
} from "./newChatHarnessWarning";
import { useNewChatAgentSelection } from "./useNewChatAgentSelection";
import { useNewChatFiles } from "./useNewChatFiles";
import { useNewChatLocationState } from "./useNewChatLocationState";
import { useNewChatMentionState } from "./useNewChatMentionState";
import { useNewChatSlashState } from "./useNewChatSlashState";
import { useNewChatSubmitState } from "./useNewChatSubmitState";

export function useNewChatLandingController() {
  const { t } = useI18n();
  const landingDraft = getNewChatLandingDraft();
  const [searchParams] = useSearchParams();
  const projectParam = searchParams.get("project") ?? "";
  const agentParam = searchParams.get("agent") ?? "";
  const serverUrl = getCliServerUrl();
  const agent = useNewChatAgentSelection({ agentParam, landingDraft });
  const location = useNewChatLocationState({
    landingDraft,
    projectParam,
    selectedAgent: agent.selectedAgent,
  });
  const [landingSurface, setLandingSurface] = useState<HTMLElement | null>(null);
  useNativeServerSwitcherForMainSurface(landingSurface, true);
  const [message, setMessage] = useState(() => landingDraft?.message ?? "");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const isComposingRef = useRef(false);
  useAutoGrowTextarea(textareaRef, message, 9);
  const files = useNewChatFiles(landingDraft?.files);
  const submittedRef = useRef(false);
  const draftRef = useRef<NewChatLandingDraft>(null as unknown as NewChatLandingDraft);
  draftRef.current = {
    message,
    files: files.files,
    pickedAgentId: agent.pickedAgentId,
    selectedHostId: location.selectedHostId,
    sandboxSelected: location.sandboxSelected,
    sandboxRepoUrl: location.sandboxRepoUrl,
    sandboxRepoBranch: location.sandboxRepoBranch,
    workspace: location.workspace,
    branchName: location.branchName,
    baseBranch: location.baseBranch,
  };
  useEffect(() => {
    return () => {
      preserveNewChatLandingDraft(submittedRef.current ? null : draftRef.current);
    };
  }, []);
  const isNativeTerminalAgent = isNativeCodingAgent(agent.selectedAgent);
  const harnessWarningHost = !location.sandboxSelected ? location.selectedHost : undefined;
  const selectedAgentUnconfigured = harnessUnconfiguredOnHost(
    agent.selectedAgent?.harness,
    harnessWarningHost,
  );
  const selectedAgentUnavailableReason = harnessUnavailableReasonOnHost(
    agent.selectedAgent?.harness,
    harnessWarningHost,
  );
  const slash = useNewChatSlashState({
    selectedAgent: agent.selectedAgent,
    isNativeTerminalAgent,
    message,
    setMessage,
    textareaRef,
  });
  const mention = useNewChatMentionState({
    isNativeTerminalAgent,
    sandboxSelected: location.sandboxSelected,
    selectedHostId: location.selectedHostId,
    workspaceValid: location.workspaceValid,
    workspaceTrimmed: location.workspaceTrimmed,
    message,
    setMessage,
    textareaRef,
  });
  const submit = useNewChatSubmitState({
    selectedAgent: agent.selectedAgent,
    selectedHost: location.selectedHost,
    selectedHostId: location.selectedHostId,
    thisMachineHostId: location.thisMachineHostId,
    sandboxSelected: location.sandboxSelected,
    sandboxRepoValid: location.sandboxRepoValid,
    sandboxRepoUrl: location.sandboxRepoUrl,
    sandboxRepoBranch: location.sandboxRepoBranch,
    workspaceTrimmed: location.workspaceTrimmed,
    workspaceValid: location.workspaceValid,
    branchName: location.branchName,
    message,
    files: files.files,
    mentionedItems: mention.mentionedItems,
    selectedProject: location.selectedProject,
    selectedAgentUnavailableReason,
    isNativeTerminalAgent,
    chooseHostWorkspaceText: t.composer.chooseHostWorkspace,
    enterMessageText: t.composer.enterMessage,
    addRecent: location.addRecent,
    markSubmitted: () => {
      submittedRef.current = true;
      preserveNewChatLandingDraft(null);
    },
  });
  const workspaceLabel = location.workspaceTrimmed
    ? (location.workspaceTrimmed.split("/").filter(Boolean).pop() ?? location.workspaceTrimmed)
    : t.composer.workingDirectory;
  const hostLabel = location.connectingThisMachine
    ? t.composer.connecting
    : location.sandboxSelected
      ? location.sandboxLabel
      : location.selectedHost
        ? hostDisplayLabel(location.selectedHost, { thisMachineHostId: location.thisMachineHostId })
        : location.onlineHosts.length === 0
          ? t.composer.noHosts
          : t.composer.selectHost;
  const worktreeLabel = location.branchName.trim() || t.composer.noWorktree;
  const agentLabel = agent.selectedAgent ? agent.selectedAgent.display_name : t.composer.selectAgent;

  return {
    title: t.composer.heading,
    placeholder: t.composer.placeholder,
    placeholderSkills: t.composer.placeholderSkills,
    setLandingSurface,
    serverUrl,
    message,
    setMessage,
    textareaRef,
    isComposingRef,
    agent,
    location,
    files,
    slash,
    mention,
    submit,
    labels: {
      workspace: workspaceLabel,
      host: hostLabel,
      branch: worktreeLabel,
      repo: location.sandboxRepoLabel || t.composer.repository,
      agent: agentLabel,
    },
    warning: {
      show: selectedAgentUnconfigured,
      agentName: agent.selectedAgent?.display_name,
      hostName: harnessWarningHost?.name,
      reason: selectedAgentUnavailableReason,
      Message: HarnessWarningMessage,
    },
  };
}

export type NewChatLandingController = ReturnType<typeof useNewChatLandingController>;

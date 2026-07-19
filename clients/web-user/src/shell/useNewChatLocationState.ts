import { useEffect, useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useServerInfo } from "@/lib/CapabilitiesContext";
import { sandboxOptionLabel } from "@/lib/capabilities";
import { getDoWorkerHostConfig } from "@/lib/host";
import {
  collapseInternalWorkspaceHostsForPicker,
  hostSupportsAgent,
  pickOnlineHostForAgent,
} from "@/lib/host-agent-match";
import { controlHost, getHostIdentity } from "@/lib/nativeBridge";
import { isCurrentServerLocal } from "@/lib/serverOrigin";
import { useHostFilesystem } from "@/hooks/useHostFilesystem";
import { useHosts } from "@/hooks/useHosts";
import { useRecentWorkspaces } from "@/hooks/useRecentWorkspaces";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import type { NewChatLandingDraft } from "./newChatLandingDraft";
import {
  deriveHomeDir,
  deriveRepoName,
  isValidSandboxRepoUrl,
  isValidWorkspace,
} from "./newChatWorkspace";
import { useNewChatDesktopHost } from "./useNewChatDesktopHost";
import { useNewChatWorkspaceOccupancy } from "./useNewChatWorkspaceOccupancy";

export function useNewChatLocationState({
  landingDraft,
  projectParam,
  selectedAgent,
}: {
  landingDraft: NewChatLandingDraft | null;
  projectParam: string;
  selectedAgent: AvailableAgent | undefined;
}) {
  const queryClient = useQueryClient();
  const info = useServerInfo();
  const managedSandboxesEnabled = info !== "loading" && info.managed_sandboxes_enabled;
  const sandboxLabel = sandboxOptionLabel(info !== "loading" ? info.sandbox_provider : null);
  const docsLinks = getDoWorkerHostConfig().docsLinks;
  const { data: hosts } = useHosts();
  const [selectedHostId, setSelectedHostId] = useState<string | null>(
    () => landingDraft?.selectedHostId ?? null,
  );
  const [sandboxSelected, setSandboxSelected] = useState(() => landingDraft?.sandboxSelected ?? false);
  const [connectingThisMachine, setConnectingThisMachine] = useState(false);
  const [sandboxRepoUrl, setSandboxRepoUrl] = useState(() => landingDraft?.sandboxRepoUrl ?? "");
  const [sandboxRepoBranch, setSandboxRepoBranch] = useState(() => landingDraft?.sandboxRepoBranch ?? "");
  const [workspace, setWorkspace] = useState(() => landingDraft?.workspace ?? "");
  const [branchName, setBranchName] = useState(() => landingDraft?.branchName ?? "");
  const [baseBranch, setBaseBranch] = useState(() => landingDraft?.baseBranch ?? "");
  const [selectedProject, setSelectedProject] = useState(() => projectParam);
  const [workspacePopoverOpen, setWorkspacePopoverOpen] = useState(false);
  const [connectOpen, setConnectOpen] = useState(false);
  const pendingConnectRef = useRef(false);
  const seededHostRef = useRef<string | null>(null);
  const { recent, addRecent } = useRecentWorkspaces(selectedHostId);

  const allHosts = hosts ?? [];
  const {
    setDesktopHost,
    thisMachineHostId,
    canConnectThisMachine,
    showConnectThisMachine,
  } = useNewChatDesktopHost(allHosts);

  useEffect(() => setSelectedProject(projectParam), [projectParam]);
  const pickerHosts = useMemo(
    () => collapseInternalWorkspaceHostsForPicker(allHosts, selectedAgent),
    [allHosts, selectedAgent],
  );
  const onlineHosts = pickerHosts.filter((host) => host.status === "online");
  const offlineHosts = pickerHosts.filter((host) => host.status === "offline");

  const needsHomeFallback = selectedHostId !== null && recent.length === 0;
  const { data: homeListing, isPlaceholderData } = useHostFilesystem(
    selectedHostId,
    needsHomeFallback ? "" : null,
  );
  const derivedHome = useMemo(() => {
    if (isPlaceholderData) return null;
    return deriveHomeDir(homeListing?.entries ?? []) ?? homeListing?.workspaceRoot ?? null;
  }, [homeListing, isPlaceholderData]);
  const localFallback = isCurrentServerLocal() ? "/workspace" : null;

  useEffect(() => {
    if (selectedHostId === null || seededHostRef.current === selectedHostId) return;
    const candidate = recent[0] ?? derivedHome ?? localFallback;
    if (!candidate) return;
    seededHostRef.current = selectedHostId;
    setWorkspace((current) => (current === "" ? candidate : current));
  }, [selectedHostId, recent, derivedHome, localFallback]);

  useEffect(() => {
    if (sandboxSelected) return;
    if (managedSandboxesEnabled) {
      if (selectedHostId === null) setSandboxSelected(true);
      return;
    }
    if (selectedHostId !== null) return;
    const match = pickOnlineHostForAgent(hosts ?? [], selectedAgent);
    if (match) setSelectedHostId(match.host_id);
  }, [hosts, selectedHostId, sandboxSelected, managedSandboxesEnabled, selectedAgent]);

  useEffect(() => {
    if (sandboxSelected || !selectedAgent || selectedHostId === null) return;
    const current = (hosts ?? []).find((host) => host.host_id === selectedHostId);
    if (current && hostSupportsAgent(current, selectedAgent)) return;
    const match = pickOnlineHostForAgent(hosts ?? [], selectedAgent);
    if (match && match.host_id !== selectedHostId) {
      seededHostRef.current = null;
      setSelectedHostId(match.host_id);
    }
  }, [hosts, selectedHostId, sandboxSelected, selectedAgent]);

  const selectHost = (hostId: string) => {
    if (hostId === selectedHostId && !sandboxSelected) return;
    setSandboxSelected(false);
    setSelectedHostId(hostId);
    setWorkspace("");
    seededHostRef.current = null;
  };
  const selectSandbox = () => {
    if (sandboxSelected) return;
    setSandboxSelected(true);
    setSelectedHostId(null);
    setWorkspace("");
    seededHostRef.current = null;
  };
  const connectThisMachine = async () => {
    if (connectingThisMachine) return;
    setConnectingThisMachine(true);
    try {
      const res = await controlHost("start");
      if (!res.ok) return;
      const identity = await getHostIdentity();
      setDesktopHost(identity);
      await queryClient.invalidateQueries({ queryKey: ["hosts"] });
      if (identity?.hostId) selectHost(identity.hostId);
    } finally {
      setConnectingThisMachine(false);
    }
  };

  const selectedHost = allHosts.find((host) => host.host_id === selectedHostId);
  const workspaceTrimmed = workspace.trim();
  const sandboxRepoName = deriveRepoName(sandboxRepoUrl);
  const occupancyForPath = useNewChatWorkspaceOccupancy(selectedHostId, branchName);
  return {
    selectedHostId,
    sandboxSelected,
    sandboxRepoUrl,
    sandboxRepoBranch,
    workspace,
    workspaceTrimmed,
    workspaceValid: isValidWorkspace(workspace),
    branchName,
    baseBranch,
    selectedProject,
    workspacePopoverOpen,
    connectOpen,
    pendingConnectRef,
    allHosts,
    onlineHosts,
    offlineHosts,
    selectedHost,
    thisMachineHostId,
    canConnectThisMachine,
    showConnectThisMachine,
    connectingThisMachine,
    managedSandboxesEnabled,
    sandboxLabel,
    showDisabledSandboxWithDocs: !managedSandboxesEnabled && Boolean(docsLinks?.newSandbox),
    newSandboxTooltipContent: docsLinks?.newSandbox,
    databricksGitCredentialsTooltipContent: docsLinks?.databricksGitCredentials,
    sandboxRepoValid: sandboxRepoUrl.trim() === "" ? sandboxRepoBranch.trim() === "" : isValidSandboxRepoUrl(sandboxRepoUrl),
    sandboxRepoName,
    sandboxRepoLabel: sandboxRepoName ? (sandboxRepoBranch.trim() ? `${sandboxRepoName}#${sandboxRepoBranch.trim()}` : sandboxRepoName) : "",
    occupancyForPath,
    setSandboxRepoUrl,
    setSandboxRepoBranch,
    setWorkspace,
    setBranchName,
    setBaseBranch,
    setSelectedProject,
    setWorkspacePopoverOpen,
    setConnectOpen,
    selectHost,
    selectSandbox,
    connectThisMachine,
    addRecent,
  };
}

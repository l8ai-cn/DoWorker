import type { MutableRefObject } from "react";
import type { Host } from "@/hooks/useHosts";
import { LandingProjectPicker } from "./LandingProjectPicker";
import { NewChatBranchChip } from "./NewChatBranchChip";
import { NewChatHostChip } from "./NewChatHostChip";
import { NewChatRepositoryChip } from "./NewChatRepositoryChip";
import { NewChatWorkspaceChip } from "./NewChatWorkspaceChip";

export function NewChatLocationTray({
  pendingConnectRef,
  selectedHostId,
  sandboxSelected,
  selectedHost,
  allHosts,
  onlineHosts,
  offlineHosts,
  thisMachineHostId,
  canConnectThisMachine,
  showConnectThisMachine,
  connectingThisMachine,
  hostLabel,
  managedSandboxesEnabled,
  showDisabledSandboxWithDocs,
  sandboxLabel,
  newSandboxTooltipContent,
  onSelectHost,
  onSelectSandbox,
  onConnectThisMachine,
  onShowConnectInstructions,
  repoLabel,
  repoHasValue,
  repoUrl,
  repoBranch,
  gitTooltip,
  onRepoUrlChange,
  onRepoBranchChange,
  workspacePopoverOpen,
  onWorkspacePopoverOpenChange,
  workspaceLabel,
  workspace,
  onWorkspaceChange,
  occupancyForPath,
  branchName,
  baseBranch,
  branchLabel,
  onBranchNameChange,
  onBaseBranchChange,
  selectedProject,
  onProjectChange,
}: NewChatLocationTrayProps) {
  return (
    <div className="relative z-0 -mt-9 flex w-full items-center rounded-b-2xl bg-tray/40 pt-8 pr-3 pb-2 pl-2">
      <div className="flex flex-wrap items-center gap-1">
        <NewChatHostChip
          pendingConnectRef={pendingConnectRef}
          selectedHostId={selectedHostId}
          sandboxSelected={sandboxSelected}
          selectedHost={selectedHost}
          allHosts={allHosts}
          onlineHosts={onlineHosts}
          offlineHosts={offlineHosts}
          thisMachineHostId={thisMachineHostId}
          canConnectThisMachine={canConnectThisMachine}
          showConnectThisMachine={showConnectThisMachine}
          connectingThisMachine={connectingThisMachine}
          hostLabel={hostLabel}
          managedSandboxesEnabled={managedSandboxesEnabled}
          showDisabledSandboxWithDocs={showDisabledSandboxWithDocs}
          sandboxLabel={sandboxLabel}
          newSandboxTooltipContent={newSandboxTooltipContent}
          onSelectHost={onSelectHost}
          onSelectSandbox={onSelectSandbox}
          onConnectThisMachine={onConnectThisMachine}
          onShowConnectInstructions={onShowConnectInstructions}
        />
        {sandboxSelected ? (
          <NewChatRepositoryChip
            label={repoLabel}
            hasRepo={repoHasValue}
            repoUrl={repoUrl}
            repoBranch={repoBranch}
            gitTooltip={gitTooltip}
            onRepoUrlChange={onRepoUrlChange}
            onRepoBranchChange={onRepoBranchChange}
          />
        ) : (
          <>
            <NewChatWorkspaceChip
              open={workspacePopoverOpen}
              onOpenChange={onWorkspacePopoverOpenChange}
              label={workspaceLabel}
              workspace={workspace}
              hostId={selectedHostId}
              onWorkspaceChange={onWorkspaceChange}
              occupancyForPath={occupancyForPath}
            />
            <NewChatBranchChip
              branchName={branchName}
              baseBranch={baseBranch}
              label={branchLabel}
              onBranchNameChange={onBranchNameChange}
              onBaseBranchChange={onBaseBranchChange}
            />
          </>
        )}
        {selectedProject && <LandingProjectPicker value={selectedProject} onChange={onProjectChange} />}
      </div>
    </div>
  );
}

type NewChatLocationTrayProps = {
  pendingConnectRef: MutableRefObject<boolean>;
  selectedHostId: string | null;
  sandboxSelected: boolean;
  selectedHost: Host | undefined;
  allHosts: Host[];
  onlineHosts: Host[];
  offlineHosts: Host[];
  thisMachineHostId: string | null;
  canConnectThisMachine: boolean;
  showConnectThisMachine: boolean;
  connectingThisMachine: boolean;
  hostLabel: string;
  managedSandboxesEnabled: boolean;
  showDisabledSandboxWithDocs: boolean;
  sandboxLabel: string;
  newSandboxTooltipContent?: string;
  onSelectHost: (hostId: string) => void;
  onSelectSandbox: () => void;
  onConnectThisMachine: () => void;
  onShowConnectInstructions: () => void;
  repoLabel: string;
  repoHasValue: boolean;
  repoUrl: string;
  repoBranch: string;
  gitTooltip?: string;
  onRepoUrlChange: (value: string) => void;
  onRepoBranchChange: (value: string) => void;
  workspacePopoverOpen: boolean;
  onWorkspacePopoverOpenChange: (open: boolean) => void;
  workspaceLabel: string;
  workspace: string;
  onWorkspaceChange: (value: string) => void;
  occupancyForPath?: (path: string) => number;
  branchName: string;
  baseBranch: string;
  branchLabel: string;
  onBranchNameChange: (value: string) => void;
  onBaseBranchChange: (value: string) => void;
  selectedProject: string;
  onProjectChange: (value: string) => void;
};

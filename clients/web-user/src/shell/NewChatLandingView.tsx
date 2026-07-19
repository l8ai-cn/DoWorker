import { TriangleAlertIcon } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { DoWorkerLogo } from "@/components/icons/DoWorkerLogo";
import { ConnectHostInstructions } from "./ConnectHostInstructions";
import { HarnessWarningMessage } from "./newChatHarnessWarning";
import { NewChatComposerCard } from "./NewChatComposerCard";
import { NewChatLocationTray } from "./NewChatLocationTray";
import type { NewChatLandingController } from "./useNewChatLandingController";

export function NewChatLandingView({ state }: { state: NewChatLandingController }) {
  const location = state.location;
  return (
    <div
      ref={state.setLandingSurface}
      className="flex flex-1 items-center justify-center bg-background"
      data-testid="new-chat-landing"
    >
      <div className="flex w-full max-w-[840px] flex-col items-center gap-8 px-4 pt-8 pb-16 md:select-none md:px-10">
        <div className="flex flex-col items-center gap-3.5 sm:flex-row">
          <DoWorkerLogo className="h-14 w-14 shrink-0" title="Do Worker" />
          <h1 className="text-center text-3xl font-medium tracking-[-0.03em] text-foreground sm:text-left">
            {state.title}
          </h1>
        </div>
        <div className="relative flex w-full flex-col gap-3">
          <NewChatComposerCard state={state} />
          <NewChatLocationTray
            pendingConnectRef={location.pendingConnectRef}
            selectedHostId={location.selectedHostId}
            sandboxSelected={location.sandboxSelected}
            selectedHost={location.selectedHost}
            allHosts={location.allHosts}
            onlineHosts={location.onlineHosts}
            offlineHosts={location.offlineHosts}
            thisMachineHostId={location.thisMachineHostId}
            canConnectThisMachine={location.canConnectThisMachine}
            showConnectThisMachine={location.showConnectThisMachine}
            connectingThisMachine={location.connectingThisMachine}
            hostLabel={state.labels.host}
            managedSandboxesEnabled={location.managedSandboxesEnabled}
            showDisabledSandboxWithDocs={location.showDisabledSandboxWithDocs}
            sandboxLabel={location.sandboxLabel}
            newSandboxTooltipContent={location.newSandboxTooltipContent}
            onSelectHost={location.selectHost}
            onSelectSandbox={location.selectSandbox}
            onConnectThisMachine={() => void location.connectThisMachine()}
            onShowConnectInstructions={() => location.setConnectOpen(true)}
            repoLabel={state.labels.repo}
            repoHasValue={Boolean(location.sandboxRepoName)}
            repoUrl={location.sandboxRepoUrl}
            repoBranch={location.sandboxRepoBranch}
            gitTooltip={location.databricksGitCredentialsTooltipContent}
            onRepoUrlChange={location.setSandboxRepoUrl}
            onRepoBranchChange={location.setSandboxRepoBranch}
            workspacePopoverOpen={location.workspacePopoverOpen}
            onWorkspacePopoverOpenChange={location.setWorkspacePopoverOpen}
            workspaceLabel={state.labels.workspace}
            workspace={location.workspaceTrimmed}
            onWorkspaceChange={location.setWorkspace}
            occupancyForPath={location.occupancyForPath}
            branchName={location.branchName}
            baseBranch={location.baseBranch}
            branchLabel={state.labels.branch}
            onBranchNameChange={location.setBranchName}
            onBaseBranchChange={location.setBaseBranch}
            selectedProject={location.selectedProject}
            onProjectChange={location.setSelectedProject}
          />
          {state.warning.show && (
            <p
              className="flex items-center gap-1.5 text-xs text-amber-600 dark:text-amber-500"
              data-testid="new-chat-landing-harness-warning"
            >
              <TriangleAlertIcon className="size-3.5 shrink-0" />
              <span>
                <HarnessWarningMessage
                  agentName={state.warning.agentName}
                  hostName={state.warning.hostName}
                  reason={state.warning.reason}
                />
              </span>
            </p>
          )}
          {state.submit.createError && (
            <p className="text-xs text-destructive" data-testid="new-chat-landing-error">
              {state.submit.createError}
            </p>
          )}
        </div>
      </div>
      <Dialog open={location.connectOpen} onOpenChange={location.setConnectOpen}>
        <DialogContent className="sm:max-w-lg" data-testid="connect-host-dialog">
          <DialogHeader>
            <DialogTitle>Connect a host</DialogTitle>
          </DialogHeader>
          <ConnectHostInstructions
            serverUrl={state.serverUrl}
            label="Run this on the machine you want to use, then pick it from the host menu:"
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

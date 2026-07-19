import { CircleHelpIcon, ChevronDownIcon, MonitorCloudIcon, MonitorIcon, PlusIcon } from "lucide-react";
import type { MutableRefObject } from "react";
import type { Host } from "@/hooks/useHosts";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { isHostActiveInPicker } from "@/lib/host-agent-match";
import { NewChatHostOption } from "./NewChatHostOption";

export function NewChatHostChip({
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
}: {
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
}) {
  const isCloudHost = sandboxSelected || (selectedHost?.name.toLowerCase().includes("cloud") ?? false);
  return (
    <DropdownMenu
      onOpenChange={(open) => {
        if (!open && pendingConnectRef.current) {
          pendingConnectRef.current = false;
          onConnectThisMachine();
        }
      }}
    >
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="flex h-6 items-center gap-1 rounded-full px-2.5 text-13 font-normal text-muted-foreground transition-colors hover:text-foreground"
          data-testid="new-chat-landing-host-chip"
        >
          {isCloudHost ? <MonitorCloudIcon className="size-4 shrink-0" /> : <MonitorIcon className="size-4 shrink-0" />}
          <span className={`hidden max-w-32 truncate sm:block ${sandboxSelected || selectedHost != null || connectingThisMachine ? "text-foreground" : ""}`}>
            {hostLabel}
          </span>
          <ChevronDownIcon className="size-3.5 shrink-0 opacity-60" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-52">
        {(managedSandboxesEnabled || showDisabledSandboxWithDocs) && (
          <>
            {managedSandboxesEnabled ? (
              <DropdownMenuItem
                onSelect={onSelectSandbox}
                data-testid="new-chat-landing-sandbox-option"
                data-active={sandboxSelected ? "true" : undefined}
                className="text-xs data-[active=true]:bg-accent/60"
              >
                <span className="flex items-center gap-2">
                  <MonitorCloudIcon className="size-4 text-muted-foreground" />
                  <span className="text-xs">{sandboxLabel}</span>
                </span>
              </DropdownMenuItem>
            ) : (
              <DisabledSandboxItem tooltip={newSandboxTooltipContent} />
            )}
            <DropdownMenuSeparator />
          </>
        )}
        {allHosts.length === 0 && !showConnectThisMachine && (
          <div className="px-2 py-1.5 text-xs text-muted-foreground">No hosts connected yet.</div>
        )}
        {onlineHosts.map((host) => (
          <DropdownMenuItem
            key={host.host_id}
            onSelect={() => onSelectHost(host.host_id)}
            data-active={isHostActiveInPicker(host, selectedHostId, allHosts) ? "true" : undefined}
            className="text-xs data-[active=true]:bg-accent/60"
          >
            <NewChatHostOption
              host={host}
              thisMachineHostId={thisMachineHostId}
              subtitle={host.host_id === thisMachineHostId ? "this machine" : undefined}
            />
          </DropdownMenuItem>
        ))}
        {offlineHosts.map((host) =>
          host.host_id === thisMachineHostId && canConnectThisMachine ? (
            <ConnectMachineItem
              key={host.host_id}
              host={host}
              thisMachineHostId={thisMachineHostId}
              connecting={connectingThisMachine}
              onQueueConnect={() => {
                pendingConnectRef.current = true;
              }}
            />
          ) : (
            <DropdownMenuItem key={host.host_id} disabled className="text-xs">
              <NewChatHostOption
                host={host}
                thisMachineHostId={thisMachineHostId}
                subtitle={host.host_id === thisMachineHostId ? "this machine" : undefined}
              />
            </DropdownMenuItem>
          ),
        )}
        {showConnectThisMachine && (
          <DropdownMenuItem
            onSelect={() => {
              pendingConnectRef.current = true;
            }}
            disabled={connectingThisMachine}
            data-testid="new-chat-landing-run-on-this-machine"
            className="gap-2 text-xs"
          >
            <MonitorIcon className="size-4 shrink-0 text-muted-foreground" />
            <span className="text-xs">{connectingThisMachine ? "Connecting this machine..." : "Run on this machine"}</span>
          </DropdownMenuItem>
        )}
        {(allHosts.length > 0 || showConnectThisMachine) && <DropdownMenuSeparator />}
        <DropdownMenuItem
          onSelect={onShowConnectInstructions}
          data-testid="new-chat-landing-connect-host"
          className="gap-2 text-xs text-muted-foreground"
        >
          <PlusIcon className="size-3.5" />
          Connect new host
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function DisabledSandboxItem({ tooltip }: { tooltip?: string }) {
  return (
    <DropdownMenuItem
      aria-disabled="true"
      onSelect={(event) => event.preventDefault()}
      className="flex items-center justify-between px-2 py-1.5 text-xs text-muted-foreground opacity-60"
      data-testid="new-chat-landing-sandbox-option-disabled"
    >
      <span className="flex items-center gap-2">
        <MonitorCloudIcon className="size-4 text-muted-foreground" />
        <span className="text-xs">New Sandbox</span>
      </span>
      <Tooltip>
        <TooltipTrigger asChild>
          <button type="button" className="inline-flex size-4 items-center justify-center rounded-sm text-muted-foreground/80 hover:text-foreground" aria-label="Why New Sandbox is unavailable">
            <CircleHelpIcon className="size-3.5" />
          </button>
        </TooltipTrigger>
        <TooltipContent className="max-w-64">{tooltip}</TooltipContent>
      </Tooltip>
    </DropdownMenuItem>
  );
}

function ConnectMachineItem({ host, thisMachineHostId, connecting, onQueueConnect }: { host: Host; thisMachineHostId: string | null; connecting: boolean; onQueueConnect: () => void }) {
  return (
    <DropdownMenuItem onSelect={onQueueConnect} disabled={connecting} data-testid="new-chat-landing-run-on-this-machine" className="text-xs">
      <NewChatHostOption
        host={host}
        thisMachineHostId={thisMachineHostId}
        subtitle={connecting ? "connecting..." : "this machine · select to connect"}
      />
    </DropdownMenuItem>
  );
}

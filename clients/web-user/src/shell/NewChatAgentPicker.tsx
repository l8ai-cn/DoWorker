import { ChevronDownIcon, PlusIcon } from "lucide-react";
import type { ReactNode } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { AgentRowTooltip } from "@/components/AgentHoverCard";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import type { Host } from "@/hooks/useHosts";
import { AGENT_PICKER_DESCRIPTIONS } from "./newChatConstants";
import {
  harnessUnconfiguredOnHost,
  harnessUnavailableReasonOnHost,
  harnessWarningBadgeText,
} from "./newChatHarnessWarning";

export function NewChatAgentPicker({
  agentEntries,
  harnessEntries,
  effectiveAgentId,
  agentLabel,
  hasAgents,
  host,
  onSelectAgent,
}: {
  agentEntries: AvailableAgent[];
  harnessEntries: AvailableAgent[];
  effectiveAgentId: string | null;
  agentLabel: string;
  hasAgents: boolean;
  host: Host | undefined | null;
  onSelectAgent: (agent: AvailableAgent) => void;
}) {
  const renderEntry = (agent: AvailableAgent) => (
    <DropdownMenuItem
      key={agent.id}
      data-testid={`new-chat-landing-agent-${agent.id}`}
      data-active={agent.id === effectiveAgentId ? "true" : undefined}
      onSelect={() => onSelectAgent(agent)}
      className="items-start gap-2 rounded-sm px-2 py-1.5 text-13 data-[active=true]:bg-accent/60 data-[active=true]:text-foreground"
    >
      <AgentRowTooltip agent={agent}>
        <div className="flex min-w-0 flex-1 items-baseline gap-2.5">
          <span className="truncate">{agent.display_name}</span>
          {AGENT_PICKER_DESCRIPTIONS[agent.name] && (
            <span className="truncate text-[11px] text-muted-foreground/70">
              {AGENT_PICKER_DESCRIPTIONS[agent.name]}
            </span>
          )}
        </div>
      </AgentRowTooltip>
      <HarnessBadge agent={agent} host={host} />
    </DropdownMenuItem>
  );

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          disabled={!hasAgents}
          data-testid="new-chat-landing-agent-select"
          className="h-8 gap-1.5 px-2.5 font-normal text-muted-foreground hover:text-foreground focus-visible:border-transparent focus-visible:ring-0"
        >
          <span className="max-w-[12rem] truncate text-xs text-foreground">
            {hasAgents ? agentLabel : "No agents"}
          </span>
          <ChevronDownIcon className="size-3.5 opacity-60" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        className="max-h-[var(--radix-dropdown-menu-content-available-height)] min-w-64 max-w-[calc(100vw-2rem)] overflow-y-auto p-1"
      >
        {harnessEntries.length > 0 && (
          <>
            <PickerSectionHeader>Harnesses</PickerSectionHeader>
            {harnessEntries.map(renderEntry)}
            <DropdownMenuSeparator />
          </>
        )}
        <PickerSectionHeader>Agents</PickerSectionHeader>
        {agentEntries.map(renderEntry)}
        <DropdownMenuItem
          data-testid="new-chat-landing-create-agent"
          aria-disabled="true"
          onSelect={(event) => event.preventDefault()}
          className="gap-2 rounded-sm px-2 py-1.5 text-13 text-muted-foreground opacity-60"
        >
          <PlusIcon className="size-3.5" />
          Register a custom WorkerTemplate first
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function PickerSectionHeader({ children }: { children: ReactNode }) {
  return (
    <div className="px-2 pt-1.5 pb-0.5 text-[11px] font-medium text-muted-foreground">
      {children}
    </div>
  );
}

function HarnessBadge({ agent, host }: { agent: AvailableAgent; host: Host | undefined | null }) {
  if (!harnessUnconfiguredOnHost(agent.harness, host)) return null;
  return (
    <Badge
      variant="outline"
      className="ml-auto self-center border-amber-300 bg-amber-50 text-[11px] text-amber-700 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-400"
      data-testid={`new-chat-landing-agent-warning-${agent.id}`}
    >
      {harnessWarningBadgeText(harnessUnavailableReasonOnHost(agent.harness, host))}
    </Badge>
  );
}

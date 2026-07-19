import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import { hasWorkerCreationSelection } from "@/lib/workerCreationSelection";

export function SwitchAgentSelector({
  choice,
  currentDisplay,
  agents,
  onChoiceChange,
}: {
  choice: string;
  currentDisplay: string | null;
  agents: AvailableAgent[];
  onChoiceChange: (agentId: string) => void;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor="switch-agent-select" className="text-xs font-medium text-muted-foreground">
        Agent
      </label>
      <Select value={choice || undefined} onValueChange={onChoiceChange}>
        <SelectTrigger
          id="switch-agent-select"
          data-testid="switch-agent-select"
          className="w-full text-xs"
        >
          <SelectValue
            placeholder={
              currentDisplay ? (
                <span data-testid="switch-agent-current">
                  <span className="text-foreground">{currentDisplay}</span>{" "}
                  <span className="text-muted-foreground">(current agent)</span>
                </span>
              ) : (
                "Choose an agent"
              )
            }
          />
        </SelectTrigger>
        <SelectContent position="popper" align="start">
          {agents.map((agent) => {
            const selectable = hasWorkerCreationSelection(agent);
            return (
              <SelectItem
                key={agent.id}
                value={agent.id}
                disabled={!selectable}
                data-testid={`switch-agent-option-${agent.id}`}
                className="text-xs"
              >
                {agent.display_name}
                {!selectable ? " (Worker creation unavailable)" : ""}
              </SelectItem>
            );
          })}
        </SelectContent>
      </Select>
    </div>
  );
}

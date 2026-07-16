import { Bot } from "lucide-react";
import type { AgentData } from "@/lib/api";

export function AgentConfigHeader({ agent }: { agent: AgentData }) {
  return (
    <div className="flex items-center gap-3">
      <Bot className="h-8 w-8 text-primary" />
      <div>
        <h2 className="text-xl font-semibold">{agent.name}</h2>
        {agent.description && (
          <p className="text-sm text-muted-foreground">{agent.description}</p>
        )}
      </div>
    </div>
  );
}

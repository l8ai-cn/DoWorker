import { Bot, ShieldCheck, Sparkles, SquareTerminal } from "lucide-react";

import type { AgentSessionSnapshot } from "./contracts";
import type { AgentSessionRuntime } from "./contracts";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import { ComposerConfigurationBar } from "./ComposerConfigurationBar";

export function ComposerCapabilityBar({
  onError,
  runtime,
  snapshot,
}: {
  onError: (error: unknown) => void;
  runtime: AgentSessionRuntime;
  snapshot: AgentSessionSnapshot;
}) {
  const text = useAgentWorkspaceText();
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-0.5 text-muted-foreground">
      <CapabilityLabel
        icon={<Sparkles className="size-3.5" />}
        label={snapshot.interactionMode === "acp" ? text.agentic : text.terminalMode}
      />
      <CapabilityLabel
        icon={<Bot className="size-3.5" />}
        label={snapshot.agentLabel}
      />
      <ComposerConfigurationBar
        onError={onError}
        runtime={runtime}
        snapshot={snapshot}
      />
      {snapshot.capabilities.resolvePermission && (
        <CapabilityLabel
          icon={<ShieldCheck className="size-3.5" />}
          label={text.approvals}
        />
      )}
      {snapshot.capabilities.terminal && snapshot.terminals.length > 0 && (
        <CapabilityLabel
          icon={<SquareTerminal className="size-3.5" />}
          label={text.terminal}
        />
      )}
    </div>
  );
}

function CapabilityLabel({
  icon,
  label,
}: {
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <span
      className="flex h-8 min-w-0 items-center gap-1.5 px-2 text-xs"
      title={label}
    >
      <span className="shrink-0">{icon}</span>
      <span className="max-w-28 truncate">{label}</span>
    </span>
  );
}

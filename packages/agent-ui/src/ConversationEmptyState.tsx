import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";

export function ConversationEmptyState({
  agentLabel,
}: {
  agentLabel: string;
}) {
  const text = useAgentWorkspaceText();
  return (
    <div className="shrink-0 px-6 text-center">
      <h2 className="max-w-3xl text-2xl font-medium leading-tight">
        {text.emptyHeading(agentLabel)}
      </h2>
    </div>
  );
}

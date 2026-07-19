import { useEffect, useMemo, useState } from "react";
import { useAvailableAgents, type AvailableAgent } from "@/hooks/useAvailableAgents";
import { readLastAgentId, writeLastAgentId } from "@/lib/agentPreferences";
import { sortAgentsForDisplay } from "@/lib/agentGrouping";
import { isNativeCodingAgent } from "@/lib/nativeCodingAgents";
import { NEW_SESSION_HIDDEN_AGENTS } from "./newChatConstants";
import type { NewChatLandingDraft } from "./newChatLandingDraft";

export function useNewChatAgentSelection({
  agentParam,
  landingDraft,
}: {
  agentParam: string;
  landingDraft: NewChatLandingDraft | null;
}) {
  const { data: agents } = useAvailableAgents();
  const agentList = useMemo(
    () =>
      sortAgentsForDisplay((agents ?? []).filter((agent) => !NEW_SESSION_HIDDEN_AGENTS.has(agent.name))),
    [agents],
  );
  const [pickedAgentId, setPickedAgentId] = useState<string | null>(
    () => landingDraft?.pickedAgentId ?? readLastAgentId(),
  );

  useEffect(() => {
    if (agentParam && agentList.some((agent) => agent.id === agentParam)) {
      setPickedAgentId(agentParam);
    }
  }, [agentParam, agentList]);

  const effectiveAgentId = agentList.some((agent) => agent.id === pickedAgentId)
    ? pickedAgentId
    : (agentList.find((agent) => agent.harness === "do-agent")?.id ?? agentList[0]?.id ?? null);
  const selectedAgent = agentList.find((agent) => agent.id === effectiveAgentId);
  const harnessEntries = useMemo(
    () => agentList.filter((agent) => isNativeCodingAgent(agent)),
    [agentList],
  );
  const agentEntries = useMemo(
    () => agentList.filter((agent) => !isNativeCodingAgent(agent)),
    [agentList],
  );

  const selectAgent = (agent: AvailableAgent) => {
    setPickedAgentId(agent.id);
    writeLastAgentId(agent.id);
  };

  return {
    agentList,
    agentEntries,
    harnessEntries,
    pickedAgentId,
    effectiveAgentId,
    selectedAgent,
    selectAgent,
  };
}

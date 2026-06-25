import { MsgType } from "@/stores/relayProtocol";
import { useDoAgentConsoleStore, type DoAgentGoal } from "@/stores/doagentConsole";

function parseGoals(data: Record<string, unknown>): DoAgentGoal[] {
  const goals = data.goals ?? data.Goals;
  if (!Array.isArray(goals)) return [];
  return goals
    .map((raw) => {
      const g = raw as Record<string, unknown>;
      const id = String(g.id ?? g.goalId ?? "");
      if (!id) return null;
      return {
        id,
        title: typeof g.title === "string" ? g.title : undefined,
        status: typeof g.status === "string" ? g.status : undefined,
      } satisfies DoAgentGoal;
    })
    .filter((g): g is DoAgentGoal => g !== null);
}

export function dispatchDoAgentRelayEvent(
  podKey: string,
  msgType: number,
  payload: unknown,
): boolean {
  if (msgType !== MsgType.AcpEvent) return false;
  const data = payload as Record<string, unknown>;
  if (data?.type !== "controlResponse") return false;

  const store = useDoAgentConsoleStore.getState();
  store.touchResponse(podKey);
  const parsed = parseGoals(data);
  if (parsed.length > 0) {
    store.setGoals(podKey, parsed);
  }
  return false;
}

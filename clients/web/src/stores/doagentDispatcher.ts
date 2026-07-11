import { MsgType } from "@/stores/relayProtocol";
import { useDoAgentConsoleStore, type DoAgentGoal } from "@/stores/doagentConsole";

function parseGoals(data: Record<string, unknown>): DoAgentGoal[] {
  const goals = data.goals ?? data.Goals;
  if (!Array.isArray(goals)) return [];
  const parsed: DoAgentGoal[] = [];
  for (const raw of goals) {
    const g = raw as Record<string, unknown>;
    const id = String(g.id ?? g.goalId ?? "");
    if (!id) continue;
    parsed.push({
      id,
      title: typeof g.title === "string" ? g.title : undefined,
      status: typeof g.status === "string" ? g.status : undefined,
    });
  }
  return parsed;
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

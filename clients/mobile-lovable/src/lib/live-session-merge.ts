import type { LiveAgentEvent } from "./live-session-reducer";

function isPersistedId(id: string): boolean {
  return id.startsWith("item_");
}

function ephemeralRedundant(ep: LiveAgentEvent, persisted: LiveAgentEvent[]): boolean {
  if (ep.type === "user_message" && ep.markdown) {
    return persisted.some((p) => p.type === "user_message" && p.markdown === ep.markdown);
  }
  if (ep.type === "agent_message" && ep.status === "in_progress") {
    return persisted.some(
      (p) =>
        p.type === "agent_message" &&
        p.status === "completed" &&
        p.markdown != null &&
        ep.markdown != null &&
        (p.markdown === ep.markdown || p.markdown.startsWith(ep.markdown)),
    );
  }
  return false;
}

export function reconcileLiveEvents(
  persisted: LiveAgentEvent[],
  ephemeral: LiveAgentEvent[],
): LiveAgentEvent[] {
  const merged = persisted.slice();
  for (const ep of ephemeral) {
    if (isPersistedId(ep.id) || ephemeralRedundant(ep, merged)) continue;
    merged.push(ep);
  }
  return merged;
}

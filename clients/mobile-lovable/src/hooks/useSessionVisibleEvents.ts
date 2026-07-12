import { useEffect, useState } from "react";
import type { AgentEvent, AgentSession } from "@/lib/session-types";

export function useSessionVisibleEvents(session: AgentSession, isLive: boolean) {
  const sessionActive =
    session.status === "running" || session.status === "waiting_approval";
  const isStreaming = isLive && sessionActive;
  const isMockAnimating = !isLive && sessionActive;
  const [visibleCount, setVisibleCount] = useState(
    isLive ? session.events.length : isMockAnimating ? Math.max(1, session.events.length - 1) : session.events.length,
  );

  useEffect(() => {
    if (isLive) {
      setVisibleCount(session.events.length);
      return;
    }
    if (!isMockAnimating) return;
    if (visibleCount >= session.events.length) return;
    const next = session.events[visibleCount];
    const delay =
      next?.type === "phase" ? 700 :
      next?.type === "agent_thought" ? 450 :
      next?.type === "agent_message" ? 900 :
      next?.type === "tool_call" && next.toolKind === "shell" ? 1100 :
      next?.type === "tool_call" ? 650 :
      next?.type === "ask_user" ? 900 :
      800;
    const t = setTimeout(() => setVisibleCount((c) => c + 1), delay);
    return () => clearTimeout(t);
  }, [isLive, isMockAnimating, visibleCount, session.events]);

  const visibleEvents = session.events.slice(0, visibleCount);
  const currentPhase = [...visibleEvents].reverse().find((e) => e.type === "phase");
  const phaseTotal = session.events.find((e) => e.type === "phase")?.phaseTotal;

  return { visibleEvents, visibleCount, isStreaming, currentPhase, phaseTotal };
}

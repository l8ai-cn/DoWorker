import type { AgentEvent } from "@/lib/session-types";
import { AskUserCard } from "@/components/session/session-ask-user-card";
import {
  AgentBubble,
  ErrorCard,
  PhaseCard,
  PlanCard,
  ThoughtBlock,
  UserBubble,
} from "@/components/session/session-message-cards";
import { ToolCard } from "@/components/session/session-tool-card";

export function EventCard({ event }: { event: AgentEvent }) {
  switch (event.type) {
    case "user_message":   return <UserBubble event={event} />;
    case "agent_message":  return <AgentBubble event={event} />;
    case "agent_thought":  return <ThoughtBlock event={event} />;
    case "plan":           return <PlanCard event={event} />;
    case "ask_user":       return <AskUserCard event={event} />;
    case "phase":          return <PhaseCard event={event} />;
    case "permission_request":
    case "tool_call":      return <ToolCard event={event} />;
    case "error":          return <ErrorCard event={event} />;
    default:               return null;
  }
}

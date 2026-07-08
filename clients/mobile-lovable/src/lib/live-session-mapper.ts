import type { AgentSession } from "./mock-agents";
import { displayAgentName } from "./agent-slugs";
import { projectIdFromName } from "./project-label";
import type { LiveSessionSummary, SessionStatus } from "./sessions-api";

function mapStatus(s: LiveSessionSummary): AgentSession["status"] {
  if (s.pendingApprovals > 0) return "waiting_approval";
  const api = s.status as SessionStatus;
  if (api === "running" || api === "launching" || api === "waiting") return "running";
  if (api === "failed") return "failed";
  return "idle";
}

export function liveSummaryToAgentSession(s: LiveSessionSummary): AgentSession {
  return {
    id: s.id,
    projectId: s.project ? projectIdFromName(s.project) : "live",
    title: s.title ?? displayAgentName(s.agentId),
    agent: s.agentName ?? displayAgentName(s.agentId),
    status: mapStatus(s),
    updatedAt: s.updatedAt,
    preview: s.workspace ?? "",
    events: [],
  };
}

export function statusRank(s: AgentSession["status"]): number {
  switch (s) {
    case "waiting_approval":
      return 0;
    case "running":
      return 1;
    case "idle":
      return 2;
    case "failed":
      return 3;
    case "completed":
      return 4;
    default:
      return 5;
  }
}

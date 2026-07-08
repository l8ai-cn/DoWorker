import type { LiveSessionSummary, SessionStatus } from "./sessions-api";
import { PROJECT_LABEL_KEY } from "./project-label";

type WireRow = Record<string, unknown> & { id: string };

function formatRelative(ts?: number): string {
  if (!ts) return "刚刚";
  const diff = Date.now() - ts * 1000;
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "刚刚";
  if (mins < 60) return `${mins} 分钟前`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs} 小时前`;
  return `${Math.floor(hrs / 24)} 天前`;
}

export function wireToSummary(wire: WireRow, base?: LiveSessionSummary): LiveSessionSummary {
  const pending =
    typeof wire.pending_elicitations_count === "number"
      ? wire.pending_elicitations_count
      : (base?.pendingApprovals ?? 0);
  const labels = wire.labels as Record<string, string> | undefined;
  const project = labels?.[PROJECT_LABEL_KEY]?.trim() || base?.project || null;
  return {
    id: wire.id,
    title: (wire.title as string | null | undefined) ?? base?.title ?? null,
    agentId: (wire.agent_id as string | undefined) ?? base?.agentId ?? "agent",
    agentName: (wire.agent_name as string | null | undefined) ?? base?.agentName ?? null,
    status: ((wire.status as SessionStatus | undefined) ?? base?.status ?? "idle") as SessionStatus,
    pendingApprovals: pending,
    updatedAt:
      typeof wire.updated_at === "number"
        ? formatRelative(wire.updated_at)
        : (base?.updatedAt ?? "刚刚"),
    workspace: (wire.workspace as string | null | undefined) ?? base?.workspace ?? null,
    project,
  };
}

export function mergeSessionRows(
  items: LiveSessionSummary[],
  rows: WireRow[],
): LiveSessionSummary[] {
  if (rows.length === 0) return items;
  const map = new Map(items.map((s) => [s.id, s]));
  for (const row of rows) {
    map.set(row.id, wireToSummary(row, map.get(row.id)));
  }
  const merged = [...map.values()];
  merged.sort((a, b) => {
    const ai = items.findIndex((s) => s.id === a.id);
    const bi = items.findIndex((s) => s.id === b.id);
    if (ai >= 0 && bi >= 0) return ai - bi;
    if (ai >= 0) return -1;
    if (bi >= 0) return 1;
    return 0;
  });
  return merged;
}

export function removeSessionRows(items: LiveSessionSummary[], ids: string[]): LiveSessionSummary[] {
  if (ids.length === 0) return items;
  const drop = new Set(ids);
  return items.filter((s) => !drop.has(s.id));
}

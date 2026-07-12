import { useCallback, useEffect, useRef, useState } from "react";
import type { AgentEvent, AgentSession, SessionStatus } from "@/lib/session-types";
import { LiveSessionReducer, parseSseStream } from "@/lib/live-session-reducer";
import {
  fetchSessionItems,
  getSession,
  openSessionStream,
  resolveElicitation,
  stopSession,
  type LiveSessionSummary,
  type SessionStatus as ApiStatus,
} from "@/lib/sessions-api";
import { postSessionMessageWithFiles } from "@/lib/session-message-upload";
import { itemsToLiveEvents } from "@/lib/session-items-hydrator";
import { displayAgentName } from "@/lib/agent-slugs";
import { readAuthToken } from "@/lib/auth-store";

function mapStatus(api: ApiStatus, pending: number): SessionStatus {
  if (pending > 0) return "waiting_approval";
  if (api === "running" || api === "launching" || api === "waiting") return "running";
  if (api === "failed") return "failed";
  return "idle";
}

function toAgentSession(summary: LiveSessionSummary, events: AgentEvent[]): AgentSession {
  return {
    id: summary.id,
    interactionMode: summary.interactionMode,
    projectId: "live",
    title: summary.title ?? summary.agentName ?? "新任务",
    agent: summary.agentName ?? displayAgentName(summary.agentId),
    branch: summary.workspace ?? "",
    status: mapStatus(summary.status, summary.pendingApprovals),
    updatedAt: summary.updatedAt,
    eventCount: events.length,
    preview: events.at(-1)?.detail ?? events.at(-1)?.title ?? "",
    events,
  };
}

function mapLiveEvents(snap: ReturnType<LiveSessionReducer["apply"]>): AgentEvent[] {
  return snap.map((e) => ({
    id: e.id,
    type: e.type as AgentEvent["type"],
    ts: e.ts,
    title: e.title,
    detail: e.detail,
    markdown: e.markdown,
    tool: e.tool,
    toolKind: e.toolKind,
    status: e.status,
    elicitationId: e.elicitationId,
  }));
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export function useLiveSession(sessionId: string) {
  const [summary, setSummary] = useState<LiveSessionSummary | null>(null);
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const reducerRef = useRef(new LiveSessionReducer());
  const sessionIdRef = useRef(sessionId);

  const syncEvents = useCallback(() => {
    setEvents(mapLiveEvents(reducerRef.current.snapshot));
  }, []);

  const resyncItems = useCallback(async () => {
    const items = await fetchSessionItems(sessionId);
    reducerRef.current.reconcile(itemsToLiveEvents(items));
    syncEvents();
  }, [sessionId, syncEvents]);

  const refresh = useCallback(async () => {
    const s = await getSession(sessionId);
    setSummary(s);
  }, [sessionId]);

  useEffect(() => {
    if (!readAuthToken()) {
      setLoading(false);
      setError("未登录");
      return;
    }

    const sessionChanged = sessionIdRef.current !== sessionId;
    sessionIdRef.current = sessionId;
    if (sessionChanged) {
      setLoading(true);
      setError(null);
      setEvents([]);
      setSummary(null);
      reducerRef.current = new LiveSessionReducer();
    }

    let cancelled = false;
    const ac = new AbortController();

    const hydrate = async () => {
      const [s, items] = await Promise.all([getSession(sessionId), fetchSessionItems(sessionId)]);
      if (cancelled) return;
      setSummary(s);
      const persisted = itemsToLiveEvents(items);
      if (reducerRef.current.snapshot.length === 0) {
        reducerRef.current.seed(persisted);
      } else {
        reducerRef.current.reconcile(persisted);
      }
      syncEvents();
      setLoading(false);
    };

    const streamLoop = async () => {
      while (!cancelled && !ac.signal.aborted) {
        try {
          const res = await openSessionStream(sessionId, ac.signal);
          if (!res.ok || !res.body) throw new Error(`SSE ${res.status}`);
          for await (const frame of parseSseStream(res.body)) {
            if (cancelled) break;
            reducerRef.current.apply(frame.event, frame.data);
            syncEvents();
            if (frame.event === "session.status") {
              void refresh();
              void resyncItems();
            }
          }
        } catch (e) {
          if (cancelled || ac.signal.aborted) break;
          setError(e instanceof Error ? e.message : "实时连接失败");
        }
        if (cancelled || ac.signal.aborted) break;
        await resyncItems().catch(() => undefined);
        await sleep(1500);
      }
    };

    void hydrate()
      .then(() => streamLoop())
      .catch((e) => {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "加载失败");
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
      ac.abort();
    };
  }, [sessionId, refresh, resyncItems, syncEvents]);

  const session: AgentSession | null = summary ? toAgentSession(summary, events) : null;

  const send = useCallback(
    async (text: string, files: File[]) => {
      await postSessionMessageWithFiles(sessionId, text, files);
    },
    [sessionId],
  );

  const stop = useCallback(async () => {
    await stopSession(sessionId);
    await refresh();
    await resyncItems();
  }, [sessionId, refresh, resyncItems]);

  const approve = useCallback(
    async (elicitationId: string, accept: boolean) => {
      await resolveElicitation(sessionId, elicitationId, accept);
    },
    [sessionId],
  );

  return { session, loading, error, send, stop, approve, refresh, isLive: true };
}

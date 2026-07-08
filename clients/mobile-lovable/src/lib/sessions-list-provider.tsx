import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { useIsAuthed } from "@/hooks/useIsAuthed";
import { readAuthToken } from "./auth-store";
import { syncNotificationsFromSessions } from "./notifications";
import { listSessions, type LiveSessionSummary } from "./sessions-api";
import { mergeSessionRows, removeSessionRows } from "./sessions-list-merge";
import { sessionUpdatesSocket } from "./session-updates-socket";

interface SessionsListContextValue {
  items: LiveSessionSummary[];
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  isLive: boolean;
  pendingCount: number;
}

const SessionsListContext = createContext<SessionsListContextValue | null>(null);

export function SessionsListProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<LiveSessionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const authed = useIsAuthed();

  const refresh = useCallback(async () => {
    if (!readAuthToken()) {
      setItems([]);
      setLoading(false);
      return;
    }
    try {
      setError(null);
      setItems(await listSessions());
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!authed) {
      setItems([]);
      setLoading(false);
      return;
    }
    void refresh();
    const poll = setInterval(() => void refresh(), 60_000);
    return () => clearInterval(poll);
  }, [refresh, authed]);

  useEffect(() => {
    if (!authed) {
      sessionUpdatesSocket.stop();
      return;
    }
    sessionUpdatesSocket.start();
    const unsub = sessionUpdatesSocket.subscribe((frame) => {
      if (frame.type === "snapshot" || frame.type === "changed") {
        setItems((prev) => mergeSessionRows(prev, frame.items));
      } else if (frame.type === "removed") {
        setItems((prev) => removeSessionRows(prev, frame.ids));
      }
    });
    return () => {
      unsub();
      sessionUpdatesSocket.stop();
    };
  }, [authed]);

  useEffect(() => {
    if (!authed) return;
    sessionUpdatesSocket.setWatched(items.map((s) => s.id));
  }, [authed, items]);

  useEffect(() => {
    if (authed && items.length >= 0) {
      syncNotificationsFromSessions(items);
    }
  }, [authed, items]);

  const pendingCount = useMemo(
    () => items.filter((s) => s.pendingApprovals > 0).length,
    [items],
  );

  const value = useMemo(
    () => ({ items, loading, error, refresh, isLive: authed, pendingCount }),
    [items, loading, error, refresh, authed, pendingCount],
  );

  return <SessionsListContext.Provider value={value}>{children}</SessionsListContext.Provider>;
}

export function useSessionsListContext(): SessionsListContextValue {
  const ctx = useContext(SessionsListContext);
  if (!ctx) {
    throw new Error("useSessionsListContext must be used within SessionsListProvider");
  }
  return ctx;
}

"use client";

import { useCallback, useEffect, useState } from "react";
import { FileInput, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { listImportedSessions, type ImportedSessionSummary } from "@/lib/api/sessionImportApi";
import {
  ACP_SNAPSHOT_MSG_TYPE,
  codexItemsToAcpSnapshot,
} from "@/lib/codexItemsToAcpSnapshot";
import { fetchAllSessionItems } from "@/lib/api/sessionImportApi";
import { dispatchAcpRelayEvent } from "@/stores/acpEventDispatcher";
import { usePodStore } from "@/stores/pod";
import { useWorkspaceStore } from "@/stores/workspace";
import { readCurrentOrg } from "@/stores/auth";
import { getPod } from "@/lib/api/facade/podConnect";

const refreshListeners = new Set<() => void>();

/** Notify all mounted ImportedSessionsSection instances to reload. */
export function refreshImportedSessionsList(): void {
  for (const listener of refreshListeners) listener();
}

interface ImportedSessionsSectionProps {
  t: (key: string, params?: Record<string, string | number>) => string;
}

export function ImportedSessionsSection({ t }: ImportedSessionsSectionProps) {
  const [sessions, setSessions] = useState<ImportedSessionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const addPane = useWorkspaceStore((s) => s.addPane);
  const panes = useWorkspaceStore((s) => s.panes);

  const load = useCallback(async () => {
    if (!readCurrentOrg()?.slug) {
      setSessions([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const rows = await listImportedSessions();
      setSessions(rows.filter((row) => row.podKey));
    } catch (e) {
      setError(e instanceof Error ? e.message : t("workspace.importedSessions.error"));
      setSessions([]);
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void load();
    refreshListeners.add(load);
    return () => {
      refreshListeners.delete(load);
    };
  }, [load]);

  const handleOpen = useCallback(async (session: ImportedSessionSummary) => {
    const podKey = session.podKey;
    if (!podKey) return;

    const orgSlug = readCurrentOrg()?.slug;
    if (orgSlug) {
      try {
        const pod = await getPod(orgSlug, podKey);
        usePodStore.getState().upsertPod(pod);
      } catch {
        // Opening by pod_key still works when pod metadata fetch fails.
      }
    }

    try {
      const items = await fetchAllSessionItems(session.id);
      if (items.length > 0) {
        const snapshot = codexItemsToAcpSnapshot(session.id, items);
        dispatchAcpRelayEvent(podKey, ACP_SNAPSHOT_MSG_TYPE, snapshot);
      }
    } catch {
      // History hydration is best-effort; the pane can still open.
    }

    if (!panes.some((p) => p.podKey === podKey)) {
      addPane(podKey);
    }
  }, [addPane, panes]);

  if (loading) {
    return (
      <div className="border-t px-3 py-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Loader2 className="h-3.5 w-3.5 animate-spin" />
          {t("workspace.importedSessions.loading")}
        </div>
      </div>
    );
  }

  if (error || sessions.length === 0) {
    return null;
  }

  return (
    <div className="border-t px-2 py-2">
      <div className="mb-2 flex items-center gap-1.5 px-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
        <FileInput className="h-3.5 w-3.5" />
        {t("workspace.importedSessions.title")}
      </div>
      <div className="space-y-1">
        {sessions.map((session) => {
          const open = session.podKey ? panes.some((p) => p.podKey === session.podKey) : false;
          const label = session.title?.trim() || session.agentId || session.id;
          return (
            <button
              key={session.id}
              type="button"
              data-testid={`imported-session-${session.id}`}
              onClick={() => void handleOpen(session)}
              className={cn(
                "w-full rounded-md px-2 py-2 text-left transition-colors",
                open ? "bg-muted text-foreground" : "hover:bg-surface-muted text-foreground/90",
              )}
            >
              <div className="truncate text-sm font-medium">{label}</div>
              <div className="truncate text-[11px] text-muted-foreground">
                {session.agentId} · {session.podKey}
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

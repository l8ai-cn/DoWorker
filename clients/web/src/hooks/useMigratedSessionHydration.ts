import { useEffect, useRef } from "react";
import { fetchAllSessionItems, fetchSessionByPodKey } from "@/lib/api/sessionImportApi";
import {
  ACP_SNAPSHOT_MSG_TYPE,
  codexItemsToAcpSnapshot,
} from "@/lib/codexItemsToAcpSnapshot";
import { dispatchAcpRelayEvent } from "@/stores/acpEventDispatcher";
import { useAcpSessionField } from "@/stores/acpSession";

/** Hydrate ACP activity stream from persisted conversation_items when opening an imported Worker. */
export function useMigratedSessionHydration(podKey: string, enabled: boolean): void {
  const messageCount = useAcpSessionField(podKey, (s) => s.messages.length);
  const hydratedRef = useRef<string | null>(null);

  useEffect(() => {
    if (!enabled || !podKey || messageCount > 0) return;
    if (hydratedRef.current === podKey) return;

    let cancelled = false;
    void (async () => {
      const session = await fetchSessionByPodKey(podKey);
      if (!session || cancelled) return;
      const items = await fetchAllSessionItems(session.id);
      if (cancelled || items.length === 0) return;

      const snapshot = codexItemsToAcpSnapshot(session.id, items);
      if (snapshot.messages.length === 0 && snapshot.toolCalls.length === 0) return;

      dispatchAcpRelayEvent(podKey, ACP_SNAPSHOT_MSG_TYPE, snapshot);
      hydratedRef.current = podKey;
    })();

    return () => {
      cancelled = true;
    };
  }, [enabled, messageCount, podKey]);
}

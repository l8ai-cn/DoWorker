import { useEffect, useSyncExternalStore } from "react";

import type { AgentSessionRuntime } from "./contracts";

export function useAgentSessionSnapshot(
  runtime: AgentSessionRuntime,
  sessionId: string,
) {
  useEffect(() => {
    void runtime.open(sessionId);
    return () => runtime.close(sessionId);
  }, [runtime, sessionId]);

  return useSyncExternalStore(
    (listener) => runtime.subscribe(sessionId, listener),
    () => runtime.getSnapshot(sessionId),
    () => runtime.getSnapshot(sessionId),
  );
}
